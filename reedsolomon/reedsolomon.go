package reedsolomon

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"

	"github.com/Wondertan/go-ipfs-recovery"
)

// Custom codec for Reed-Solomon recovery Nodes.
const Codec = 0x700 // random number // TODO Register in IPFS codec table.

func init() {
	// register global decoder
	format.Register(Codec, DecodeNode)

	// register codec
	cid.Codecs["recovery-reedsolomon"] = Codec
	cid.CodecToStr[Codec] = "recovery-reedsolomon"
}

type reedsolomon struct {
	dag format.DAGService
}

// NewRestorer creates new Reed-Solomon Recoverer.
func NewRestorer(dag format.DAGService) recovery.Recoverer {
	return &reedsolomon{dag: dag}
}

// NewEncoder creates new Reed-Solomon Encoder.
func NewEncoder(dag format.DAGService) recovery.Encoder {
	return &reedsolomon{dag: dag}
}

func (rs *reedsolomon) Recover(ctx context.Context, nd recovery.Node, rids ...cid.Cid) ([]format.Node, error) {
	pnd, ok := nd.(*Node)
	if !ok {
		return nil, fmt.Errorf("reedsolomon: wrong Node type")
	}

	return Recover(ctx, rs.dag, pnd, rids...)
}

func (rs *reedsolomon) Encode(ctx context.Context, nd format.Node, r recovery.Recoverability) (recovery.Node, error) {
	return Encode(ctx, rs.dag, nd, r)
}
