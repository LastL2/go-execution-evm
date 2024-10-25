#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Variables
REPO_NAME="go-execution-evm"
EVM_EXECUTOR_ADDRESS="127.0.0.1:50052"
EVM_ENDPOINT="http://127.0.0.1:8545"

# Create the repository directory
#mkdir -p $REPO_NAME
#cd $REPO_NAME

# Initialize Go module
#go mod init github.com/rollkit/go-execution-evm

# Create Directory Structure
mkdir -p cmd/evm-executor
mkdir -p proto
mkdir -p pkg/client
mkdir -p pkg/execution
mkdir -p pkg/server

# Create proto/execution.proto
cat <<EOF > proto/execution.proto
syntax = "proto3";

package execution;

option go_package = "github.com/rollkit/go-execution-evm/proto;proto";

// Block represents a blockchain block.
message Block {
    string parent_hash = 1;
    uint64 number = 2;
    uint64 timestamp = 3;
    repeated bytes transactions = 4;
    bytes state_root = 5;
    // Add other necessary fields as per EVM Engine API.
}

// Execution service messages.

message InitChainRequest {
    string genesis_time = 1; // ISO8601 format
    uint64 initial_height = 2;
    string chain_id = 3;
    Block genesis_block = 4;
}

message InitChainResponse {
    bytes state_root = 1;
    uint64 max_bytes = 2;
    string error = 3;
}

message GetTxsRequest {
    uint64 max_bytes = 1;
}

message GetTxsResponse {
    repeated bytes txs = 1;
}

message ExecuteTxsRequest {
    Block block = 1;
}

message ExecuteTxsResponse {
    bytes updated_state_root = 1;
    uint64 max_bytes = 2;
    string error = 3;
}

message SetFinalRequest {
    uint64 block_height = 1;
}

message SetFinalResponse {
    string error = 1;
}

// Execution Service Definition
service ExecutionService {
    rpc InitChain (InitChainRequest) returns (InitChainResponse);
    rpc GetTxs (GetTxsRequest) returns (GetTxsResponse);
    rpc ExecuteTxs (ExecuteTxsRequest) returns (ExecuteTxsResponse);
    rpc SetFinal (SetFinalRequest) returns (SetFinalResponse);
}
EOF

# Generate Go code from protobuf
protoc --go_out=. --go-grpc_out=. proto/execution.proto

# Create pkg/execution/evm.go
cat <<'EOF' > pkg/execution/evm.go
package execution

import (
    "context"
    "encoding/hex"
    "fmt"

    "github.com/rollkit/go-execution-evm/proto"
    "github.com/rollkit/go-execution-evm/pkg/client"
)

// EVMExecution implements the Execution interface using the EVM Engine API.
type EVMExecution struct {
    EVMClient *client.EVMClient
}

// NewEVMExecution creates a new EVMExecution instance.
func NewEVMExecution(evmEndpoint string) (*EVMExecution, error) {
    evmClient, err := client.NewEVMClient(evmEndpoint)
    if err != nil {
        return nil, fmt.Errorf("failed to create EVM client: %w", err)
    }
    return &EVMExecution{
        EVMClient: evmClient,
    }, nil
}

// InitChain initializes the blockchain using the EVM Engine API's forkchoiceUpdated method.
func (e *EVMExecution) InitChain(ctx context.Context, req *proto.InitChainRequest) (*proto.InitChainResponse, error) {
    // Prepare forkchoiceUpdated parameters based on the genesis block.
    forkChoiceReq := client.ForkChoiceUpdatedRequest{
        HeadBlockHash:      req.GenesisBlock.ParentHash,
        FinalizedBlockHash: req.GenesisBlock.ParentHash,
        SafeBlockHash:      req.GenesisBlock.ParentHash,
        // Add other necessary fields as per Engine API spec.
    }

    // Call the EVM Engine API's forkchoiceUpdated
    resp, err := e.EVMClient.ForkChoiceUpdated(ctx, &forkChoiceReq)
    if err != nil {
        return &proto.InitChainResponse{
            Error: err.Error(),
        }, fmt.Errorf("forkchoiceUpdated failed: %w", err)
    }

    // Process response to extract state root and max bytes
    var stateRoot [32]byte
    decodedStateRoot, err := hex.DecodeString(resp.PayloadStatus) // Placeholder: Replace with actual state root extraction.
    if err != nil {
        return &proto.InitChainResponse{
            Error: "invalid state root format",
        }, fmt.Errorf("invalid state root format: %w", err)
    }
    copy(stateRoot[:], decodedStateRoot)

    return &proto.InitChainResponse{
        StateRoot: stateRoot[:],
        MaxBytes:  resp.MaxBytes,
    }, nil
}

