package rs

import (
	"context"
	"testing"

	restore "github.com/Wondertan/go-ipfs-restore"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	// Arrange
	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))
	r := NewRestorer(dstest.Mock())
	ctx := context.Background()
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)
	r.dag.AddMany(ctx, []format.Node{in, in2, in3})

	// Act
	nd, err := r.Encode(ctx, in)

	rnd, ok := nd.(*restore.Node)

	// Assert
	for _, r := range rnd.Redundant() {
		assert.NotNil(t, r)
	}
	require.NoError(t, err)
	assert.NotNil(t, nd)
	assert.NotEqual(t, in, nd)
	assert.True(t, ok)
}

func TestValidateNode(t *testing.T) {
	// Arrange
	protoNode := merkledag.NodeWithData([]byte("1234567890"))
	protoNodeWithLink := merkledag.NodeWithData([]byte("1234567890"))
	protoNodeWithDiffLenLinks := merkledag.NodeWithData([]byte("1234567890"))
	rawNode := merkledag.NewRawNode([]byte("1234567890"))
	link1 := merkledag.NodeWithData([]byte("1234567890"))
	link2 := merkledag.NodeWithData([]byte("12345"))
	protoNodeWithLink.AddNodeLink("link", link1)
	protoNodeWithDiffLenLinks.AddNodeLink("link", link1)
	protoNodeWithDiffLenLinks.AddNodeLink("link", link2)

	// Act
	_, err1 := ValidateNode(protoNode)
	_, err2 := ValidateNode(protoNodeWithLink)
	_, err3 := ValidateNode(rawNode)
	_, err4 := ValidateNode(protoNodeWithDiffLenLinks)

	// Assert
	assert.Error(t, err1)
	assert.NoError(t, err2)
	assert.Error(t, err3)
	assert.Error(t, err4)
}
