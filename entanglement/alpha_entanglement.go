package entanglement

import (
	"context"
	"fmt"

	restore "github.com/Wondertan/go-ipfs-restore"
	"github.com/Wondertan/go-ipfs-restore/rs"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
)

type AlphaRestorer struct {
	dag format.DAGService
}

const (
	alpha = 1
	s     = 1
	p     = 1
)

func NewAlphaRestorer(ds format.DAGService) *AlphaRestorer {
	r := &AlphaRestorer{
		dag: ds,
	}
	return r
}

func xorByteSlice(a []byte, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("length of byte slices is not equivalent: %d != %d", len(a), len(b))
	}

	buf := make([]byte, len(a))

	for i, _ := range a {
		buf[i] = a[i] ^ b[i]
	}

	return buf, nil
}

func (ar *AlphaRestorer) Encode(ctx context.Context, nd format.Node) (format.Node, error) {
	// get protonode
	pnd, err := rs.ValidateNode(nd)
	if err != nil {
		return nil, err
	}

	var nodes []format.Node
	var reds [][]byte

	// get all links of the node
	nodes = append(nodes, nd)
	ndps := format.GetDAG(ctx, ar.dag, pnd)
	for _, ndp := range ndps {
		nd, err = ndp.Get(ctx)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, nd)
	}

	// for all links, create an alpha entanglement
	lastRedundancy := make([]byte, len(nodes[0].RawData()))

	// n.b. how does this work in terms of ordering?
	for _, linkedNode := range nodes {

		red, err := xorByteSlice(lastRedundancy, linkedNode.RawData())
		lastRedundancy = red
		if err != nil {
			return nil, err
		}
		reds = append(reds, red)

		redNode := merkledag.NodeWithData(red)

		linkedNodeProto, err := rs.ValidateNode(linkedNode)

		rnode := restore.NewNode(linkedNodeProto)
		if err != nil {
			return nil, err
		}

		rnode.AddRedundantNode(redNode)

		ar.dag.Add(ctx, redNode)
	}

	return pnd, nil
}