// GetTxs fetches transactions from the mempool.
func (e *EVMExecution) GetTxs(ctx context.Context, req *proto.GetTxsRequest) (*proto.GetTxsResponse, error) {
    // Fetch mempool transactions up to maxBytes.
    txs, err := e.EVMClient.GetMempoolTxs(ctx, req.MaxBytes)
    if err != nil {
        return &proto.GetTxsResponse{
            Txs: [][]byte{},
        }, fmt.Errorf("GetMempoolTxs failed: %w", err)
    }

    return &proto.GetTxsResponse{
        Txs: txs,
    }, nil
}

// ExecuteTxs executes transactions by proposing and retrieving a new payload.
func (e *EVMExecution) ExecuteTxs(ctx context.Context, req *proto.ExecuteTxsRequest) (*proto.ExecuteTxsResponse, error) {
    // Map ExecuteTxsRequest to Engine API's newPayloadV1 and getPayloadV1.

    // 1. Create a new payload request based on the block.
    newPayloadReq := client.NewPayloadRequest{
        ParentHash:   req.Block.ParentHash,
        BlockNumber:  req.Block.Number,
        Timestamp:    req.Block.Timestamp,
        Transactions: req.Block.Transactions,
        StateRoot:    hex.EncodeToString(req.Block.StateRoot),
        // Add other necessary fields like ReceiptsRoot, LogsBloom, etc.
    }

    // 2. Call newPayloadV1 to propose a new payload.
    newPayloadResp, err := e.EVMClient.NewPayload(ctx, &newPayloadReq)
    if err != nil {
        return &proto.ExecuteTxsResponse{
            Error: err.Error(),
        }, fmt.Errorf("newPayloadV1 failed: %w", err)
    }

    // 3. Call getPayloadV1 to retrieve the payload status.
    getPayloadReq := client.GetPayloadRequest{
        BlockHash: newPayloadResp.BlockHash,
    }

    getPayloadResp, err := e.EVMClient.GetPayload(ctx, &getPayloadReq)
    if err != nil {
        return &proto.ExecuteTxsResponse{
            Error: err.Error(),
        }, fmt.Errorf("getPayloadV1 failed: %w", err)
    }

    // 4. Process the getPayloadV1 response to extract the updated state root.
    updatedStateRootBytes, err := hex.DecodeString(getPayloadResp.Payload.StateRoot)
    if err != nil {
        return &proto.ExecuteTxsResponse{
            Error: "invalid updated state root format",
        }, fmt.Errorf("invalid updated state root format: %w", err)
    }
    var updatedStateRoot [32]byte
    copy(updatedStateRoot[:], updatedStateRootBytes)

    return &proto.ExecuteTxsResponse{
        UpdatedStateRoot: updatedStateRoot[:],
        MaxBytes:         getPayloadResp.MaxBytes,
    }, nil
}

// SetFinal marks a block as final using the EVM Engine API's forkchoiceUpdated method.
func (e *EVMExecution) SetFinal(ctx context.Context, req *proto.SetFinalRequest) (*proto.SetFinalResponse, error) {
    // Prepare forkchoiceUpdated parameters to mark the block as finalized.
    // Fetch the current head block hash if needed.

    // Placeholder values:
    headBlockHash := "0xcurrentheadblockhash"      // Replace with actual head block hash.
    finalizedBlockHash := fmt.Sprintf("0xblockhash%d", req.BlockHeight) // Replace with actual block hash.

    forkChoiceReq := client.ForkChoiceUpdatedRequest{
        HeadBlockHash:      headBlockHash,
        FinalizedBlockHash: finalizedBlockHash,
        SafeBlockHash:      headBlockHash,
        // Populate other necessary fields as per Engine API spec.
    }

    // Call the EVM Engine API's forkchoiceUpdated
    resp, err := e.EVMClient.ForkChoiceUpdated(ctx, &forkChoiceReq)
    if err != nil {
        return &proto.SetFinalResponse{
            Error: err.Error(),
        }, fmt.Errorf("forkchoiceUpdated (SetFinal) failed: %w", err)
    }

    // Optionally, process the response if needed.

    return &proto.SetFinalResponse{}, nil
}
