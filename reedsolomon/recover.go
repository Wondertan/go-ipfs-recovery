package reedsolomon

import (
	"context"
	"strings"

	"github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"
	"github.com/multiformats/go-multihash"
	"github.com/multiformats/go-varint"
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
				ids[i] = wrongCid // this is needed to fail blockservice cid validation.
				continue outer
			}
		}

		ids[i] = l.Cid
	}

	for i, l := range pnd.RecoveryLinks() {
		ids[i+lpnd] = l.Cid
	}

	// track `not lost` or `lost` blocks indexes with actual data.
	var nlst, lst []int
	bs, s := make([][]byte, lpnd+lrpnd), pnd.RecoveryLinks()[0].Size

	for i, ndp := range format.GetNodes(ctx, dag, ids) {
		nd, err := ndp.Get(ctx)
		switch err {
		case nil:
			bs[i] = make([]byte, s)
			if i < lpnd {
				n := varint.PutUvarint(bs[i], uint64(len(nd.RawData())))
				copy(bs[i][n:], nd.RawData())
			} else {
				copy(bs[i], nd.RawData())
			}

			nlst = append(nlst, i)
			continue
		case context.DeadlineExceeded, context.Canceled:
			return nil, err
		default:
			bs[i] = make([]byte, s)
			lst = append(lst, i)
		}

		if !strings.Contains(err.Error(), "failed to fetch all nodes") { //
			log.Errorf("Failed to get child Node(%s) for recovery: %s", ids[i].String(), err)
		}
	}

	if lrpnd < len(lst) {
		return nil, recovery.ErrRecoveryExceeded
	}

	rs, err := reedsolomon.New(lpnd, lrpnd)
	if err != nil {
		return nil, err
	}

	err = rs.Reconst(bs, nlst, lst)
	if err != nil {
		return nil, err
	}

	// decode and save recomputed nodes, filter and return known to be lost.
	nds := make([]format.Node, 0, len(lost))
	for _, i := range lst {
		id := ids[i]
		if id.Equals(wrongCid) {
			id = pnd.Links()[i].Cid
		}

		var b blocks.Block
		if i < lpnd {
			s, n, err := varint.FromUvarint(bs[i])
			if err != nil {
				return nil, err
			}

			b, _ = blocks.NewBlockWithCid(bs[i][n:int(s)+n], id)
		} else {
			b, _ = blocks.NewBlockWithCid(bs[i], id)
		}

		nd, err := format.Decode(b)
		if err != nil {
			return nil, err
		}

		err = dag.Add(ctx, nd)
		if err != nil {
			return nil, err
		}

		if ids[i].Equals(wrongCid) {
			nds = append(nds, nd)
		}
	}

	return nds, nil
}

var wrongCid, _ = cid.Prefix{
	Version:  1,
	Codec:    1,
	MhType:   multihash.SHA2_256,
	MhLength: 1,
}.Sum([]byte("f"))
