package execution

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	executionv1 "github.com/rollkit/go-execution-evm/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// EVMExecution implements the ExecutionServiceClient interface using the Engine API.
type EVMExecution struct {
	client executionv1.ExecutionServiceClient
	conn   *grpc.ClientConn // Optional: To manage connection lifecycle
}

func authInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Inject the Authorization header into the outgoing context
		ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", token))
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// loadTLSCredentials loads the client-side TLS credentials from file.
func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load the client certificates from disk
	// Adjust the path and parameters as necessary
	creds, err := credentials.NewClientTLSFromFile("path/to/cert.pem", "")
	if err != nil {
		return nil, fmt.Errorf("cannot load TLS credentials: %v", err)
	}
	return creds, nil
}

// NewEVMExecution creates a new EVMExecution instance by establishing a gRPC connection.
// Use this constructor for production usage.
func NewEVMExecution(evmEndpoint string) (*EVMExecution, error) {
	// Establish a gRPC connection. Consider using secure connections (with TLS) in production.
	// Load TLS credentials
	//creds, err := loadTLSCredentials()
	//if err != nil {
	//	return nil, err
	//}
	//
	//// Create a new gRPC connection with TLS and the authorization interceptor
	//conn, err := grpc.Dial(
	//	evmEndpoint,
	//	grpc.WithTransportCredentials(creds),
	//	grpc.WithUnaryInterceptor(authInterceptor(token)),
	//)
	conn, err := grpc.NewClient(evmEndpoint, grpc.WithInsecure(), grpc.WithUnaryInterceptor(authInterceptor("c28b9140f7d8797821306c7a4fd611f20f3963238b13d624359ec0a45c24c315")))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to EVM endpoint: %w", err)
	}

	// Create a new ExecutionServiceClient from the connection.
	evmClient := executionv1.NewExecutionServiceClient(conn)

	return &EVMExecution{
		client: evmClient,
		conn:   conn,
	}, nil
}

// NewEVMExecutionWithClient creates a new EVMExecution instance with an existing ExecutionServiceClient.
// Use this constructor for testing with mocks.
func NewEVMExecutionWithClient(client executionv1.ExecutionServiceClient) *EVMExecution {
	return &EVMExecution{
		client: client,
	}
}

// Close gracefully closes the gRPC connection if it exists.
// It's good practice to call this method when shutting down your application.
func (e EVMExecution) Close() error {
	if e.conn != nil {
		return e.conn.Close()
	}
	return nil
}

// InitChain initializes the blockchain using the Engine API's forkchoiceUpdated method.
func (e EVMExecution) InitChain(ctx context.Context, req *executionv1.ForkchoiceUpdatedRequestV1) (*executionv1.ForkchoiceUpdatedResponseV1, error) {
	resp, err := e.client.EngineForkchoiceUpdatedV1(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, fmt.Errorf("forkchoiceUpdated RPC failed: %s", st.Message())
	}

	return resp, nil
}

// GetTxs retrieves transactions from the mempool.
// Since there's no direct API, this method can be a stub or integrate with your mempool management.
func (e EVMExecution) GetTxs(ctx context.Context) ([]byte, error) {
	// Stub implementation. Replace with actual mempool integration.
	return nil, nil
}

// ExecuteTxs executes transactions by calling the Engine API's newPayload and getPayload methods.
func (e EVMExecution) ExecuteTxs(ctx context.Context, req *executionv1.ExecutionPayloadV1) (*executionv1.PayloadStatusV1, error) {
	// Call engine_newPayloadV1
	statusResp, err := e.client.EngineNewPayloadV1(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, fmt.Errorf("newPayload RPC failed: %s", st.Message())
	}

	// Optionally, call engine_getPayloadV1 if needed
	// payloadReq := &executionv1.GetPayloadRequestV1{PayloadId: ...}
	// payloadResp, err := e.client.EngineGetPayloadV1(ctx, payloadReq)
	// Handle response as needed

	return statusResp, nil
}

// SetFinal marks a block as final using the Engine API's forkchoiceUpdated method.
func (e EVMExecution) SetFinal(ctx context.Context, req *executionv1.ForkchoiceUpdatedRequestV1) (*executionv1.ForkchoiceUpdatedResponseV1, error) {
	resp, err := e.client.EngineForkchoiceUpdatedV1(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, fmt.Errorf("forkchoiceUpdated RPC failed: %s", st.Message())
	}

	return resp, nil
}
