package entanglement

import (
	"context"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	assert "github.com/stretchr/testify/assert"
)

func FilledTestDag() (format.DAGService, format.Node) {
	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))
	in4 := merkledag.NodeWithData([]byte("0192837465"))
	in5 := merkledag.NodeWithData([]byte("0392817465"))
	dag := dstest.Mock()
	ctx := context.Background()
	in3.AddNodeLink("link", in4)
	in2.AddNodeLink("link", in5)
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)

	dag.AddMany(ctx, []format.Node{in, in2, in3, in4, in5})

	return dag, in
}

func TestNewEncoder(t *testing.T) {
	// Arrange
	dag, in := FilledTestDag()

	// Act
	enc, err := NewEncoder(dag, in)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, enc)

	ent := enc.(*entangler)

	assert.Equal(t, ent.Length, 5)
	for i := 0; i < ent.Length; i++ {
		assert.NotNil(t, ent.LTbl[i+1])
	}
}

func TestEncodeDag(t *testing.T) {
	// Arrange
	ctx := context.Background()
	dag, in := FilledTestDag()
	enc, _ := NewEncoder(dag, in)
	ent := enc.(*entangler)

	// Act
	err := ent.EncodeDag(ctx)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, ent.Parities)
	for i := 0; i < 15; i++ {
		assert.NotNil(t, ent.ParityMemory[i])
	}
}
