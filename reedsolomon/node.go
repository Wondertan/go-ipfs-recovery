package reedsolomon

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"

	"github.com/Wondertan/go-ipfs-recovery"
	cpb "github.com/Wondertan/go-ipfs-recovery/reedsolomon/pb"
)

// TODO Do not save sizes of all links independently as they are always the same.

// Node is a recovery Node based ob Reed-Solomon coding.
type Node struct {
	*merkledag.ProtoNode

	recovery []*format.Link
	cache    []byte
	cid      cid.Cid
}

func NewNode(proto *merkledag.ProtoNode) *Node {
	nd := &Node{ProtoNode: proto.Copy().(*merkledag.ProtoNode)}
	nd.SetCidBuilder(nd.CidBuilder().WithCodec(Codec))
	return nd
}

func (n *Node) Recoverability() recovery.Recoverability {
	return len(n.recovery)
}

func (n *Node) RecoveryLinks() []*format.Link {
	return n.recovery
}

func (n *Node) AddRedundantNode(nd format.Node) {
	if nd == nil {
		return
	}

	n.cache = nil
	n.recovery = append(n.recovery, &format.Link{
		Name: strconv.Itoa(len(n.recovery)),
		Size: uint64(len(nd.RawData())),
		Cid:  nd.Cid(),
	})
}

func (n *Node) RemoveRedundantNode(id cid.Cid) {
	if !id.Defined() {
		return
	}

	ref := n.recovery[:0]
	for _, v := range n.recovery {
		if v.Cid.Equals(id) {
			n.cache = nil
		} else {
			ref = append(ref, v)
		}
	}
	n.recovery = ref
}

func (n *Node) RawData() []byte {
	if n.cache != nil {
		return n.cache
	}

	var err error
	n.cache, err = MarshalNode(n)
	if err != nil {
		panic(fmt.Sprintf("can't marshal Node: %s", err))
	}

	return n.cache
}

func (n *Node) Cid() cid.Cid {
	if n.cache != nil && n.cid.Defined() {
		return n.cid
	}

	var err error
	n.cid, err = n.CidBuilder().Sum(n.RawData())
	if err != nil {
		panic(fmt.Sprintf("can't form CID: %s", err))
	}

	return n.cid
}

func (n *Node) String() string {
	return n.Cid().String()
}

func (n *Node) Copy() format.Node {
	nd := new(Node)
	nd.ProtoNode = n.ProtoNode.Copy().(*merkledag.ProtoNode)
	l := len(n.recovery)
	if l > 0 {
		nd.recovery = make([]*format.Link, l)
		for i, r := range n.recovery {
			nd.recovery[i] = &format.Link{
				Name: r.Name,
				Size: r.Size,
				Cid:  r.Cid,
			}
		}
	}

	return nd
}

func (n *Node) Stat() (*format.NodeStat, error) {
	l := len(n.RawData())
	cumSize, err := n.Size()
	if err != nil {
		return nil, err
	}

	return &format.NodeStat{
		Hash:           n.Cid().String(),
		NumLinks:       len(n.Links()),
		BlockSize:      l,
		DataSize:       len(n.Data()),
		CumulativeSize: int(cumSize),
	}, nil
}

func (n *Node) Size() (uint64, error) {
	s := uint64(len(n.RawData()))
	for _, l := range n.Links() {
		s += l.Size
	}
	for _, l := range n.recovery {
		s += l.Size
	}
	return s, nil
}

func MarshalNode(n *Node) ([]byte, error) {
	var err error
	pb := &cpb.PBNode{}
	pb.Proto, err = n.ProtoNode.Marshal()
	if err != nil {
		return nil, err
	}

	l := len(n.recovery)
	if l > 0 {
		sort.Stable(merkledag.LinkSlice(n.recovery))
		pb.Recovery = make([]*cpb.PBLink, l)
		for i, r := range n.recovery {
			pb.Recovery[i] = &cpb.PBLink{
				Name:  r.Name,
				Size_: r.Size,
				Hash:  r.Cid.Bytes(),
			}
		}
	}

	return pb.Marshal()
}

func UnmarshalNode(data []byte) (*Node, error) {
	nd := &Node{}
	pb := &cpb.PBNode{}
	err := pb.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	nd.ProtoNode, err = merkledag.DecodeProtobuf(pb.Proto)
	if err != nil {
		return nil, err
	}

	l := len(pb.Recovery)
	if l > 0 {
		nd.recovery = make([]*format.Link, l)
		for i, r := range pb.Recovery {
			nd.recovery[i] = &format.Link{
				Name: r.Name,
				Size: r.Size_,
			}

			nd.recovery[i].Cid, err = cid.Cast(r.Hash)
			if err != nil {
				return nil, err
			}
		}
		sort.Stable(merkledag.LinkSlice(nd.recovery))
	}

	return nd, nil
}

func DecodeNode(b blocks.Block) (format.Node, error) {
	id := b.Cid()
	if id.Prefix().Codec != Codec {
		return nil, fmt.Errorf("can only decode restorable node")
	}

	nd, err := UnmarshalNode(b.RawData())
	if err != nil {
		return nil, err
	}

	nd.cid = b.Cid()
	nd.SetCidBuilder(b.Cid().Prefix())
	return nd, nil
}

// Shadowed methods to reset caching.
//
func (n *Node) AddRawLink(name string, l *format.Link) error {
	n.cache = nil
	return n.ProtoNode.AddRawLink(name, l)
}

func (n *Node) AddNodeLink(name string, that format.Node) error {
	n.cache = nil
	return n.ProtoNode.AddNodeLink(name, that)
}

func (n *Node) RemoveNodeLink(name string) error {
	n.cache = nil
	return n.ProtoNode.RemoveNodeLink(name)
}

func (n *Node) SetData(d []byte) {
	n.cache = nil
	n.ProtoNode.SetData(d)
}
