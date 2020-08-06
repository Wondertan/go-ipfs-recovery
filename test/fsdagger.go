package test

import (
	"context"
	"sync"
	"testing"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-unixfs/io"
	testu "github.com/ipfs/go-unixfs/test"
	"github.com/stretchr/testify/require"
)

type Morpher func(format.Node) (format.Node, error)

// FSDagger is a test helper useful to build various UnixFS DAGs.
type FSDagger struct {
	Morpher Morpher

	t   testing.TB
	ctx context.Context
	dag format.DAGService

	nodes map[string]*FSDaggerNode
	l     sync.Mutex
}

func NewFSDagger(t testing.TB, ctx context.Context, dag format.DAGService) *FSDagger {
	return &FSDagger{t: t, ctx: ctx, dag: dag, nodes: make(map[string]*FSDaggerNode)}
}

func (d *FSDagger) RandNode(name string) *FSDaggerNode {
	data, nd := testu.GetRandomNode(d.t, d.dag, 10000, testu.UseCidV1)
	return d.addNode(&FSDaggerNode{d: d, node: nd, name: name, Data: data})
}

func (d *FSDagger) NewNode(name string, data []byte) *FSDaggerNode {
	return d.addNode(&FSDaggerNode{
		d:    d,
		node: testu.GetNode(d.t, d.dag, data, testu.NodeOpts{Prefix: merkledag.V1CidPrefix()}),
		name: name,
		Data: data,
	})
}

func (d *FSDagger) NewDir(name string, es ...*FSDaggerNode) *FSDaggerNode {
	dir := io.NewDirectory(d.dag)
	dir.SetCidBuilder(merkledag.V1CidPrefix())

	for _, e := range es {
		err := dir.AddChild(d.ctx, e.name, e.node)
		require.NoError(d.t, err)
	}

	nd, err := dir.GetNode()
	require.NoError(d.t, err)

	err = d.dag.Add(d.ctx, nd)
	require.NoError(d.t, err)

	return d.addNode(&FSDaggerNode{d: d, node: nd, name: name, isDir: true})
}

func (d *FSDagger) Node(name string) *FSDaggerNode {
	d.l.Lock()
	defer d.l.Unlock()

	e, ok := d.nodes[name]
	if !ok {
		d.t.Fatal("dagger: entry not found")
	}

	return e
}

func (d *FSDagger) Remove(name string) {
	e := d.Node(name)
	err := d.dag.Remove(d.ctx, e.node.Cid())
	require.NoError(d.t, err)
}

func (d *FSDagger) addNode(e *FSDaggerNode) *FSDaggerNode {
	d.l.Lock()
	defer d.l.Unlock()

	_, ok := d.nodes[e.name]
	if ok {
		d.t.Fatal("dagger: entry name is used")
	}

	if d.Morpher != nil {
		m, err := d.Morpher(e.node)
		if err != nil {
			d.t.Fatalf("dagger: morpher failed with: %s", err)
		}

		e.node = m
	}

	d.nodes[e.name] = e
	return e
}
