package streaming

type Config struct {

	// The port to listen on for incoming connections.
	Port int

	// If not null, then this node will act as a source of data.
	// Valid configuration will have exactly one of SourceConfig or DestinationConfig be non-null.
	SourceConfig *SourceConfig
	// If not null, then this node will act as a destination of data.
	// Valid configuration will have exactly one of SourceConfig or DestinationConfig be non-null.
	DestinationConfig *DestinationConfig
}

const (
	Stream = 1
	Put    = 2
	Get    = 3
)

// SourceConfig is configuration for a node that acts as a source of data.
type SourceConfig struct {

	// A map of the IP address/port of the destination nodes to the number of connections to make to each.
	// Ignored if the destination initiates the RPC.
	Destinations map[string]int

	// The type of transfer strategy to use. Either Stream, Put, or Get.
	TransferStrategy int

	// The number of bytes per message.
	BytesPerMessage int

	// The IP address/port of this node.
	Hostname string

	// If the transfer strategy requires the source to initiate the transfer (i.e. Stream or Put),
	// then this is the number of messages per second to send. The number of messages per second
	// are split evenly over all connections, such that the sum of all connections' messages per second
	// is equal to this value.
	MessagesPerSecond int
}

// DestinationConfig is configuration for a node that acts as a destination of data.
type DestinationConfig struct {

	// The IP address/port of the source node. Ignored if the source initiates the RPC.
	SourceHostname string

	// The type of transfer strategy to use. Either Stream, Put, or Get.
	TransferStrategy int

	// If the transfer strategy requires the destination to initiate the transfer (i.e. Get),
	// then this is the number of messages per second to request.
	MessagesPerSecond int

	// The number of parallel connections to make to the source. Ignored if the source initiates the RPC.
	// The number of messages per second are split evenly over all connections, such that the sum of all
	// connections' messages per second is equal to the MessagesPerSecond field.
	NumberOfConnections int
}
