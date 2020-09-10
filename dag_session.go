package recovery

import (
	"context"
	"sync"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	exchange "github.com/ipfs/go-ipfs-exchange-interface"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-verifcid"
)

// getter implements NodeGetter capable for restoring missing DAG nodes using redundant nodes.
// It is important for restoration to traverse the whole DAG using the only one getter, cause it is the only way to
// remember reverse links to prnts, unless DAGService interface is changed.
type dagSession struct {
	ctx context.Context

	r  Recoverer
	ex exchange.Interface
	bs blockstore.Blockstore

	prnts map[cid.Cid]Node
	pl    sync.Mutex

	f  exchange.Fetcher
	fo sync.Once
}

func NewDagSession(ctx context.Context, r Recoverer, ex exchange.Interface, bs blockstore.Blockstore) format.NodeGetter {
	return &dagSession{
		ctx:   ctx,
		r:     r,
		ex:    ex,
		bs:    bs,
		prnts: make(map[cid.Cid]Node),
	}
}

func (ds *dagSession) Get(ctx context.Context, id cid.Cid) (format.Node, error) {
	err := verifcid.ValidateCid(id)
	if err != nil {
		return nil, err
	}

	// 1. Try to get from the Blockstore.
	b, err := ds.bs.Get(id)
	switch err {
	default:
		return nil, err
	case nil:
		return ds.decode(b)
	case blockstore.ErrNotFound:
	}

	// 2. Try to recover.
	nd, err := ds.recover(ctx, id)
	if err != format.ErrNotFound {
		return nd, err
	}

	// 3. Try to get from the network.
	b, err = ds.fetcher().GetBlock(ctx, id)
	switch err {
	default:
		return nil, err
	case nil:
		return ds.decode(b)
	case blockstore.ErrNotFound:
	}

	// 4. Fail :(
	return nil, format.ErrNotFound
}

func (ds *dagSession) GetMany(ctx context.Context, in []cid.Cid) <-chan *format.NodeOption {
	out := make(chan *format.NodeOption, len(in))
	ids := make([]cid.Cid, len(in))
	copy(ids, in)

	go func() {
		defer close(out)

		i := 0
		for _, id := range ids {
			err := verifcid.ValidateCid(id)
			if err != nil {
				continue
			}

			b, err := ds.bs.Get(id)
			if err == nil {
				nd, err := ds.decode(b)
				select {
				case out <- &format.NodeOption{Node: nd, Err: err}:
					continue
				case <-ctx.Done():
					return
				}
			}

			// TODO: Recovery might be long blocking, so run it async
			nd, err := ds.recover(ctx, id)
			if err == nil {
				select {
				case out <- &format.NodeOption{Node: nd}:
					continue
				case <-ctx.Done():
					return
				}
			}

			ids[i] = id
			i++
		}

		ids = ids[:i]
		if len(ids) == 0 {
			return
		}

		bs, err := ds.fetcher().GetBlocks(ctx, ids)
		if err != nil {
			return
		}

		for b := range bs {
			nd, err := ds.decode(b)
			select {
			case out <- &format.NodeOption{Node: nd, Err: err}:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

func (ds *dagSession) recover(ctx context.Context, id cid.Cid) (format.Node, error) {
	prnt := ds.getParentFor(id)
	if prnt == nil {
		return nil, format.ErrNotFound
	}

	nds, err := ds.r.Recover(ctx, prnt, id)
	if err != nil {
		log.Warnf("Recovery attempt failed(%s): %s", id, err)
		return nil, format.ErrNotFound
	}

	select {
	case no := <-nds:
		if no.Err != nil {
			log.Warnf("Recovery attempt failed(%s): %s", id, err)
			return nil, format.ErrNotFound
		}

		log.Infof("Successful recovery(%s)", id)
		return ds.decode(no.Node)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (ds *dagSession) fetcher() exchange.Fetcher {
	ds.fo.Do(func() {
		ex, ok := ds.ex.(exchange.SessionExchange)
		if !ok {
			ds.f = ds.ex
			return
		}

		ds.f = ex.NewSession(ds.ctx)
	})

	return ds.f
}

// decode transforms block into the Node and caches it, if it's a recovery one.
func (ds *dagSession) decode(b blocks.Block) (format.Node, error) {
	nd, err := format.Decode(b)
	if err != nil {
		return nil, err
	}

	rn, ok := nd.(Node)
	if !ok {
		return nd, nil
	}

	ds.pl.Lock()
	ds.prnts[rn.Cid()] = rn.Copy().(Node) // it is better to make a copy here, since node can be altered by the caller.
	ds.pl.Unlock()

	return rn.Proto(), nil
}

// getParentFor tries to find the parent gotten within the session for the given CID.
func (ds *dagSession) getParentFor(id cid.Cid) Node {
	ds.pl.Lock()
	defer ds.pl.Unlock()

	for _, n := range ds.prnts {
		for _, l := range n.Links() {
			if l.Cid.Equals(id) {
				// restoration reconstructs all linked nodes at fo hence parent can't be reused, so uncache it.
			//	delete(ds.prnts, p)
				return n
			}
		}
	}

	return nil
}
