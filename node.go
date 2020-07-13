package restore

import (
	"fmt"
	"sort"
	"strconv"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"

	cpb "github.com/Wondertan/go-ipfs-restore/pb"
)

// DagProtobufRestorable is custom codec for the restorable node.
const DagProtobufRestorable = 0x700

func init() {
	// register global decoder
	format.Register(DagProtobufRestorable, DecodeNode)

	// register codec
	cid.Codecs["protobuf-correction"] = DagProtobufRestorable
	cid.CodecToStr[DagProtobufRestorable] = "protobuf-correction"
}

type Node struct {
	*merkledag.ProtoNode

	redundant []*format.Link
	cache     []byte
	cid       cid.Cid
}

func NewNode(proto *merkledag.ProtoNode) *Node {
	nd := &Node{ProtoNode: proto.Copy().(*merkledag.ProtoNode)}
	nd.SetCidBuilder(nd.CidBuilder().WithCodec(DagProtobufRestorable))
	return nd
}

func (n *Node) Redundant() []*format.Link {
	return n.redundant
}

func (n *Node) AddRedundantNode(nd format.Node) {
	if nd == nil {
		return
	}

	n.cache = nil
	n.redundant = append(n.redundant, &format.Link{
		Name: strconv.Itoa(len(n.redundant)),
		Size: uint64(len(nd.RawData())),
		Cid:  nd.Cid(),
	})
}

func (n *Node) RemoveRedundantNode(id cid.Cid) {
	if !id.Defined() {
		return
	}

	ref := n.redundant[:0]
	for _, v := range n.redundant {
		if v.Cid.Equals(id) {
			n.cache = nil
		} else {
			ref = append(ref, v)
		}
	}
	n.redundant = ref
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
	l := len(n.redundant)
	if l > 0 {
		nd.redundant = make([]*format.Link, l)
		for i, r := range n.redundant {
			nd.redundant[i] = &format.Link{
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
	for _, l := range n.redundant {
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

	l := len(n.redundant)
	if l > 0 {
		sort.Stable(merkledag.LinkSlice(n.redundant))
		pb.Redundant = make([]*cpb.PBLink, l)
		for i, r := range n.redundant {
			pb.Redundant[i] = &cpb.PBLink{
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

	l := len(pb.Redundant)
	if l > 0 {
		nd.redundant = make([]*format.Link, l)
		for i, r := range pb.Redundant {
			nd.redundant[i] = &format.Link{
				Name: r.Name,
				Size: r.Size_,
			}

			nd.redundant[i].Cid, err = cid.Cast(r.Hash)
			if err != nil {
				return nil, err
			}
		}
		sort.Stable(merkledag.LinkSlice(nd.redundant))
	}

	return nd, nil
}

func DecodeNode(b blocks.Block) (format.Node, error) {
	id := b.Cid()
	if id.Prefix().Codec != DagProtobufRestorable {
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
