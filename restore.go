package restore

import (
	"context"
	"errors"

	format "github.com/ipfs/go-ipld-format"
)

var ErrSizesNotEqual = errors.New("restore: node's links must have equal size")

type Restorer interface {
	Encode(context.Context, format.Node) (format.Node, error)
}