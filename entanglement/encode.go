package entanglement

import (
	"context"
	"fmt"

	recovery "github.com/Wondertan/go-ipfs-recovery"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
)

func Encode(ctx context.Context, dag format.DAGService, nd format.Node, r recovery.Recoverability) (*Node, error) {
	// get protonode
	err := ValidateNode(nd)
	if err != nil {
		return nil, err
	}

	var nodes []format.Node
	var reds [][]byte

	// get all links of the node
	nodes = append(nodes, nd)
	ndps := format.GetDAG(ctx, dag, nd)
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
