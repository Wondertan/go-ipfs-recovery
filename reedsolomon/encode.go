package reedsolomon

import (
	"context"

	"github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/templexxx/reedsolomon"

	recovery "github.com/Wondertan/go-ipfs-recovery"
)

// Encode applies Reed-Solomon coding on the given IPLD Node promoting it to a recovery Node.
// Use `r` to specify needed amount of generated recovery Nodes.
func Encode(ctx context.Context, dag format.DAGService, nd format.Node, r recovery.Recoverability) (*Node, error) {
	rd, err := NewNode(nd)
	if err != nil {
		return nil, err
	}

	nds, s := make([]format.Node, len(nd.Links())), 0
	for i, l := range nd.Links() {
		nds[i], err = l.GetNode(ctx, dag)
		if err != nil {
			return nil, err
		}

		if len(nds[i].RawData()) > s { // finding the biggest child
			s = len(nds[i].RawData())
		}
	}

	l := len(nd.Links())
	bs := make([][]byte, l+r)
	for i := range bs {
		bs[i] = make([]byte, s)
		if i < l {
			copy(bs[i], nds[i].RawData())
		}
	}

	rs, err := reedsolomon.New(l, r)
	if err != nil {
		return nil, err
	}

	err = rs.Encode(bs)
	if err != nil {
		return nil, err
	}

	for _, b := range bs[l:] {
		rnd := merkledag.NewRawNode(b)
		err = dag.Add(ctx, rnd)
		if err != nil {
			return nil, err
		}

		rd.AddRedundantNode(rnd)
	}

	err = dag.Add(ctx, rd)
	if err != nil {
		return nil, err
	}

	return rd, dag.Remove(ctx, nd.Cid()) // there is no need to keep original
}
