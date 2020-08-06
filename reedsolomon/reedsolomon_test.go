package reedsolomon

import (
	"context"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	recovery "github.com/Wondertan/go-ipfs-recovery"
	"github.com/Wondertan/go-ipfs-recovery/test"
)

func TestEncodeRecover(t *testing.T) {
	ctx := context.Background()
	dag := dstest.Mock()

	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("03243423423423"))
	in3 := merkledag.NodeWithData([]byte("123450"))
	in4 := merkledag.NodeWithData([]byte("1234509876"))
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)
	in.AddNodeLink("link", in4)
	dag.AddMany(ctx, []format.Node{in, in2, in3, in4})

	enc, err := Encode(ctx, dag, in, 2)
	require.NoError(t, err)

	dag.Remove(ctx, in2.Cid())
	dag.Remove(ctx, in4.Cid())

	out, err := Recover(ctx, dag, enc, in2.Cid(), in4.Cid())
	require.NoError(t, err)
	assert.Equal(t, in2.RawData(), out[0].RawData())
	assert.Equal(t, in4.RawData(), out[1].RawData())

	out2, err := dag.Get(ctx, in2.Cid())
	assert.NoError(t, err)
	assert.Equal(t, in2.RawData(), out2.RawData())

	out4, err := dag.Get(ctx, in4.Cid())
	assert.NoError(t, err)
	assert.Equal(t, in4.RawData(), out4.RawData())
}

func TestUnixFS(t *testing.T) {
	ctx := context.Background()
	dag := dstest.Mock()
	dr := test.NewFSDagger(t, ctx, &merkledag.ComboService{
		Read:  recovery.NewNodeGetter(dag, NewRecoverer(dag)),
		Write: dag,
	})
	dr.Morpher = func(nd format.Node) (format.Node, error) {
		if len(nd.Links()) == 0 {
			return nd, nil
		}

		return Encode(ctx, dag, nd, 3)
	}

	root :=
		dr.NewDir("root",
			dr.RandNode("file1"),
			dr.RandNode("file2"),
			dr.NewDir("dir1",
				dr.RandNode("file3"),
				dr.RandNode("file4"),
			),
			dr.NewDir("dir2",
				dr.RandNode("file5"),
			),
		)

	dr.Remove("file1")
	dr.Remove("dir2")
	dr.Remove("file4")

	root.Validate()
}
