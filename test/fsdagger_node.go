package test

import (
	"bytes"
	"io/ioutil"

	files "github.com/ipfs/go-ipfs-files"
	format "github.com/ipfs/go-ipld-format"
	unixfile "github.com/ipfs/go-unixfs/file"
	"github.com/stretchr/testify/require"
)

type FSDaggerNode struct {
	d *FSDagger

	node  format.Node
	name  string
	isDir bool

	Data []byte
}

func (de *FSDaggerNode) IsDir() bool {
	return de.isDir
}

func (de *FSDaggerNode) File() files.Node {
	nd, err := de.d.dag.Get(de.d.ctx, de.node.Cid()) // in case DAG does some hacky node transformations :)
	require.NoError(de.d.t, err)

	f, err := unixfile.NewUnixfsFile(de.d.ctx, de.d.dag, nd)
	require.NoError(de.d.t, err)
	return f
}

func (de *FSDaggerNode) Validate() {
	nd := de.File()
	if !de.IsDir() {
		data, err := ioutil.ReadAll(files.ToFile(nd))
		require.NoError(de.d.t, err)

		if !bytes.Equal(de.Data, data) {
			de.d.t.Fatalf("Data for %s is wrong.", de.name)
		}

		return
	}

	it := files.ToDir(nd).Entries()
	for it.Next() {
		de.d.Node(it.Name()).Validate()
	}
	require.NoError(de.d.t, it.Err())
}
