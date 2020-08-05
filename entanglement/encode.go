package entanglement

import (
	"context"
	"fmt"

	recovery "github.com/Wondertan/go-ipfs-recovery"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
)

func Encode(ctx context.Context, dag format.DAGService, nd format.Node, r recovery.Recoverability, ent entangler) (*Node, error) {
	err := ValidateNode(nd)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateNode checks whenever the given IPLD Node can be applied with Entanglement coding.
func ValidateNode(nd format.Node) error {
	_, ok := nd.(*merkledag.ProtoNode)
	if !ok {
		return fmt.Errorf("reedsolomon: node must be proto")
	}

	ls := nd.Links()
	// can relax need for links
	if len(ls) != 0 {

		size := ls[0].Size
		for _, l := range ls[1:] {
			if l.Size != size {
				return fmt.Errorf("reedsolomon: node's links must have equal size")
			}
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
