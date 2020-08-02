package rs

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/templexxx/reedsolomon"

	restore "github.com/Wondertan/go-ipfs-restore"
)

const parity = 3

type restorer struct {
	dag format.DAGService
}

func (r *restorer) Restore(ctx context.Context, id cid.Cid, tr ...cid.Cid) ([]format.Node, error) {
	nd, err := r.dag.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	rn, ok := nd.(*restore.Node)
	if !ok {
		return nil, restore.ErrNotRestorable
	}

	lrn := len(rn.Redundant())
	if lrn < len(tr) {
		return nil, fmt.Errorf("restore-RS: can't restore more nodes(%d) than the amount of redundant ones(%d)", len(ns), lrn)
	}

	lnd := len(nd.Links())
	bs := make([][]byte, lnd+lrn)
	ids := make([]cid.Cid, 0, lnd+lrn-len(tr))
	for _, l := range rn.Links() {
		ids = append(ids, l.Cid)
	}
	for _, l := range rn.Redundant() {
		ids = append(ids, l.Cid)
	}

	for i, ndp := range format.GetNodes(ctx, r.dag, ids) {
		nd, err := ndp.Get(ctx)
		if err != nil {

		}
	}
}

func (r *restorer) Encode(ctx context.Context, id cid.Cid) (cid.Cid, error) {
	nd, err := r.dag.Get(ctx, id)
	if err != nil {
		return cid.Undef, err
	}

	pnd, err := ValidateNode(nd)
	if err != nil {
		return cid.Undef, err
	}

	l := len(pnd.Links())
	bs := make([][]byte, l+parity)
	for i, ndp := range format.GetDAG(ctx, r.dag, pnd) {
		nd, err = ndp.Get(ctx)
		if err != nil {
			return cid.Undef, err
		}

		bs[i] = nd.RawData()
	}

	rs, err := reedsolomon.New(l, parity)
	if err != nil {
		return cid.Undef, err
	}

	err = rs.Encode(bs)
	if err != nil {
		return cid.Undef, err
	}

	rd := restore.NewNode(pnd)
	for _, b := range bs[l:] {
		rnd := merkledag.NewRawNode(b)
		err = r.dag.Add(ctx, rnd)
		if err != nil {
			return cid.Undef, err
		}

		rd.AddRedundantNode(rnd)
	}

	return rd.Cid(), r.dag.Add(ctx, rd)
}

func ValidateNode(nd format.Node) (*merkledag.ProtoNode, error) {
	pnd, ok := nd.(*merkledag.ProtoNode)
	if !ok {
		return nil, fmt.Errorf("restore-RS: node must be proto")
	}

	ls := nd.Links()
	if len(ls) == 0 {
		return nil, fmt.Errorf("restore-RS: node must have links")
	}

	size := ls[0].Size
	for _, l := range ls[1:] {
		if l.Size != size {
			return nil, restore.ErrSizesNotEqual
		}
	}

	return pnd, nil
}
