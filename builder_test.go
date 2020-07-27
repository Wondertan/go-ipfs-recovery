package restore

import (
	"context"
	"bytes"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dstest "github.com/ipfs/go-merkledag/test"
	chunker "github.com/ipfs/go-ipfs-chunker"
	unixfs "github.com/ipfs/go-unixfs"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/ipfs/go-ipfs-util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestDag(ds format.DAGService, spl chunker.Splitter, r Restorer) (format.Node, error) {
	dbp := helpers.DagBuilderParams{
		Dagserv:  ds,
		Maxlinks: helpers.DefaultLinksPerBlock,
	}

	db, err := dbp.New(spl)
	if err != nil {
		return nil, err
	}

	nd, err := Layout(db, r)
	if err != nil {
		return nil, err
	}
	return nd, nil
}

func getTestDag(ds format.DAGService, size int64, blksize int64, r Restorer) (format.Node, []byte) {
	data := make([]byte, size)
	util.NewTimeSeededRand().Read(data)
	reader := bytes.NewReader(data)

	nd, _ := buildTestDag(ds, chunker.NewSizeSplitter(reader, blksize), r)

	return nd, data
}

func NewRestorer() (Restorer, format.DAGService) {
	ds := dstest.Mock()
	var r Restorer
	return r, ds
}

func TestSingleChunk(t *testing.T) {
	r, dag := NewRestorer()

	nd, expect := getTestDag(dag, 100, 200, r)
	pnd := nd.(*merkledag.ProtoNode)
	// extract data from data node
	fsNode, err := unixfs.FSNodeFromBytes(pnd.Data())
	require.NoError(t, err)

	assert.Equal(t, fsNode.Data(), expect)
}

func TestTwoChunks(t *testing.T) {
	r, dag := NewRestorer()
	ctx := context.Background()
	root, expect := getTestDag(dag, 200, 100, r)
	// get data from the 2 children data nodes
	var actual []byte
	for _, l := range root.Links() {
		nd, err := l.GetNode(ctx, dag)
		require.NoError(t, err)

		pnd := nd.(*merkledag.ProtoNode)
		fsNode, err := unixfs.FSNodeFromBytes(pnd.Data())
		require.NoError(t, err)

		actual = append(actual, fsNode.Data()...)
	}

	assert.Equal(t, actual, expect)
}
