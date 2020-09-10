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

func TestShards(t *testing.T) {
	ctx := context.Background()
	dag := dstest.Mock()

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

	sh, err := newShards(enc)
	require.NoError(t, err)

	sh.Fill(ch2)
	rnd1,_  := dag.Get(ctx, enc.RecoveryLinks()[0].Cid)
	sh.Fill(rnd1)
	rnd2,_  := dag.Get(ctx, enc.RecoveryLinks()[1].Cid)
	sh.Fill(rnd2)

	err = sh.Want(ch1.Cid())
	require.NoError(t, err)
	err = sh.Want(ch3.Cid())
	require.NoError(t, err)

	out1, err := sh.Get(ch1.Cid())
	require.NoError(t, err)
	assert.Equal(t, ch1.RawData(), out1.RawData())

	out2, err := sh.Get(ch2.Cid())
	require.NoError(t, err)
	assert.Equal(t, ch2.RawData(), out2.RawData())
}
