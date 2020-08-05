package entanglement

import (
	"context"
	"fmt"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	assert "github.com/stretchr/testify/assert"
)

func TestNewEncoder(t *testing.T) {
	// Arrange
	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))
	dag := dstest.Mock()
	ctx := context.Background()
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)
	dag.AddMany(ctx, []format.Node{in, in2, in3})

	// Act
	enc, err := NewEncoder(dag, in)
	ent := enc.(*entangler)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, enc)

	assert.Equal(t, ent.Length, 3)
	for i := 0; i < ent.Length; i++ {
		assert.NotNil(t, ent.LTbl[i+1])
	}
}
