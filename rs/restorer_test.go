package rs

import (
	"context"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))

	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)

	ctx := context.Background()
	ds := dstest.Mock()
	r := &restorer{
		dag: ds,
	}

	r.dag.AddMany(ctx, []format.Node{in, in2, in3})

	nd, err := r.Encode(ctx, in)

	require.NoError(t, err)
	assert.NotNil(t, nd)
}
