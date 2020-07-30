package ae

import (
	"context"
	"fmt"

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

func xorByteSlice(b1, b2 []byte) {

}
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
func (ar *AlphaRestorer) Encode(ctx context.Context, nd format.Node) (format.Node, error) {
	pnd, err := rs.ValidateNode(nd)
	if err != nil {
		return nil, err
	}

	var nodes []format.Node
	var reds [][]byte

	nodes = append(nodes, nd)
	ndps := format.GetDAG(ctx, ar.dag, pnd)
	for _, ndp := range ndps {
		nd, err = ndp.Get(ctx)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, nd)
	}

	lastRedundancy := make([]byte, len(nodes[0].RawData()))

	for _, node := range nodes {
		red, err := XORByteSlice(lastRedundancy, node.RawData())
		if err != nil {
			return nil, err
		}
		reds = append(reds, red)

		redNode := merkledag.NodeWithData(red)
		nodeProto, err := rs.ValidateNode(node)
		if err != nil {
			return nil, err
		}

		nodeProto.AddNodeLink("redundancy", redNode)

		ar.dag.Add(ctx, redNode)
	}

	return pnd, nil
}
