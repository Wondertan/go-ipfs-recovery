package reedsolomon

import (
	"fmt"

	"github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"
	"github.com/multiformats/go-varint"
	"github.com/templexxx/reedsolomon"
)

// TODO Handle case with identical shards
// TODO TESTS!!!!

type shard struct {
	ss *shards

	i    int
	nd   format.Node
	want bool
}

type shards struct {
	rs        *reedsolomon.RS
	id        cid.Cid
	ids       []cid.Cid
	m         map[cid.Cid]*shard
	vects     [][]byte
	hvs, wnts []int
	lln       int
}

func newShards(rnd *Node) (*shards, error) {
	lln, rln, s := len(rnd.Links()), len(rnd.RecoveryLinks()), rnd.RecoveryLinks()[0].Size
	ln := lln + rln

	rs, err := reedsolomon.New(lln, rln)
	if err != nil {
		return nil, err
	}

	ss := &shards{
		rs:    rs,
		ids:   make([]cid.Cid, ln),
		id:    rnd.Cid(),
		m:     make(map[cid.Cid]*shard, ln),
		vects: make([][]byte, ln),
		hvs:   make([]int, 0, lln),
		wnts:  make([]int, 0, ln),
		lln:   lln,
	}

	for i, l := range rnd.Links() {
		ss.ids[i] = l.Cid
		ss.vects[i] = make([]byte, s)
		ss.m[l.Cid] = &shard{
			ss: ss,
			i:  i,
		}
	}

	for i, l := range rnd.RecoveryLinks() {
		i += lln
		ss.ids[i] = l.Cid
		ss.vects[i] = make([]byte, s)
		ss.m[l.Cid] = &shard{
			ss: ss,
			i:  i,
		}
	}

	// TODO Shuffle ids ???
	return ss, nil
}

func (ss *shards) Recoverable() bool {
	return len(ss.hvs) >= ss.lln
}

func (ss *shards) Parent() cid.Cid {
	return ss.id
}

func (ss *shards) IDs() []cid.Cid {
	return ss.ids
}

func (ss *shards) Want(id cid.Cid) error {
	_, err := ss.shard(id)
	return err
}

func (ss *shards) WantData() {
	for i, id := range ss.ids[:ss.lln] {
		ss.m[id].want = true
		ss.wnts = append(ss.wnts, i)
	}
}

func (ss *shards) WantAll() {
	for _, sh := range ss.m {
		sh.want = true
		ss.wnts = append(ss.wnts, sh.i)
	}
}

func (ss *shards) Wanted() (nds []format.Node, err error) {
	nds = make([]format.Node, len(ss.wnts))
	for i, j := range ss.wnts {
		nds[i], err = ss.Get(ss.ids[j])
		if err != nil {
			return nil, err
		}
	}

	return
}

func (ss *shards) Fill(nd format.Node) bool {
	sh, ok := ss.m[nd.Cid()]
	if !ok || sh.nd != nil {
		return !ss.Recoverable()
	}

	if sh.i < ss.lln {
		n := varint.PutUvarint(ss.vects[sh.i], uint64(len(nd.RawData())))
		copy(ss.vects[sh.i][n:], nd.RawData())
	} else {
		copy(ss.vects[sh.i], nd.RawData())
	}

	sh.fill(nd)
	return !ss.Recoverable()
}

func (ss *shards) Get(id cid.Cid) (format.Node, error) {
	sh, err := ss.shard(id)
	if err != nil {
		return nil, err
	}
	if sh.nd != nil {
		return sh.nd, nil
	}
	if !ss.Recoverable() {
		return nil, fmt.Errorf("reedsolomon: not enough recoverability for available Nodes")
	}

	err = ss.rs.Reconst(ss.vects, ss.hvs, ss.wnts)
	if err != nil {
		return nil, err
	}

	vec := ss.vects[sh.i]
	if sh.i < ss.lln {
		s, n, err := varint.FromUvarint(vec)
		if err != nil {
			return nil, err
		}

		vec = vec[n : int(s)+n]
	}

	b, _ := blocks.NewBlockWithCid(vec, id)
	nd, err := format.Decode(b)
	if err != nil {
		return nil, err
	}

	sh.fill(nd)
	return nd, nil
}

func (ss *shards) shard(id cid.Cid) (*shard, error) {
	sh, ok := ss.m[id]
	if !ok {
		return nil, fmt.Errorf("reedsolomon: wrong child")
	}

	if !sh.want {
		sh.want = true
		ss.wnts = append(ss.wnts, sh.i)
	}

	return sh, nil
}

func (sh *shard) fill(nd format.Node) {
	sh.nd = nd
	sh.ss.hvs = append(sh.ss.hvs, sh.i)

	for i, j := range sh.ss.wnts {
		if j == sh.i {
			sh.ss.wnts[i] = sh.ss.wnts[len(sh.ss.wnts)-1]
			sh.ss.wnts = sh.ss.wnts[:len(sh.ss.wnts)-1]
		}
	}
}
