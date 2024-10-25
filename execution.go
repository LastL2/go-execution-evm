package execution

import (
	"context"
	proto "github.com/rollkit/go-execution-evm/proto/v1"
)

// Execution defines the common interface for different execution engines.
type Execution interface {
	InitChain(ctx context.Context, req *proto.ForkchoiceUpdatedRequestV1) (*proto.ForkchoiceUpdatedResponseV1, error)
	GetTxs(ctx context.Context, req *proto.ExecutionPayloadV1) (*proto.PayloadStatusV1_Status, error)     //incorect defintion as we have to reap mempool
	ExecuteTxs(ctx context.Context, req *proto.ExecutionPayloadV1) (*proto.PayloadStatusV1_Status, error) // unsure on the response here tbh, just have these to let it compile
	SetFinal(ctx context.Context, req *proto.ForkchoiceUpdatedRequestV1) (*proto.ForkchoiceUpdatedResponseV1, error)
}
