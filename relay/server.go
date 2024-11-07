package relay

import (
	"context"
	"fmt"
	pb "github.com/Layr-Labs/eigenda/api/grpc/relay"
	"github.com/Layr-Labs/eigenda/core/v2"
	"github.com/Layr-Labs/eigenda/disperser/common/v2/blobstore"
	"github.com/Layr-Labs/eigenda/relay/chunkstore"
	"github.com/Layr-Labs/eigensdk-go/logging"
)

var _ pb.RelayServer = &Server{}

// Server implements the Relay service defined in api/proto/relay/relay.proto
type Server struct {
	pb.UnimplementedRelayServer

	config      *Config
	blobStore   *blobstore.BlobStore
	chunkReader *chunkstore.ChunkReader

	// metadataServer encapsulates logic for fetching metadata for blobs.
	metadataServer *metadataServer

	// blobServer encapsulates logic for fetching blobs.
	blobServer *blobServer
}

// NewServer creates a new relay Server.
func NewServer(
	ctx context.Context,
	logger logging.Logger,
	config *Config,
	metadataStore *blobstore.BlobMetadataStore,
	blobStore *blobstore.BlobStore,
	chunkReader *chunkstore.ChunkReader) (*Server, error) {

	ms, err := newMetadataServer(
		ctx,
		logger,
		metadataStore,
		config.MetadataCacheSize,
		config.MetadataWorkPoolSize,
		config.Shards)
	if err != nil {
		return nil, fmt.Errorf("error creating metadata server: %w", err)
	}

	bs, err := newBlobServer(
		ctx,
		logger,
		blobStore,
		config.BlobCacheSize,
		config.BlobWorkPoolSize)
	if err != nil {
		return nil, fmt.Errorf("error creating blob server: %w", err)
	}

	return &Server{
		config:         config,
		metadataServer: ms,
		blobServer:     bs,
		blobStore:      blobStore,
		chunkReader:    chunkReader,
	}, nil
}

// GetBlob retrieves a blob stored by the relay.
func (s *Server) GetBlob(ctx context.Context, request *pb.GetBlobRequest) (*pb.GetBlobReply, error) {

	// TODO:
	//  - global throttle requests / sec
	//  - per-connection throttle requests / sec
	//  - timeouts

	keys := []v2.BlobKey{v2.BlobKey(request.BlobKey)}
	metadataMap, err := s.metadataServer.GetMetadataForBlobs(keys)
	if err != nil {
		return nil, fmt.Errorf(
			"error fetching metadata for blob, check if blob exists and is assigned to this relay: %w", err)
	}
	metadata := (*metadataMap)[v2.BlobKey(request.BlobKey)]
	if metadata == nil {
		return nil, fmt.Errorf("blob not found")
	}

	// TODO
	//  - global bytes/sec throttle
	//  - per-connection bytes/sec throttle

	key := v2.BlobKey(request.BlobKey)
	data, err := s.blobServer.GetBlob(key)
	if err != nil {
		return nil, fmt.Errorf("error fetching blob: %w", err)
	}

	reply := &pb.GetBlobReply{
		Blob: data,
	}

	return reply, nil
}

// GetChunks retrieves chunks from blobs stored by the relay.
func (s *Server) GetChunks(context.Context, *pb.GetChunksRequest) (*pb.GetChunksReply, error) {

	// Future work: rate limiting
	// Future work: authentication
	// TODO: max request size
	// TODO: limit parallelism

	return nil, nil // TODO
}
