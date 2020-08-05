package entanglement

import (
	"errors"
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

type RedundantNode struct {
	*merkledag.ProtoNode

	Position [2]int // from, to lattice pos
}

func NewRedundantNode(n *merkledag.ProtoNode, pos [2]int) *RedundantNode {
	nd := &RedundantNode{ProtoNode: n.Copy().(*merkledag.ProtoNode)}
	nd.SetCidBuilder(nd.CidBuilder().WithCodec(Codec))
	nd.Position = pos
	return nd
}

func (r *RedundantNode) AddRedundantNode(n *Node, prev *RedundantNode, strand int) error {
	data, _ := XORByteSlice(n.Data(), prev.Data())
	r.SetData(data)
	switch strand {
	case LH:
		r.Position = [2]int{n.Position, nextEntangleNode(n.Position, LH)}
	case RH:
		r.Position = [2]int{n.Position, nextEntangleNode(n.Position, RH)}
	case H:
		r.Position = [2]int{n.Position, nextEntangleNode(n.Position, H)}
	default:
		return errors.New("Strand must be LH=0,H=1,RH=2")
	}
	return nil
}

// Node is a recovery Node based on Entanglement coding.
type Node struct {
	*merkledag.ProtoNode

	Position int // position of node in lattice

	Inputs  map[int]*format.Link // left->0, horizontal->1, right->2
	Outputs map[int]*format.Link // left->0, horizontal->1, right->2

	cache []byte
	cid   cid.Cid
}

func NewNode(proto *merkledag.ProtoNode, pos int) *Node {
	nd := &Node{ProtoNode: proto.Copy().(*merkledag.ProtoNode)}
	nd.SetCidBuilder(nd.CidBuilder().WithCodec(Codec))
	nd.Position = pos
	return nd
}

func (n *Node) Recoverability() recovery.Recoverability {
	return alpha
}

func (n *Node) RecoveryLinks() []*format.Link {
	var rls []*format.Link

	for _, v := range n.Outputs {
		rls = append(rls, v)
	}

	return rls
}

// Don't use
func (n *Node) AddRedundantNode(nd format.Node) {
	if nd == nil {
		return
	}

	if len(n.Outputs) == 3 {
		return // error?
	}

	n.cache = nil
}

func (n *Node) GetInputs() (ins []*format.Link) {
	for _, v := range n.Inputs {
		ins = append(ins, v)
	}
	return
}

func (n *Node) GetOutputs() (ins []*format.Link) {
	for _, v := range n.Inputs {
		ins = append(ins, v)
	}
	return
}

func (n *Node) RemoveRedundantNode(id cid.Cid) {
	if !id.Defined() {
		return
	}

	for k, v := range n.Inputs {
		if v.Cid.Equals(id) {
			delete(n.Inputs, k)
			n.cache = nil
			return
		}
	}
	for k, v := range n.Outputs {
		if v.Cid.Equals(id) {
			delete(n.Outputs, k)
			n.cache = nil
			return
		}
	}
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
	return "CID: " + n.Cid().String() + "; Position: " + strconv.FormatInt(int64(n.Position), 10)
}

func (n *Node) Copy() format.Node {
	nd := new(Node)
	nd.ProtoNode = n.ProtoNode.Copy().(*merkledag.ProtoNode)
	lIn := len(n.Inputs)
	if lIn > 0 {
		nd.Inputs = make(map[int]*format.Link, lIn)
		for i := 0; i <= lIn; i++ {
			r := n.Inputs[i]
			nd.Inputs[i] = &format.Link{
				Name: r.Name,
				Size: r.Size,
				Cid:  r.Cid,
			}
		}
	}

	lOut := len(n.Outputs)
	if lOut > 0 {
		nd.Outputs = make(map[int]*format.Link, lOut)
		for i := 0; i <= lOut; i++ {
			r := n.Inputs[i]
			nd.Outputs[i] = &format.Link{
				Name: r.Name,
				Size: r.Size,
				Cid:  r.Cid,
			}
		}
	}

	nd.Position = n.Position

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
	for _, l := range n.GetInputs() {
		s += l.Size
	}
	for _, l := range n.GetOutputs() {
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

	l := len(n.Inputs) + len(n.Outputs)
	if l > 0 {
		recovery := append(n.GetInputs(), n.GetOutputs()...)
		sort.Stable(merkledag.LinkSlice(recovery))
		pb.Recovery = make([]*cpb.PBLink, l)
		for i, r := range recovery {
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

	// l := len(pb.Recovery)
	// if l > 0 {
	// nd.recovery = make([]*format.Link, l)
	// for i, r := range pb.Recovery {
	// nd.recovery[i] = &format.Link{
	// Name: r.Name,
	// Size: r.Size_,
	// }

	// nd.recovery[i].Cid, err = cid.Cast(r.Hash)
	// if err != nil {
	// return nil, err
	// }
	// }
	// sort.Stable(merkledag.LinkSlice(nd.recovery))
	// }

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
