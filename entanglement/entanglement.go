package entanglement

import (
	"context"
	"fmt"

	recovery "github.com/Wondertan/go-ipfs-recovery"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
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
	dag format.DAGService
}

// NewRestorer creates a new Entanglement Recoverer.
func NewRecoverer(dag format.DAGService) recovery.Recoverer {
	return &entangler{dag: dag}
}

// NewEncoder creates new Entanglement Encoder.
func NewEncoder(dag format.DAGService) recovery.Encoder {
	return &entangler{dag: dag}
}

func (ent *entangler) Recover(ctx context.Context, nd recovery.Node, rids ...cid.Cid) ([]format.Node, error) {
	pnd, ok := nd.(*Node)
	if !ok {
		return nil, fmt.Errorf("Entanglement: wrong Node type")
	}

	return Recover(ctx, ent.dag, pnd, rids...)
}

func (ent *entangler) Encode(ctx context.Context, nd format.Node, r recovery.Recoverability) (recovery.Node, error) {

	return nil, nil
}

// XORByteSlice returns an XOR slice of 2 input slices
func XORByteSlice(a []byte, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("length of byte slices is not equivalent: %d != %d", len(a), len(b))
	}

	buf := make([]byte, len(a))

	for i := range a {
		buf[i] = a[i] ^ b[i]
	}

	return buf, nil
}
