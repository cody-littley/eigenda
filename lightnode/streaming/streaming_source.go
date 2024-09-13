package streaming

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/eigenda/api/grpc/lightnode"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var _ lightnode.SourceServer = &Source{}

type Source struct {
	lightnode.SourceServer

	ctx    context.Context
	cancel context.CancelFunc
	config *Config

	captureMetrics atomic.Bool
	bytesSent      atomic.Uint64

	// Reusable random data to send.
	randomData [][]byte

	// The ID of the next connection to be created.
	nextConnectionID int

	// A map from connection ID to the time the connection was started.
	// This is used to determine the number of messages that should have been sent on a connection.
	connectionStartTime map[int]time.Time

	// A map from connection ID to the index of the next message to send.
	nextMessageIndex map[int]uint64

	// Protects access to the metadata for each connection.
	connectionMetadataLock sync.Mutex
}

func NewSource(
	ctx context.Context,
	config *Config) *Source {

	ctx, cancel := context.WithCancel(ctx)

	if config == nil {
		panic("config cannot be nil")
	}

	// Reuse random data, it gets costly to generate it every time at scale
	randomData := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		data := make([]byte, config.SourceConfig.BytesPerMessage)
		_, err := rand.Read(data)
		if err != nil {
			panic(err)
		}
		randomData[i] = data
	}

	return &Source{
		ctx:                 ctx,
		cancel:              cancel,
		config:              config,
		randomData:          randomData,
		connectionStartTime: make(map[int]time.Time),
		nextMessageIndex:    make(map[int]uint64),
	}
}

func (s *Source) Start() error {
	go s.monitor()

	fmt.Printf("listening on :%d\n", s.config.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	fmt.Println("Creating GRPC server")

	grpcServer := grpc.NewServer()

	lightnode.RegisterSourceServer(grpcServer, s)

	fmt.Println("attaching GRPC server to socket")

	go func() {
		err = grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("Server online")

	// No need to do anything, destination will initiate
	select {
	case <-s.ctx.Done():
		fmt.Println("context cancelled, exiting")
		grpcServer.GracefulStop()
	}

	fmt.Println("Server shutting down")

	return nil
}

func (s *Source) monitor() {
	// Sleep for a while before starting to capture metrics
	time.Sleep(time.Duration(s.config.SourceConfig.SecondsBeforeMetricsCapture) * time.Second)

	startTime := time.Now()
	s.captureMetrics.Store(true)

	fmt.Println("capturing metrics")

	// Check the number of bytes sent 100 times per second
	ticker := time.NewTicker(time.Second / 100)
	i := 0
	for {
		select {
		case <-s.ctx.Done():
			fmt.Println("context cancelled, exiting")
			return
		case <-ticker.C:
			bytesSent := s.bytesSent.Load()

			if i%100 == 0 {
				fmt.Printf("sent %.2f mb\n", float64(bytesSent)/(1<<20))
			}
			i++

			if bytesSent >= uint64(s.config.SourceConfig.GigabytesToSend)*(1<<30) {
				elapsed := time.Since(startTime)
				s.cancel()

				s.connectionMetadataLock.Lock()
				numberOfConnections := s.nextConnectionID
				s.connectionMetadataLock.Unlock()

				fmt.Printf("Total number of connections: %d\n", numberOfConnections)

				fmt.Println("-----------------------------")

				expectedBytesPerSecond := float64(
					s.config.SourceConfig.MessagesPerSecond *
						s.config.SourceConfig.BytesPerMessage *
						numberOfConnections)
				expectedKilobytesPerSecond := expectedBytesPerSecond / (1 << 10)
				expectedMegabytesPerSecond := expectedKilobytesPerSecond / (1 << 10)
				expectedGigabytesPerSecond := expectedMegabytesPerSecond / (1 << 10)

				fmt.Printf("expected %.2f B/s\n", expectedBytesPerSecond)
				fmt.Printf("expected %.2f KB/s\n", expectedKilobytesPerSecond)
				fmt.Printf("expected %.2f MB/s\n", expectedMegabytesPerSecond)
				fmt.Printf("expected %.2f GB/s\n", expectedGigabytesPerSecond)

				fmt.Println("-----------------------------")

				bytesPerSecond := float64(bytesSent) / elapsed.Seconds()
				kilobytesPerSecond := bytesPerSecond / (1 << 10)
				megabytesPerSecond := kilobytesPerSecond / (1 << 10)
				gigabytesPerSecond := megabytesPerSecond / (1 << 10)

				fmt.Printf("sent %d bytes in %v\n", bytesSent, elapsed)
				fmt.Printf("sent at %.2f B/s\n", bytesPerSecond)
				fmt.Printf("sent at %.2f KB/s\n", kilobytesPerSecond)
				fmt.Printf("sent at %.2f MB/s\n", megabytesPerSecond)
				fmt.Printf("sent at %.2f GB/s\n", gigabytesPerSecond)

				return
			}
		}
	}
}

// getUnsentMessages returns the number of messages are now due to be sent to a particular destination.
// The first integer returned is the first message index that should be sent (inclusive), and the second integer is the
// last message index that should be sent (exclusive).
func (s *Source) getUnsentMessages(connectionID int, now time.Time) (int, int) {
	s.connectionMetadataLock.Lock()
	defer s.connectionMetadataLock.Unlock()

	totalTimeElapsed := now.Sub(s.connectionStartTime[connectionID])
	totalMessagesThatShouldBeSent := int(totalTimeElapsed.Seconds() * float64(s.config.SourceConfig.MessagesPerSecond))
	remainingMessagesToSend := totalMessagesThatShouldBeSent - int(s.nextMessageIndex[connectionID])

	lowerBound := int(s.nextMessageIndex[connectionID])
	upperBound := int(s.nextMessageIndex[connectionID]) + remainingMessagesToSend

	s.nextMessageIndex[connectionID] = uint64(upperBound)

	return lowerBound, upperBound
}

// Register a new connection. Returns the ID of the connection.
func (s *Source) registerNewConnection() int {
	s.connectionMetadataLock.Lock()
	defer s.connectionMetadataLock.Unlock()

	connectionID := s.nextConnectionID
	s.nextConnectionID++

	s.connectionStartTime[connectionID] = time.Now()
	s.nextMessageIndex[connectionID] = 0

	return connectionID
}

func (s *Source) StreamData(request *lightnode.StreamDataRequest, server lightnode.Source_StreamDataServer) error {

	connectionID := s.registerNewConnection()

	period := time.Duration(1e9 / s.config.SourceConfig.PushesPerSecond)
	ticker := time.NewTicker(period)

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-server.Context().Done():
			fmt.Println("server context cancelled, exiting")
			return nil
		case <-ticker.C:

			startIndex, endIndex := s.getUnsentMessages(connectionID, time.Now())

			for dataIndex := startIndex; dataIndex < endIndex; dataIndex++ {
				data := s.randomData[dataIndex%len(s.randomData)]

				err := server.Send(&lightnode.StreamDataReply{
					Data: data,
				})
				if err != nil {
					return err
				}

				if s.captureMetrics.Load() {
					s.bytesSent.Add(uint64(len(data)))
				}
			}
		}
	}

	return nil
}

