package reedsolomon

import (
	"context"
	"testing"

	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	recovery "github.com/Wondertan/go-ipfs-recovery"
	"github.com/Wondertan/go-ipfs-recovery/test"
)

func TestRecoverer(t *testing.T) {
	ctx := context.Background()
	dag := dstest.Mock()
	rec := NewRecoverer(ctx, dag, recovery.Requested)

	prnt := merkledag.NodeWithData([]byte("1234567890"))
	ch1 := merkledag.NodeWithData([]byte("03243423423423"))
	ch2 := merkledag.NodeWithData([]byte("123450"))
	ch3 := merkledag.NodeWithData([]byte("1234509876"))
	prnt.AddNodeLink("link", ch1)
	prnt.AddNodeLink("link", ch2)
	prnt.AddNodeLink("link", ch3)
	dag.AddMany(ctx, []format.Node{prnt, ch1, ch2, ch3})

	enc, err := Encode(ctx, dag, prnt, 2)
	require.NoError(t, err)

	dag.Remove(ctx, ch1.Cid())
	dag.Remove(ctx, ch3.Cid())

	out, err := rec.Recover(ctx, enc, ch1.Cid(), ch3.Cid())
	require.NoError(t, err)

	no := <-out
	require.NoError(t, no.Err)
	assert.Equal(t, ch1.RawData(), no.Node.RawData())

	no = <-out
	require.NoError(t, no.Err)
	assert.Equal(t, ch3.RawData(), no.Node.RawData())
}

func TestRecovererUnixFS(t *testing.T) {
	ctx := context.Background()

	bstore := blockstore.NewBlockstore(sync.MutexWrap(datastore.NewMapDatastore()))
	ex := offline.Exchange(bstore)
	dag := merkledag.NewDAGService(blockservice.New(bstore, ex))

	dr := test.NewFSDagger(t, ctx, &merkledag.ComboService{
		Read:  recovery.NewDagSession(ctx, NewRecoverer(ctx, dag, recovery.All), ex, bstore),
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
