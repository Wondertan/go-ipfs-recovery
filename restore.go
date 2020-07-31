package restore

import (
	"context"
	"errors"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
)

var (
	ErrNotRestorable = errors.New("restore: node must be restorable")
	ErrSizesNotEqual = errors.New("restore: node's links must have equal size")
)

type Restorer interface {
	Restore(context.Context, cid.Cid, ...cid.Cid) ([]format.Node, error)
	Encode(context.Context, cid.Cid) (cid.Cid, error)
}
