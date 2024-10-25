package execution

import (
	"fmt"
	"github.com/LastL2/go-execution"
	"github.com/ethereum/go-ethereum/common"
	"time"

	"github.com/rollkit/go-execution-evm/pkg/client"
	rollkitTypes "github.com/rollkit/rollkit/types"
	"google.golang.org/grpc/status"
)

// EVMExecution implements the Execute interface using the Engine API.
type EVMExecution struct {
	client execution.Execute
}

func NewEVMExecutionWithClient(client execution.Execute) *EVMExecution {
	return &EVMExecution{
		client: client,
	}
}

func NewEVMExecution(ethURL string, engineURL string) *EVMExecution {
	ethClient, err := client.NewEngineAPIExecutionClient(ethURL, engineURL, common.Hash{}, common.Address{})
	if err != nil {
		println(err.Error())
		return nil
	}
	return &EVMExecution{
		client: ethClient,
	}
}

// InitChain initializes the blockchain using the Engine API's forkchoiceUpdated method.
func (e EVMExecution) InitChain(genesisTime time.Time, initialHeight uint64, chainID string) (stateRoot rollkitTypes.Hash, maxBytes uint64, err error) {
	stateRoot, gasLimit, err := e.client.InitChain(genesisTime, initialHeight, chainID)
	if err != nil {
		st, _ := status.FromError(err)
		return rollkitTypes.Hash{}, uint64(0), fmt.Errorf("forkchoiceUpdated RPC failed: %s", st.Message())
	}
	return stateRoot, gasLimit, nil
}

// GetTxs retrieves all available transactions from the execution client's mempool.
func (e EVMExecution) GetTxs() ([]rollkitTypes.Tx, error) {
	txs, err := e.client.GetTxs()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}
	return txs, nil
}

// ExecuteTxs executes a set of transactions to produce a new block header.
func (e EVMExecution) ExecuteTxs(txs []rollkitTypes.Tx, blockHeight uint64, timestamp time.Time, prevStateRoot rollkitTypes.Hash) (updatedStateRoot rollkitTypes.Hash, maxBytes uint64, err error) {
	updatedStateRoot, maxBytes, err = e.client.ExecuteTxs(txs, blockHeight, timestamp, prevStateRoot)
	if err != nil {
		st, _ := status.FromError(err)
		return rollkitTypes.Hash{}, uint64(0), fmt.Errorf("execution failed: %s", st.Message())
	}
	return updatedStateRoot, maxBytes, nil
}

// SetFinal marks a block at the given height as final.
func (e EVMExecution) SetFinal(blockHeight uint64) error {
	if err := e.client.SetFinal(blockHeight); err != nil {
		st, _ := status.FromError(err)
		return fmt.Errorf("failed to mark block as final: %s", st.Message())
	}
	return nil
}
