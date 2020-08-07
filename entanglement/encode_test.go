package entanglement

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
	in2 := merkledag.NodeWithData([]byte("987654321"))
	in3 := merkledag.NodeWithData([]byte("1khgk234509876"))
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)
	dag.AddMany(ctx, []format.Node{in, in2, in3})

	nd, err := Encode(ctx, dag, in, 3)
	assert.NoError(t, err)
	assert.Equal(t, 2, nd.Recoverability())

	for _, r := range nd.RecoveryLinks() {
		assert.NotNil(t, r)
	}
	r1, _ := nd.RecoveryLinks()[0].GetNode(ctx, dag)
	r2, _ := nd.RecoveryLinks()[1].GetNode(ctx, dag)
	rec, _ := XORByteSlice(r1.RawData(), r2.RawData())

	assert.Equal(t, in2.RawData(), r1.RawData()[1:12])
	assert.Equal(t, in3.RawData(), rec[1:])
}