func (s *Source) newConnection(target string) (*grpc.ClientConn, *lightnode.DestinationClient, error) {
	//fmt.Printf("dialing %s\n", target)

	conn, err := grpc.DialContext(
		s.ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial: %w", err)
	}

	client := lightnode.NewDestinationClient(conn)

	//fmt.Println("connection established")

	return conn, &client, nil
}

func (s *Source) RequestPushes(ctx context.Context, request *lightnode.RequestPushesRequest) (*lightnode.RequestPushesReply, error) {
	fmt.Printf("received request for pushes\n")

	ticker := time.NewTicker(time.Second / time.Duration(s.config.SourceConfig.PushesPerSecond))

	connectionID := s.registerNewConnection()

	destination := request.Destination

	go func() {
		var conn *grpc.ClientConn
		var client *lightnode.DestinationClient

		for {

			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				if conn == nil {
					var err error
					conn, client, err = s.newConnection(destination)
					if err != nil {
						fmt.Printf("failed to connect: %v\n", err)
						return
					}
				}

				startIndex, endIndex := s.getUnsentMessages(connectionID, time.Now())
				message := lightnode.ReceiveDataRequest{
					Data: make([][]byte, 0, endIndex-startIndex),
				}

				for dataIndex := startIndex; dataIndex < endIndex; dataIndex++ {
					data := s.randomData[dataIndex%len(s.randomData)]
					message.Data = append(message.Data, data)
				}
				if s.captureMetrics.Load() {
					s.bytesSent.Add(uint64(s.config.SourceConfig.BytesPerMessage * (endIndex - startIndex)))
				}

				_, err := (*client).ReceiveData(s.ctx, &message)
				if err != nil {
					fmt.Printf("failed to send data: %v\n", err)
					return
				}

				if !s.config.SourceConfig.KeepConnectionsOpen {
					err = conn.Close()
					if err != nil {
						fmt.Printf("failed to close connection: %v\n", err)
						return
					}
					conn = nil
					client = nil
				}
			}
		}
	}()

	return &lightnode.RequestPushesReply{}, nil
}
