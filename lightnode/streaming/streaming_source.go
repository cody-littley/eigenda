package streaming

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/eigenda/api/grpc/lightnode"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc"
	"net"
	"time"
)

var _ lightnode.SourceServer = &Source{}

type Source struct {
	lightnode.SourceServer

	ctx    context.Context
	config *Config
}

func NewSource(
	ctx context.Context,
	config *Config) *Source {

	if config == nil {
		panic("config cannot be nil")
	}

	return &Source{
		ctx:    ctx,
		config: config,
	}
}

func (s *Source) Start() error {

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

	fmt.Println("Sever online")

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

func (s *Source) StreamData(request *lightnode.StreamDataRequest, server lightnode.Source_StreamDataServer) error {

	period := time.Duration(1e9 / s.config.SourceConfig.MessagesPerSecond)
	ticker := time.NewTicker(period)

	select {
	case <-s.ctx.Done():
		fmt.Println("context cancelled, exiting")
	case <-server.Context().Done():
		fmt.Println("server context cancelled, exiting")
	case <-ticker.C:
		data := make([]byte, s.config.SourceConfig.BytesPerMessage)
		_, err := rand.Read(data)
		if err != nil {
			return err
		}

		fmt.Printf("sending data") // TODO

		err = server.Send(&lightnode.StreamDataReply{
			Data: data,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Source) GetData(ctx context.Context, request *lightnode.GetDataRequest) (*lightnode.GetDataReply, error) {
	//TODO implement me
	panic("implement me")
}
