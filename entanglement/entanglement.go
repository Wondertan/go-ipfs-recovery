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
	dag          format.DAGService
	Length       int                // length of all nodes in the DAG
	LTbl         map[int]cid.Cid    // Lookup Table for lattice position
	Parities     map[[2]int]cid.Cid // Lookup table for parity [from pos, to pos]
	ParityMemory [15]*RedundantNode // 5 left strands + 5 right + 5 horizontal
}

// orderDagNode gives a lattice position to all nodes in the DAG
func (ent *entangler) orderDagNode(nd format.Node) error {
	ctx := context.Background()
	err := ValidateNode(nd) // Variable length sizes are passing!!
	if err != nil {
		return err
	}

	for _, l := range nd.Links() {
		c, err := ent.dag.Get(ctx, l.Cid)
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

	return nil, nil
}

func (ent *entangler) updateParities(ctx context.Context, r, h, l *RedundantNode) {
	ent.Parities[r.Position] = r.Cid()
	ent.Parities[h.Position] = h.Cid()
	ent.Parities[l.Position] = l.Cid()
	ent.dag.AddMany(ctx, []format.Node{r, h, l})
}

func (ent *entangler) entangle(ctx context.Context, nd *Node, i int) {
	r, h, l := GetMemoryPosition(i)
	rBack, hBack, lBack := GetBackwardNeighbours(i)
	rParity := ent.ParityMemory[r]
	hParity := ent.ParityMemory[h]
	lParity := ent.ParityMemory[l]

	rNext, _ := XORByteSlice(nd.RawData(), rParity.Data())
	rnd := NewRedundantNode(merkledag.NodeWithData(rNext), [2]int{rBack, i})
	ent.ParityMemory[r] = rnd

	hNext, _ := XORByteSlice(nd.RawData(), hParity.Data())
	hnd := NewRedundantNode(merkledag.NodeWithData(hNext), [2]int{hBack, i})
	ent.ParityMemory[r] = hnd

	lNext, _ := XORByteSlice(nd.RawData(), lParity.Data())
	lnd := NewRedundantNode(merkledag.NodeWithData(lNext), [2]int{lBack, i})
	ent.ParityMemory[r] = lnd
	fmt.Println(rnd.Cid(), hnd.Cid(), lnd.Cid())

	ent.dag.AddMany(ctx, []format.Node{rnd, hnd, lnd})
}

// EncodeDag creates an entangled lattice of redundancies
func (ent *entangler) EncodeDag(ctx context.Context) error {
	for i := 1; i < ent.Length; i++ {
		nd, err := ent.dag.Get(ctx, ent.LTbl[1])
		if err != nil {
			return err
		}
		fmt.Println("BRHKBA")
		ent.entangle(ctx, nd.(*Node), i)
	}

	return nil
}
