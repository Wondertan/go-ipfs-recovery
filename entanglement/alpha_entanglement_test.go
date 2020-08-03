package entanglement

import (
	"context"
	"fmt"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	assert "github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	// Arrange
	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))
	ar := NewAlphaRestorer(dstest.Mock())
	ctx := context.Background()
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)
	ar.dag.AddMany(ctx, []format.Node{in, in2, in3})

	var ng ipld.NodeGetter = ar.dag

	ng = merkledag.NewSession(ctx, ng)

	// Act
	pnd, err := ar.Encode(ctx, in)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, pnd)

	for _, link := range pnd.Links() {
		nd, err := link.GetNode(ctx, ng)
		assert.Nil(t, err)
		fmt.Println(nd.RawData())
	}
}
