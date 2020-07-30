package ae

import (
	"context"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
)

func NewRestorer() *AlphaRestorer {
	ds := dstest.Mock()
	r := &AlphaRestorer{
		dag: ds,
	}
	return r
}

func TestEncode(t *testing.T) {
	// Arrange
	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))
	r := NewRestorer()
	ctx := context.Background()
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)
	r.dag.AddMany(ctx, []format.Node{in, in2, in3})

	// Act

	// Assert
}
