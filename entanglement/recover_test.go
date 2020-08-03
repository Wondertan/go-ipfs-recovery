package entanglement

import (
	"context"
	"testing"

	"github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	// Arrange
	ctx := context.Background()

	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))

	dag := dstest.Mock()
	dag.AddMany(ctx, []format.Node{in, in2, in3})

	e := NewEncoder(dag)

	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)

	rnd, err := e.Encode(ctx, in, 2)

	nd, ok := rnd.(*Node)

	// Act
	failures, err := Recover(ctx, dag, nd)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, failures)
	assert.True(t, ok)
}

func TestRecoverOneFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()

	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))

	dag := dstest.Mock()
	dag.AddMany(ctx, []format.Node{in, in2, in3})

	e := NewEncoder(dag)

	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)

	rnd, err := e.Encode(ctx, in, 2)

	nd, ok := rnd.(*Node)

	// Act
	dag.Remove(ctx, in2.Cid())
	failures, err := Recover(ctx, dag, nd)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, failures)
	assert.True(t, ok)
	assert.Empty(t, failures)

	lnk := nd.Links()[0]
	r, err := dag.Get(ctx, lnk.Cid)

	assert.Equal(t, r.RawData(), in2.RawData())

	assert.Nil(t, err)
}

func TestRecoverTwoFailures(t *testing.T) {
	// Arrange
	ctx := context.Background()

	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))

	dag := dstest.Mock()
	dag.AddMany(ctx, []format.Node{in, in2, in3})

	e := NewEncoder(dag)

	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)

	rnd, err := e.Encode(ctx, in, 2)

	nd, ok := rnd.(*Node)

	// Act
	dag.Remove(ctx, in2.Cid())
	dag.Remove(ctx, in3.Cid())
	failures, err := Recover(ctx, dag, nd)

	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, failures)
	assert.True(t, ok)
	assert.Empty(t, failures)

	lnk := nd.Links()[0]
	r, err := dag.Get(ctx, lnk.Cid)

	lnk1 := nd.Links()[1]
	r1, err := dag.Get(ctx, lnk1.Cid)

	assert.Equal(t, r.RawData(), in2.RawData())
	assert.Equal(t, r1.RawData(), in3.RawData())

	assert.Nil(t, err)
}
