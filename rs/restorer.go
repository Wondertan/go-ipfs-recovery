package rs

import (
	"context"
	"fmt"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/templexxx/reedsolomon"

	restore "github.com/Wondertan/go-ipfs-restore"
)

const parity = 3

type restorer struct {
	dag format.DAGService
}

func NewRestorer(ds format.DAGService) *restorer {
	r := &restorer{
		dag: ds,
	}
	return r
}

func (r *restorer) Encode(ctx context.Context, nd format.Node) (format.Node, error) {
	pnd, err := ValidateNode(nd)
	if err != nil {
		return nil, err
	}

	l := len(pnd.Links())
	bs := make([][]byte, l+parity)
	ndps := format.GetDAG(ctx, r.dag, pnd)
	for i, ndp := range ndps {
		nd, err = ndp.Get(ctx)
		if err != nil {
			return nil, err
		}

		bs[i] = nd.RawData()
	}

	for i := 0; i < parity; i++ {
		bs[len(ndps)+i] = make([]byte, len(bs[0]))
	}

	rs, err := reedsolomon.New(l, parity)
	if err != nil {
		return nil, err
	}

	err = rs.Encode(bs)
	if err != nil {
		return nil, err
	}

	rd := restore.NewNode(pnd)
	for _, b := range bs[l:] {
		rnd := merkledag.NewRawNode(b)
		err = r.dag.Add(ctx, rnd)
		if err != nil {
			return nil, err
		}

		rd.AddRedundantNode(rnd)
	}

	return rd, r.dag.Add(ctx, rd)
}

func ValidateNode(nd format.Node) (*merkledag.ProtoNode, error) {
	pnd, ok := nd.(*merkledag.ProtoNode)
	if !ok {
		return nil, fmt.Errorf("restore: node must be proto")
	}

	ls := nd.Links()
	if len(ls) == 0 {
		return nil, fmt.Errorf("restore: node must have links")
	}

	size := ls[0].Size
	for _, l := range ls[1:] {
		if l.Size != size {
			return nil, restore.ErrSizesNotEqual
		}
	}

	return pnd, nil
}
