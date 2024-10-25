package execution

import (
	"context"
	"testing"

	executionv1 "github.com/rollkit/go-execution-evm/proto/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockEVMClient mocks the ExecutionServiceClient interface for testing.
type MockEVMClient struct {
	mock.Mock
}

// EngineNewPayloadV1 mocks the engine_newPayloadV1 RPC method.
func (m *MockEVMClient) EngineNewPayloadV1(ctx context.Context, req *executionv1.ExecutionPayloadV1, opts ...grpc.CallOption) (*executionv1.PayloadStatusV1, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*executionv1.PayloadStatusV1), args.Error(1)
}

// EngineForkchoiceUpdatedV1 mocks the engine_forkchoiceUpdatedV1 RPC method.
func (m *MockEVMClient) EngineForkchoiceUpdatedV1(ctx context.Context, req *executionv1.ForkchoiceUpdatedRequestV1, opts ...grpc.CallOption) (*executionv1.ForkchoiceUpdatedResponseV1, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*executionv1.ForkchoiceUpdatedResponseV1), args.Error(1)
}

// EngineGetPayloadV1 mocks the engine_getPayloadV1 RPC method.
func (m *MockEVMClient) EngineGetPayloadV1(ctx context.Context, req *executionv1.GetPayloadRequestV1, opts ...grpc.CallOption) (*executionv1.ExecutionPayloadV1, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*executionv1.ExecutionPayloadV1), args.Error(1)
}

// mustEmbedUnimplementedExecutionServiceClient is required to satisfy the ExecutionServiceClient interface.
func (m *MockEVMClient) mustEmbedUnimplementedExecutionServiceClient() {}

// --------------------
// Test Cases
// --------------------

// TestEVMExecution_InitChain tests the InitChain method of EVMExecution.
func TestEVMExecution_InitChain(t *testing.T) {
	mockClient := new(MockEVMClient)

	// Define the expected request and response.
	req := &executionv1.ForkchoiceUpdatedRequestV1{
		ForkchoiceState: &executionv1.ForkchoiceStateV1{
			HeadBlockHash:      []byte("0xheadblockhash"),
			SafeBlockHash:      []byte("0xsafeblockhash"),
			FinalizedBlockHash: []byte("0xfinalizedblockhash"),
		},
		PayloadAttributes: &executionv1.ForkchoiceUpdatedRequestV1_PayloadAttributesData{
			PayloadAttributesData: &executionv1.PayloadAttributesV1{
				Timestamp:             1700851200,
				PrevRandao:            []byte("0xprevrandao"),
				SuggestedFeeRecipient: []byte("0xsuggestedfeerecipient"),
			},
		},
	}

	expectedResp := &executionv1.ForkchoiceUpdatedResponseV1{
		PayloadStatus: &executionv1.PayloadStatusV1{
			Status:          executionv1.PayloadStatusV1_VALID,
			LatestValidHash: []byte("0xdummyvalidhash"),
			ValidationError: "",
		},
		PayloadId: []byte("payloadid1"),
	}

	// Set up expectations.
	mockClient.On("EngineForkchoiceUpdatedV1", mock.Anything, req).Return(expectedResp, nil)

	// Initialize EVMExecution with the mock client.
	evmExec, _ := NewEVMExecution("0.0.0.0:8551")

	// Call the InitChain method.
	resp, err := evmExec.InitChain(context.Background(), req)

	// Assertions.
	assert.NoError(t, err)
	assert.Equal(t, expectedResp.PayloadStatus, resp.PayloadStatus)
	assert.Equal(t, expectedResp.PayloadId, resp.PayloadId)

	// Assert that the expectations were met.
	mockClient.AssertExpectations(t)
}

// TestEVMExecution_ExecuteTxs tests the ExecuteTxs method of EVMExecution.
func TestEVMExecution_ExecuteTxs(t *testing.T) {
	mockClient := new(MockEVMClient)

	// Define the expected request and response.
	req := &executionv1.ExecutionPayloadV1{
		ParentHash:    []byte("0xparenthash"),
		FeeRecipient:  []byte("0xfeerecipient"),
		StateRoot:     []byte("0xstateroot"),
		ReceiptsRoot:  []byte("0xreceiptsroot"),
		LogsBloom:     []byte("0xlogsbloom"),
		PrevRandao:    []byte("0xprevrandao"),
		BlockNumber:   1,
		GasLimit:      1000000,
		GasUsed:       900000,
		Timestamp:     1700851200,
		ExtraData:     []byte("0xextradata"),
		BaseFeePerGas: []byte("0xbasefeepergas"),
		BlockHash:     []byte("0xblockhash"),
		Transactions:  [][]byte{[]byte("tx1"), []byte("tx2")},
	}

	expectedStatusResp := &executionv1.PayloadStatusV1{
		Status:          executionv1.PayloadStatusV1_VALID,
		LatestValidHash: []byte("0xdummyvalidhash"),
		ValidationError: "",
	}

	// Set up expectations.
	mockClient.On("EngineNewPayloadV1", mock.Anything, req).Return(expectedStatusResp, nil)

	// Initialize EVMExecution with the mock client.
	evmExec := NewEVMExecutionWithClient(mockClient)

	// Call the ExecuteTxs method.
	resp, err := evmExec.ExecuteTxs(context.Background(), req)

	// Assertions.
	assert.NoError(t, err)
	assert.Equal(t, expectedStatusResp, resp)

	// Assert that the expectations were met.
	mockClient.AssertExpectations(t)
}

// TestEVMExecution_GetTxs tests the GetTxs method of EVMExecution.
// Since there's no direct API, this is a stub test.
func TestEVMExecution_GetTxs(t *testing.T) {
	evmExec := NewEVMExecutionWithClient(nil) // Pass nil or implement a mock if necessary

	// Call the GetTxs method.
	resp, err := evmExec.GetTxs(context.Background())

	// Assertions.
	assert.NoError(t, err)
	assert.Nil(t, resp) // Since it's a stub
}

// TestEVMExecution_SetFinal tests the SetFinal method of EVMExecution.
func TestEVMExecution_SetFinal(t *testing.T) {
	mockClient := new(MockEVMClient)

	// Define the expected request and response.
	req := &executionv1.ForkchoiceUpdatedRequestV1{
		ForkchoiceState: &executionv1.ForkchoiceStateV1{
			HeadBlockHash:      []byte("0xheadblockhash"),
			SafeBlockHash:      []byte("0xsafeblockhash"),
			FinalizedBlockHash: []byte("0xfinalizedblockhash"),
		},
		PayloadAttributes: nil, // Assuming no new payload attributes
	}

	expectedResp := &executionv1.ForkchoiceUpdatedResponseV1{
		PayloadStatus: &executionv1.PayloadStatusV1{
			Status:          executionv1.PayloadStatusV1_INVALID,
			LatestValidHash: []byte("0xdummyvalidhash"),
			ValidationError: "Invalid fork choice",
		},
		PayloadId: nil,
	}

	// Set up expectations.
	mockClient.On("EngineForkchoiceUpdatedV1", mock.Anything, req).Return(expectedResp, nil)

	// Initialize EVMExecution with the mock client.
	evmExec := NewEVMExecutionWithClient(mockClient)

	// Call the SetFinal method.
	resp, err := evmExec.SetFinal(context.Background(), req)

	// Assertions.
	assert.NoError(t, err)
	assert.Equal(t, expectedResp.PayloadStatus, resp.PayloadStatus)
	assert.Nil(t, resp.PayloadId)

	// Assert that the expectations were met.
	mockClient.AssertExpectations(t)
}
