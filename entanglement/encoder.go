package entanglement

import (
	"context"
	"fmt"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/templexxx/reedsolomon"
	"golang.org/x/tools/godoc/redirect"

	recovery "github.com/Wondertan/go-ipfs-recovery"
)

// Encode applies Reed-Solomon coding on the given IPLD Node promoting it to a recovery Node.
// Use `r` to specify needed amount of generated recovery Nodes.
func Encode(ctx context.Context, dag format.DAGService, nd format.Node, r recovery.Recoverability) (*Node, error) {
	err := ValidateNode(nd)
	if err != nil {
		return nil, err
	}

	ls := nd.Links()
	rd := NewNode(nd.(*merkledag.ProtoNode))

	nd1, err := ls[0].GetNode(ctx, dag)
	if err != nil {
		return nil, err
	}

	// create 1st redundancy and add to dag, recovery
	red := merkledag.NewRawNode(nd1.RawData())
	err = dag.Add(ctx, red)
	if err != nil {
		return nil, err
	}
	rd.AddRedundantNode(red)

	for i := 1; i < len(ls); i++ {
		nd1, err = ls[i].GetNode(ctx, dag)
		if err != nil {
			return nil, err
		}
		bs, err := XORByteSlice(nd1.RawData(), red.RawData())
		if err != nil {
			return nil, err
		}

		// create 1st redundancy and add to dag, recovery
		red = merkledag.NewRawNode(bs)
		err = dag.Add(ctx, red)
		if err != nil {
			return nil, err
		}
		// add link in order of redundancy value
		rd.AddRedundantNode(red)
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
		return fmt.Errorf("entanglement: node must be proto")
	}

	ls := nd.Links()
	if len(ls) == 0 {
		return fmt.Errorf("entanglement: node must have links")
	}

	size := ls[0].Size
	for _, l := range ls[1:] {
		if l.Size != size {
			return fmt.Errorf("entanglement: node's links must have equal size")
		}
	}

	return nil
}
