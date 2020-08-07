package reedsolomon

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"

	"github.com/Wondertan/go-ipfs-recovery"
)

var log = logging.Logger("recovery")

// Custom codec for Reed-Solomon recovery Nodes.
const Codec uint64 = 0x700 // random number // TODO Register in IPFS codec table.

func init() {
	// register global decoder
	format.Register(Codec, DecodeNode)

	// register codec
	cid.Codecs["recovery-reedsolomon"] = Codec
	cid.CodecToStr[Codec] = "recovery-reedsolomon"
}

type reedSolomon struct {
	dag format.DAGService
}

// NewRecoverer creates new Reed-Solomon Recoverer.
func NewRecoverer(dag format.DAGService) recovery.Recoverer {
	return &reedSolomon{dag: dag}
}

// NewEncoder creates new Reed-Solomon Encoder.
func NewEncoder(dag format.DAGService) recovery.Encoder {
	return &reedSolomon{dag: dag}
}

func (rs *reedSolomon) Recover(ctx context.Context, nd recovery.Node, rids ...cid.Cid) ([]format.Node, error) {
	pnd, ok := nd.(*Node)
	if !ok {
		return nil, fmt.Errorf("reedsolomon: wrong Node type")
	}

	return Recover(ctx, rs.dag, pnd, rids...)
}

func (rs *reedSolomon) Encode(ctx context.Context, nd format.Node, r recovery.Recoverability) (recovery.Node, error) {
	rd, ok := nd.(recovery.Node)
	if ok {
		return rd, nil
	}

	return Encode(ctx, rs.dag, nd, r)
}
