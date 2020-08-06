package entanglement

import (
	"context"
	"fmt"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/multiformats/go-varint"

	recovery "github.com/Wondertan/go-ipfs-recovery"
)

// Encode applies Reed-Solomon coding on the given IPLD Node promoting it to a recovery Node.
// Use `r` to specify needed amount of generated recovery Nodes.
func Encode(ctx context.Context, dag format.DAGService, nd format.Node, r recovery.Recoverability) (*Node, error) {
	var err error
	ls := nd.Links()
	rd := NewNode(nd.(*merkledag.ProtoNode))

	nds, s := make([]format.Node, len(rd.Links())), 0
	for i, l := range rd.Links() {
		nds[i], err = l.GetNode(ctx, dag)
		if err != nil {
			return nil, err
		}

		if len(nds[i].RawData()) > s { // finding the largest child
			s = len(nds[i].RawData())
		}
	}

	// encode size of redundant data
	s += varint.UvarintSize(uint64(s))
	ln := len(ls)
	bs := make([][]byte, ln*2)
	for i := range bs {
		bs[i] = make([]byte, s)
		if i < ln {
			l := len(nds[i].RawData())
			n := varint.PutUvarint(bs[i], uint64(l))
			copy(bs[i][n:], nds[i].RawData())
		}
	}

	for i, b := range bs[ln:] {
		if i == 0 {
			b = bs[0]
		} else {
			b, err = XORByteSlice(bs[ln+i], bs[i])
			if err != nil {
				return nil, err
			}
		}

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

// XORByteSlice returns an XOR slice of 2 input slices
func XORByteSlice(a []byte, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("length of byte slices is not equivalent: %d != %d", len(a), len(b))
	}

	buf := make([]byte, len(a))

	for i, _ := range a {
		buf[i] = a[i] ^ b[i]
	}

	return buf, nil
}
