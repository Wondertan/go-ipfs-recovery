package reedsolomon

import (
	"context"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeRecover(t *testing.T) {
	ctx := context.Background()
	dag := dstest.Mock()

	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0"))
	in3 := merkledag.NodeWithData([]byte("123450"))
	in4 := merkledag.NodeWithData([]byte("1234509876"))
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)
	in.AddNodeLink("link", in4)
	dag.AddMany(ctx, []format.Node{in, in2, in3, in4})

	enc, err := Encode(ctx, dag, in, 3)
	require.NoError(t, err)

	dag.Remove(ctx, in2.Cid())
	dag.Remove(ctx, in3.Cid())
	dag.Remove(ctx, in4.Cid())

	out, err := Recover(ctx, dag, enc, in2.Cid(), in3.Cid(), in4.Cid())
	require.NoError(t, err)
	assert.Equal(t, in2.RawData(), out[0].RawData())
	assert.Equal(t, in3.RawData(), out[1].RawData())
	assert.Equal(t, in4.RawData(), out[2].RawData())

	out2, err := dag.Get(ctx, in2.Cid())
	assert.NoError(t, err)
	assert.Equal(t, in2.RawData(), out2.RawData())

	out3, err := dag.Get(ctx, in3.Cid())
	assert.NoError(t, err)
	assert.Equal(t, in3.RawData(), out3.RawData())

	out4, err := dag.Get(ctx, in4.Cid())
	assert.NoError(t, err)
	assert.Equal(t, in4.RawData(), out4.RawData())
}
