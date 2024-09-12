package streaming

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/eigenda/api/grpc/lightnode"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc"
	"net"
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

	randomData [][]byte
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
		ctx:        ctx,
		cancel:     cancel,
		config:     config,
		randomData: randomData,
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

	switch s.config.SourceConfig.TransferStrategy {
	case Stream:
		// No need to do anything, destination will initiate the stream
		select {
		case <-s.ctx.Done():
			fmt.Println("context cancelled, exiting")
			grpcServer.GracefulStop()
		}
	case Put:
		// TODO implement me
		panic("implement me")
	case Get:
		// TODO implement me
		panic("implement me")
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
				fmt.Printf("sent %f.2 mb\n", float64(bytesSent)/1e6)
			}
			i++

			if bytesSent >= uint64(s.config.SourceConfig.GigabytesToSend)*1e9 {
				// TODO end the program and report
				elapsed := time.Since(startTime)
				s.cancel()

				bytesPerSecond := float64(bytesSent) / elapsed.Seconds()
				kilobytesPerSecond := bytesPerSecond / 1e3
				megabytesPerSecond := kilobytesPerSecond / 1e3
				gigabytesPerSecond := megabytesPerSecond / 1e3

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

func (s *Source) StreamData(request *lightnode.StreamDataRequest, server lightnode.Source_StreamDataServer) error {

	period := time.Duration(1e9 / s.config.SourceConfig.MessagesPerSecond)
	ticker := time.NewTicker(period)

	i := 0

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-server.Context().Done():
			fmt.Println("server context cancelled, exiting")
			return nil
		case <-ticker.C:
			data := s.randomData[i%100]
			i++
			_, err := rand.Read(data)
			if err != nil {
				return err
			}

			//fmt.Println("sending data") // TODO

			err = server.Send(&lightnode.StreamDataReply{
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

	return nil
}

func (s *Source) GetData(ctx context.Context, request *lightnode.GetDataRequest) (*lightnode.GetDataReply, error) {
	//TODO implement me
	panic("implement me")
}
