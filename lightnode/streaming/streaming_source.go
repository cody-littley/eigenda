package streaming

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/eigenda/api/grpc/lightnode"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

var _ lightnode.SourceServer = &Source{}

type Source struct {
	lightnode.SourceServer

	ctx    context.Context
	config *SourceConfig
}

func NewSource(
	ctx context.Context,
	config *SourceConfig) *Source {

	if config == nil {
		panic("config cannot be nil")
	}

	return &Source{
		ctx:    ctx,
		config: config,
	}
}

func (s *Source) Start() error {

	selfAddress := "0.0.0.0:12345"
	listener, err := net.Listen("tcp", selfAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcOptions := grpc.MaxRecvMsgSize(1024 * 1024 * 300) // 300 MiB
	grpcServer := grpc.NewServer(grpcOptions)

	reflection.Register(grpcServer)
	lightnode.RegisterSourceServer(grpcServer, s)

	err = grpcServer.Serve(listener)
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	switch s.config.TransferStrategy {
	case Stream:
		// No need to do anything, destination will initiate the stream
		select {
		case <-s.ctx.Done():
			fmt.Println("context cancelled, exiting")
		}
	case Put:
		// TODO implement me
		panic("implement me")
	case Get:
		// TODO implement me
		panic("implement me")
	}

	return nil
}

func (s *Source) StreamData(request *lightnode.StreamDataRequest, server lightnode.Source_StreamDataServer) error {

	period := time.Duration(1e9 / s.config.MessagesPerSecond)
	ticker := time.NewTicker(period)

	select {
	case <-s.ctx.Done():
		fmt.Println("context cancelled, exiting")
	case <-server.Context().Done():
		fmt.Println("server context cancelled, exiting")
	case <-ticker.C:
		data := make([]byte, s.config.BytesPerMessage)
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
