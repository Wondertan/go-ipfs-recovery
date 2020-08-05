package reedsolomon

import (
	"context"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	ctx := context.Background()
	dag := dstest.Mock()

	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)
	dag.AddMany(ctx, []format.Node{in, in2, in3})

	nd, err := Encode(ctx, dag, in, 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, nd.Recoverability())
	for _, r := range nd.RecoveryLinks() {
		assert.NotNil(t, r)
	}
}

func TestValidateNode(t *testing.T) {
	rnd := merkledag.NewRawNode([]byte("1234567890"))
	err := ValidateNode(rnd)
	assert.Error(t, err)

	nd := merkledag.NodeWithData([]byte("1234567890"))
	err = ValidateNode(nd)
	assert.Error(t, err)

	nd = merkledag.NodeWithData([]byte("1234567890"))
	nd.AddNodeLink("link", merkledag.NodeWithData([]byte("123456789")))
	nd.AddNodeLink("link", merkledag.NodeWithData([]byte("12345678")))
	err = ValidateNode(nd)
	assert.Error(t, err)
}
