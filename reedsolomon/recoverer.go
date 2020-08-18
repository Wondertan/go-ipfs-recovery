package reedsolomon

import (
	"context"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"

	"github.com/Wondertan/go-ipfs-recovery"
)

type recoverer struct {
	ctx context.Context
	dag format.DAGService

	recs map[cid.Cid]*recoverySes
	rl   sync.RWMutex

	strg recovery.Strategy
}

// NewRecoverer creates new Reed-Solomon Recoverer.
// Strategy have to be an option,
func NewRecoverer(ctx context.Context, dag format.DAGService, strg recovery.Strategy) recovery.Recoverer {
	return &recoverer{
		ctx:  ctx,
		dag:  dag,
		recs: make(map[cid.Cid]*recoverySes),
		strg: strg,
	}
}

func (r *recoverer) Recover(ctx context.Context, nd recovery.Node, ids ...cid.Cid) (_ <-chan *format.NodeOption, err error) {
	rnd, ok := nd.(*Node)
	if !ok {
		return nil, fmt.Errorf("reedsolomon: wrong Node type")
	}

	r.rl.RLock()
	rc, ok := r.recs[rnd.Cid()]
	r.rl.RUnlock()

	if !ok {
		rc, err = r.newRecovery(ctx, rnd)
		if err != nil {
			return nil, err
		}

		r.rl.Lock()
		r.recs[rnd.Cid()] = rc
		r.rl.Unlock()
	}

	return rc.recover(ctx, ids), nil
}

type recoverySes struct {
	r *recoverer

	ctx     context.Context
	getCncl context.CancelFunc

	reqCh chan *rcvrReq
	reqs  []*rcvrReq

	in <-chan *format.NodeOption
	sh *shards
}

func (r *recoverer) newRecovery(ctx context.Context, rnd *Node) (*recoverySes, error) {
	sh, err := newShards(rnd)
	if err != nil {
		return nil, err
	}

	if r.strg.All() {
		sh.WantAll()
	} else if r.strg.Data() {
		sh.WantData()
	}

	getCtx, getCncl := context.WithCancel(r.ctx)
	in := r.dag.GetMany(getCtx, sh.IDs())

	rc := &recoverySes{
		r:       r,
		ctx:     ctx,
		getCncl: getCncl,
		reqCh:   make(chan *rcvrReq, 1),
		reqs:    make([]*rcvrReq, 0, 1),
		in:      in,
		sh:      sh,
	}
	go rc.handle()
	return rc, nil
}

type rcvrReq struct {
	ctx context.Context
	ids []cid.Cid
	out chan *format.NodeOption
}

func (r *recoverySes) recover(ctx context.Context, ids []cid.Cid) <-chan *format.NodeOption {
	out := make(chan *format.NodeOption, len(ids))
	select {
	case r.reqCh <- &rcvrReq{ctx: ctx, ids: ids, out: out}:
	case <-r.ctx.Done():
		return nil
	case <-ctx.Done():
		return nil
	}
	return out
}

func (r *recoverySes) handle() {
	defer func() {
		r.getCncl()
		r.respond()

		r.r.rl.Lock()
		delete(r.r.recs, r.sh.Parent())
		r.r.rl.Unlock()

		nds, err := r.sh.Wanted()
		if err != nil {
			log.Error(err)
		}

		err = r.r.dag.AddMany(r.ctx, nds)
		if err != nil {
			log.Error(err)
		}
	}()

	for {
		select {
		case nd := <-r.in:
			if nd.Err != nil {
				log.Error(nd.Err)
				continue
			}

			if !r.sh.Fill(nd.Node) {
				return
			}
		case req := <-r.reqCh:
			r.reqs = append(r.reqs, req)
			for _, id := range req.ids {
				err := r.sh.Want(id)
				if err != nil {
					log.Error(err)
				}
			}
		case <-r.ctx.Done():
			if len(r.reqs) <= 1 {
				return
			}

			r.reqs = r.reqs[1:]
			r.ctx = r.reqs[0].ctx
		case <-r.r.ctx.Done():
			return
		}
	}
}

// TODO Be more reactive and respond on Fill if possible
func (r *recoverySes) respond() {
	for _, req := range r.reqs {
		for _, id := range req.ids {
			nd, err := r.sh.Get(id)
			select {
			case req.out <- &format.NodeOption{Node: nd, Err: err}:
				continue
			case <-req.ctx.Done():
			}

			break
		}
	}
}
