package reedsolomon

import (
	"context"
	"fmt"

	"github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/templexxx/reedsolomon"

	recovery "github.com/Wondertan/go-ipfs-recovery"
)

// Encode applies Reed-Solomon coding on the given IPLD Node promoting it to a recovery Node.
// Use `r` to specify needed amount of generated recovery Nodes.
func Encode(ctx context.Context, dag format.DAGService, nd format.Node, r recovery.Recoverability) (*Node, error) {
	err := ValidateNode(nd)
	if err != nil {
		return nil, err
	}

	l := len(nd.Links())
	bs := make([][]byte, l+r)
	for i, ndp := range format.GetDAG(ctx, dag, nd) {
		nd, err := ndp.Get(ctx)
		if err != nil {
			return nil, err
		}

		bs[i] = nd.RawData()
	}

	for i := range bs[l:] {
		bs[i+l] = make([]byte, nd.Links()[0].Size)
	}

	rs, err := reedsolomon.New(l, r)
	if err != nil {
		return nil, err
	}

	err = rs.Encode(bs)
	if err != nil {
		return nil, err
	}

	rd := NewNode(nd.(*merkledag.ProtoNode))
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

	return rd, dag.Remove(ctx, nd.Cid())
}

// ValidateNode checks whenever the given IPLD Node can be applied with Reed-Solomon coding.
func ValidateNode(nd format.Node) error {
	_, ok := nd.(*merkledag.ProtoNode)
	if !ok {
		return fmt.Errorf("reedsolomon: node must be proto")
	}

	ls := nd.Links()
	if len(ls) == 0 {
		return fmt.Errorf("reedsolomon: node must have links")
	}

	size := ls[0].Size
	for _, l := range ls[1:] {
		if l.Size != size {
			return fmt.Errorf("reedsolomon: node's links must have equal size")
		}
	}

	return nil
}
