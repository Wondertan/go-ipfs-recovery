package recovery

import (
	"context"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetter(t *testing.T) {
	ctx := context.Background()
	get := newFakeGetter()
	r := newFakeRecoverer()
	ng := NewNodeGetter(get, r)

	in := merkledag.NodeWithData([]byte("1234567890"))
	in2 := merkledag.NodeWithData([]byte("0987654321"))
	in3 := merkledag.NodeWithData([]byte("1234509876"))
	in.AddNodeLink("link", in2)
	in.AddNodeLink("link", in3)

	get.nodes[in.Cid()] = &fakeNode{in}
	r.nodes[in2.Cid()] = in2
	r.nodes[in3.Cid()] = in3

	out, err := ng.Get(ctx, in.Cid())
	require.NoError(t, err)
	assert.Equal(t, in.RawData(), out.RawData())

	out, err = ng.Get(ctx, in2.Cid())
	require.NoError(t, err)
	assert.Equal(t, in2.RawData(), out.RawData())

	out, err = ng.Get(ctx, in.Cid())
	require.NoError(t, err)
	assert.Equal(t, in.RawData(), out.RawData())

	och := ng.GetMany(ctx, []cid.Cid{in2.Cid(), in3.Cid()})

	out = (<-och).Node
	assert.Equal(t, in2.RawData(), out.RawData())

	out = (<-och).Node
	assert.Equal(t, in3.RawData(), out.RawData())
}

type fakeRecoverer struct {
	nodes map[cid.Cid]format.Node
}

func newFakeRecoverer() *fakeRecoverer {
	return &fakeRecoverer{
		nodes: make(map[cid.Cid]format.Node),
	}
}

func (f *fakeRecoverer) Recover(_ context.Context, nd Node, ids ...cid.Cid) ([]format.Node, error) {
	nds := make([]format.Node, len(ids))
outer:
	for i, id := range ids {
		for _, l := range nd.Links() {
			if l.Cid.Equals(id) {
				nd, ok := f.nodes[id]
				if !ok {
					break
				}

				nds[i] = nd
				continue outer
			}
		}

		return nil, format.ErrNotFound
	}

	return nds, nil
}

type fakeGetter struct {
	nodes map[cid.Cid]format.Node
}

func newFakeGetter() *fakeGetter {
	return &fakeGetter{
		nodes: make(map[cid.Cid]format.Node),
	}
}

func (f *fakeGetter) Get(_ context.Context, id cid.Cid) (format.Node, error) {
	nd, ok := f.nodes[id]
	if !ok {
		return nil, format.ErrNotFound
	}

	pnd, ok := nd.(*merkledag.ProtoNode)
	if !ok {
		return nd, nil
	}

	return &fakeNode{pnd}, nil
}

func (f *fakeGetter) GetMany(ctx context.Context, ids []cid.Cid) <-chan *format.NodeOption {
	out := make(chan *format.NodeOption, len(ids))
	go func() {
		for _, id := range ids {
			nd, err := f.Get(ctx, id)
			select {
			case out <- &format.NodeOption{Node: nd, Err: err}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

type fakeNode struct {
	*merkledag.ProtoNode
}

func (f *fakeNode) Copy() format.Node {
	return &fakeNode{f.ProtoNode.Copy().(*merkledag.ProtoNode)}
}

func (f *fakeNode) Recoverability() Recoverability {
	return 0
}

func (f *fakeNode) RecoveryLinks() []*format.Link {
	return nil
}
