package entanglement

import (
	"context"
	"fmt"

	recovery "github.com/Wondertan/go-ipfs-recovery"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
)

// Custom codec for Entanglement recovery Nodes.
const Codec = 0x701 // random number // TODO Register in IPFS codec table.

func init() {
	// register global decoder
	format.Register(Codec, DecodeNode)

	// register codec
	cid.Codecs["recovery-entanglement`"] = Codec
	cid.CodecToStr[Codec] = "recovery-entanglement`"

}

type entangler struct {
	dag    format.DAGService
	Length int             // length of all nodes in the DAG
	LTbl   map[int]cid.Cid // Lookup Table for lattice position
}

// orderDagNode gives a lattice position to all nodes in the DAG
func (ent *entangler) orderDagNode(nd format.Node) error {
	err := ValidateNode(nd)
	ctx := context.Background()
	if err != nil {
		return err
	}

	for _, ndp := range format.GetDAG(ctx, ent.dag, nd) {
		c, err := ndp.Get(ctx)
		if err != nil {
			return err
		}
		ent.orderDagNode(c)
	}

	end := NewNode(nd.(*merkledag.ProtoNode), ent.Length)
	ent.Length += 1
	ent.LTbl[ent.Length] = end.Cid()
	return nil
}

// NewRestorer creates a new Entanglement Recoverer.
func NewRestorer(dag format.DAGService) recovery.Recoverer {
	return &entangler{dag: dag}
}

// NewEncoder creates new Entanglement Encoder.
func NewEncoder(dag format.DAGService, nd format.Node) (recovery.Encoder, error) {

	ent := &entangler{dag: dag, Length: 0}
	ent.LTbl = make(map[int]cid.Cid)
	err := ent.orderDagNode(nd)
	if err != nil {
		return nil, err
	}

	return ent, nil

}

func (ent *entangler) Recover(ctx context.Context, nd recovery.Node, rids ...cid.Cid) ([]format.Node, error) {
	pnd, ok := nd.(*Node)
	if !ok {
		return nil, fmt.Errorf("reedsolomon: wrong Node type")
	}

	return Recover(ctx, ent.dag, pnd, rids...)
}

func (ent *entangler) Encode(ctx context.Context, nd format.Node, r recovery.Recoverability) (recovery.Node, error) {
	// return Encode(ctx, ent.dag, nd, r, ent)
	return nil, nil
}
