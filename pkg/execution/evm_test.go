package execution

import (
	"testing"
	"time"

	"github.com/rollkit/rollkit/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEngineAPIExecutionClient is a mock for the EngineAPIExecutionClient.
type MockEngineAPIExecutionClient struct {
	mock.Mock
}

func (m *MockEngineAPIExecutionClient) InitChain(genesisTime time.Time, initialHeight uint64, chainID string) (types.Hash, uint64, error) {
	args := m.Called(genesisTime, initialHeight, chainID)
	return args.Get(0).(types.Hash), args.Get(1).(uint64), args.Error(2)
}

func (m *MockEngineAPIExecutionClient) GetTxs() ([]types.Tx, error) {
	args := m.Called()
	return args.Get(0).([]types.Tx), args.Error(1)
}

func (m *MockEngineAPIExecutionClient) ExecuteTxs(txs []types.Tx, blockHeight uint64, timestamp time.Time, prevStateRoot types.Hash) (types.Hash, uint64, error) {
	args := m.Called(txs, blockHeight, timestamp, prevStateRoot)
	return args.Get(0).(types.Hash), args.Get(1).(uint64), args.Error(2)
}

func (m *MockEngineAPIExecutionClient) SetFinal(blockHeight uint64) error {
	args := m.Called(blockHeight)
	return args.Error(0)
}

func TestEVMExecution_InitChain(t *testing.T) {
	mockClient := new(MockEngineAPIExecutionClient)
	genesisTime := time.Now()
	initialHeight := uint64(1)
	chainID := "test-chain"
	expectedStateRoot := types.Hash{0xde, 0xad, 0xbe, 0xef}
	maxBytes := uint64(1000000)

	// Set up the mock expectations
	mockClient.On("InitChain", genesisTime, initialHeight, chainID).Return(expectedStateRoot, maxBytes, nil)

	evmExec := NewEVMExecutionWithClient(mockClient)
	stateRoot, resultMaxBytes, err := evmExec.InitChain(genesisTime, initialHeight, chainID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedStateRoot, stateRoot)
	assert.Equal(t, maxBytes, resultMaxBytes)
	mockClient.AssertExpectations(t)
}

func TestEVMExecution_GetTxs(t *testing.T) {
	mockClient := new(MockEngineAPIExecutionClient)
	expectedTxs := []types.Tx{[]byte("tx1"), []byte("tx2")}

	// Set up the mock expectations
	mockClient.On("GetTxs").Return(expectedTxs, nil)

	evmExec := NewEVMExecutionWithClient(mockClient)
	txs, err := evmExec.GetTxs()

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedTxs, txs)
	mockClient.AssertExpectations(t)
}

func TestEVMExecution_ExecuteTxs(t *testing.T) {
	mockClient := new(MockEngineAPIExecutionClient)
	txs := []types.Tx{[]byte("tx1"), []byte("tx2")}
	blockHeight := uint64(1)
	timestamp := time.Now()
	prevStateRoot := types.Hash{0xca, 0xfe, 0xba, 0xbe}
	expectedStateRoot := types.Hash{0xde, 0xad, 0xbe, 0xef}
	maxBytes := uint64(1000000)

	// Set up the mock expectations
	mockClient.On("ExecuteTxs", txs, blockHeight, timestamp, prevStateRoot).Return(expectedStateRoot, maxBytes, nil)

	evmExec := NewEVMExecutionWithClient(mockClient)
	updatedStateRoot, resultMaxBytes, err := evmExec.ExecuteTxs(txs, blockHeight, timestamp, prevStateRoot)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedStateRoot, updatedStateRoot)
	assert.Equal(t, maxBytes, resultMaxBytes)
	mockClient.AssertExpectations(t)
}

func TestEVMExecution_SetFinal(t *testing.T) {
	mockClient := new(MockEngineAPIExecutionClient)
	blockHeight := uint64(1)

	// Set up the mock expectations
	mockClient.On("SetFinal", blockHeight).Return(nil)

	evmExec := NewEVMExecutionWithClient(mockClient)
	err := evmExec.SetFinal(blockHeight)

	// Assertions
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestEVMExecution_InitChainIntegration(t *testing.T) {
	// Get node URLs from environment variables.
	ethURL := "http://localhost:8545"
	engineURL := "http://localhost:8551"

	//if ethURL == "" || engineURL == "" {
	//	t.Skip("ETH_NODE_URL or ENGINE_NODE_URL is not set, skipping integration test")
	//}

	// Create an instance of EVMExecution with actual node URLs
	evmExec := NewEVMExecution(ethURL, engineURL)
	if evmExec == nil {
		t.Fatalf("failed to initialize EVMExecution")
	}

	// Set parameters for InitChain
	genesisTime := time.Now()
	initialHeight := uint64(1)
	chainID := "test-chain"

	// Call InitChain and assert results
	stateRoot, maxBytes, err := evmExec.InitChain(genesisTime, initialHeight, chainID)

	// Assertions to verify results from the live node
	assert.NoError(t, err, "InitChain should not return an error")
	assert.NotEqual(t, types.Hash{}, stateRoot, "stateRoot should not be empty")
	assert.Greater(t, maxBytes, uint64(0), "maxBytes should be greater than zero")
}
