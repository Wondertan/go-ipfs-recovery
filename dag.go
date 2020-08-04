package recovery

import (
	"context"

	format "github.com/ipfs/go-ipld-format"
)

// EncodeDAG encodes whole DAG under the given node with given Encoder and recoverability.
func EncodeDAG(ctx context.Context, dag format.NodeGetter, e Encoder, nd format.Node, r Recoverability) (format.Node, error) {
	if len(nd.Links()) == 0 {
		return nd, nil
	}

	nd = nd.Copy()
	for _, l := range nd.Links() {
		nd, err := l.GetNode(ctx, dag)
		if err != nil {
			return nil, err
		}

		end, err := EncodeDAG(ctx, dag, e, nd, r)
		if err != nil {
			return nil, err
		}

		if !nd.Cid().Equals(end.Cid()) {
			l.Size, err = end.Size()
			if err != nil {
				return nil, err
			}

			l.Cid = end.Cid()
		}
	}

	return e.Encode(ctx, nd, r)
}
