package recovery

import (
	"context"
	"errors"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipfs/go-merkledag"
)

var log = logging.Logger("recovery")

var ErrRecoveryExceeded = errors.New("recovery: maximum recoverability exceeded")

// Recoverability param for the Node defines the max amount of nodes that can be recovered from it.
type Recoverability = int

// Node wraps IPLD Node with an ability to recover lost linked Nodes.
type Node interface {
	format.Node

	// Recoverability of the Node.
	Recoverability() Recoverability

	// RecoveryLinks lists links to all recovery Nodes.
	RecoveryLinks() []*format.Link

	// FIXME This is awful, but there is no workaround fot that. IPFS is very strict about using only ProtoNode
	//  in multiple cases.
	Proto() *merkledag.ProtoNode
}

type Recoverer interface {

	// Recovers Nodes by ids from the recovery Node.
	Recover(context.Context, Node, ...cid.Cid) ([]format.Node, error)
}

type Encoder interface {

	// Encodes Node to a recovery Node.
	// Bigger recoverability - higher storage usage.
	Encode(context.Context, format.Node, Recoverability) (Node, error)
}
