package recovery

import (
	"context"
	"sync"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
)

// FetchRedundant defines whenever redundant nodes have to be always fetched or only on restoration.
// TODO This better be new IPFS config field.
var FetchRedundant = false

// getter implements NodeGetter capable for restoring missing DAG nodes using redundant nodes.
// It is important for restoration to traverse the whole DAG using the only one getter, cause it is the only way to
// remember reverse links to parents, unless DAGService interface is changed.
type getter struct {
	get format.NodeGetter
	r   Recoverer

	parents map[cid.Cid]Node
	l       sync.Mutex
}

func NewNodeGetter(get format.NodeGetter, r Recoverer) format.NodeGetter {
	return &getter{
		r:       r,
		get:     get,
		parents: make(map[cid.Cid]Node),
	}
}

// Get restores node, if possible.
func (g *getter) Get(ctx context.Context, id cid.Cid) (format.Node, error) {
	nd, err := g.get.Get(ctx, id)
	switch err {
	default:
		return nil, err
	case format.ErrNotFound:
		prnt := g.getParent(id)
		if prnt == nil {
			return nil, err
		}

		nds, err := g.r.Recover(ctx, prnt, id)
		if err != nil {
			log.Infof("Restoration attempt failed with: %s", err)
			return nil, format.ErrNotFound
		}

		nd = nds[0]
	case nil:
	}

	return g.maybeParent(nd), nil
}

// GetMany restores missing nodes on the fly, if possible.
// All given keys have to be related to one parent. // TODO Remove this requirement
func (g *getter) GetMany(ctx context.Context, cids []cid.Cid) <-chan *format.NodeOption {
	ch := g.get.GetMany(ctx, cids)
	out := make(chan *format.NodeOption, len(cids))

	// track remaining nodes
	lost := cid.NewSet()
	for _, id := range cids {
		lost.Add(id)
	}

	go func() {
		defer close(out)
		for {
			select {
			case no, ok := <-ch:
				if ok && no.Err == nil {
					no.Node = g.maybeParent(no.Node)

					select {
					case out <- no:
						// untrack the succeed node and proceed looping if others still remaining.
						lost.Remove(no.Node.Cid())
						if lost.Len() != 0 {
							continue
						}
					case <-ctx.Done():
					}

					return
				}

				// it is assumed that all the passed keys have a same parent, so it can be found by any key.
				lost := lost.Keys()
				prnt := g.getParent(lost[0])
				if prnt == nil {
					// if parent not found, we have nothing to do.
					return
				}

				nds, err := g.r.Recover(ctx, prnt, lost...)
				if err != nil {
					// restoration is a DAG implementation detail, so no need to pass the error up, just log the failed attempt.
					log.Infof("Restoration attempt failed with: %s", err)
					return
				}

				// finally send restored nodes.
				for _, nd := range nds {
					nd = g.maybeParent(nd) // restored nodes can also be parents.

					select {
					case out <- &format.NodeOption{Node: nd}:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

// maybeParent caches the given node if it's a parent for possible future restorations.
func (g *getter) maybeParent(nd format.Node) format.Node {
	rn, ok := nd.(Node)
	if !ok {
		return nd
	}

	cp := rn.Copy().(Node) // it is better to make a copy here, since node can be altered by the caller.

	g.l.Lock()
	g.parents[rn.Cid()] = cp
	g.l.Unlock()

	if FetchRedundant {
		go g.fetchRedundant(cp)
	}

	return rn.Proto()
}

// fetchRedundant triggers fetching of redundant nodes linked to parent.
func (g *getter) fetchRedundant(nd Node) {
	ids := make([]cid.Cid, len(nd.RecoveryLinks()))
	for i, l := range nd.RecoveryLinks() {
		ids[i] = l.Cid
	}

	// FIXME Unluckily, if redundant nodes are stored locally already, getting them here just loads the memory for
	//  nothing and there is no workaround for that unless blockstore is added.
	for no := range g.get.GetMany(context.Background(), ids) {
		if no.Err != nil {
			log.Errorf("Failed to load redundant nodes: %s", no.Err)
		}
	}
}

// getParent tries to find the parent gotten within the getter for the given CID.
func (g *getter) getParent(id cid.Cid) Node {
	g.l.Lock()
	defer g.l.Unlock()

	for p, n := range g.parents {
		for _, l := range n.Links() {
			if l.Cid.Equals(id) {
				// restoration reconstructs all linked nodes at once hence parent can't be reused, so uncache it.
				delete(g.parents, p)
				return n
			}
		}
	}

	return nil
}
