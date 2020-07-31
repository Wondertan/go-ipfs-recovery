package restore

import (
	"context"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
)

type getter struct {
	r   Restorer
	get format.NodeGetter

	parents map[cid.Cid]*Node

	loadRedundant bool
}

func (g *getter) Get(ctx context.Context, cid cid.Cid) (format.Node, error) {
	nd, err := g.get.Get(ctx, cid)
	if err != nil {
		return nil, err
	}

	rn, ok := nd.(*Node)
	if ok {

	}

	return nd, err
}

func (g *getter) GetMany(ctx context.Context, cids []cid.Cid) <-chan *format.NodeOption {
	ch := g.get.GetMany(ctx, cids)
	out := make(chan *format.NodeOption, len(cids))

	left := cid.NewSet()
	for _, id := range cids {
		left.Add(id)
	}

	go func() {
		defer close(out)

		for {
			select {
			case no, ok := <-ch:
				if ok && no.Err == nil {
					select {
					case out <- no:
						left.Remove(no.Node.Cid())
						if left.Len() != 0 {
							continue
						}
					case <-ctx.Done():
					}

					return
				}

				ids := left.Keys()
				for _, id := range ids {

				}

				g.r.Restore(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

func (g *getter) getCommonParent(ids []cid.Cid) (cid.Cid, []cid.Cid) {
	for p, n := range g.parents {
		for _, l := range n.Links() {
			for _, id := range ids {
				if l.Cid.Equals(id) {
					n.RemoveNodeLink(l.Name) // TODO Remove by position
					if len(n.Links()) == 0 {
						delete(g.parents, p)
					}
				}
			}
		}
	}

	return cid.Undef, nil, nil
}
