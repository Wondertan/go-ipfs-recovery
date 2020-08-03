package entanglement

import (
	"context"

	"github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"
	"github.com/templexxx/reedsolomon"

	"github.com/Wondertan/go-ipfs-recovery"
)

// Recover tries to recompute all lost IPLD Nodes using Reed-Solomon coded recovery Node.
// Pass known lost ids explicitly to avoid re-requesting them and to return corresponding Nodes on success.
func Recover(ctx context.Context, dag format.DAGService, pnd *Node, lost ...cid.Cid) ([]format.Node, error) {
	// collect ids of all linked nodes.
	lpnd := len(pnd.Links())
	lrpnd := len(pnd.RecoveryLinks())
	ids := make([]cid.Cid, lpnd+lrpnd)

outer:
	for i, l := range pnd.Links() {
		// exclude known lost ids.
		for _, id := range lost {
			if l.Cid.Equals(id) {
				ids[i] = cid.Undef
				break outer
			}
		}

		ids[i] = l.Cid
	}

	for i, l := range pnd.RecoveryLinks() {
		ids[i+lpnd] = l.Cid
	}

	// track `lost` or `not lost` blocks indexes with actual data.
	var lst, nlst []int
	bs := make([][]byte, lpnd+lrpnd)

	for i, ndp := range format.GetNodes(ctx, dag, ids) {
		nd, err := ndp.Get(ctx)
		switch err {
		case context.DeadlineExceeded, context.Canceled:
			return nil, err
		case nil:
			bs[i] = nd.RawData()
			lst = append(lst, i)
		default:
			bs[i] = make([]byte, pnd.Links()[0].Size) // the size is always the same, validated by Encode.
			nlst = append(nlst, i)
		}
	}

	if lrpnd < len(nlst) {
		return nil, recovery.ErrRecoveryExceeded
	}

	rs, err := reedsolomon.New(lpnd, lrpnd)
	if err != nil {
		return nil, err
	}

	err = rs.Reconst(bs, lst, nlst)
	if err != nil {
		return nil, err
	}

	// decode and save recomputed nodes, filter and return known to be lost.
	nds := make([]format.Node, 0, len(lost))
	for _, i := range nlst {
		id := ids[i]
		if !id.Defined() {
			id = pnd.Links()[i].Cid
		}

		b, _ := blocks.NewBlockWithCid(bs[i], id)
		nd, err := format.Decode(b)
		if err != nil {
			return nil, err
		}

		err = dag.Add(ctx, nd)
		if err != nil {
			return nil, err
		}

		if !ids[i].Defined() {
			nds = append(nds, nd)
		}
	}

	return nds, nil
}
