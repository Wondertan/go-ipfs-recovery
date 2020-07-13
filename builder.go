package restore

import (
	"context"
	"errors"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-ipld-format"
	fs "github.com/ipfs/go-unixfs"
	"github.com/ipfs/go-unixfs/importer/helpers"
)

// Modified version of UnixFS's balanced DAG builder.
func Layout(db *helpers.DagBuilderHelper, r Restorer) (format.Node, error) {
	if db.Done() {
		root, err := db.NewLeafNode(nil, fs.TFile)
		if err != nil {
			return nil, err
		}

		return root, db.Add(root)
	}

	root, fileSize, err := db.NewLeafDataNode(fs.TFile)
	if err != nil {
		return nil, err
	}

	for depth := 1; !db.Done(); depth++ {
		newRoot := db.NewFSNodeOverDag(fs.TFile)
		newRoot.AddChild(root, fileSize, db)

		root, fileSize, err = fillNodeRec(db, r, newRoot, depth)
		if err != nil {
			return nil, err
		}
	}

	return root, db.Add(root)
}

func fillNodeRec(db *helpers.DagBuilderHelper, r Restorer, node *helpers.FSNodeOverDag, depth int) (filledNode format.Node, nodeFileSize uint64, err error) {
	if depth < 1 {
		return nil, 0, errors.New("attempt to fillNode at depth < 1")
	}

	if node == nil {
		node = db.NewFSNodeOverDag(fs.TFile)
	}

	var childNode format.Node
	var childFileSize uint64
	nodes := make([]blocks.Block, db.Maxlinks())

	for i := 0; node.NumChildren() < db.Maxlinks() && !db.Done(); i++ {
		if depth == 1 {
			childNode, childFileSize, err = db.NewLeafDataNode(fs.TFile)
		} else {
			childNode, childFileSize, err = fillNodeRec(db, r, nil, depth-1)
		}
		if err != nil {
			return nil, 0, err
		}

		err = node.AddChild(childNode, childFileSize, db)
		if err != nil {
			return nil, 0, err
		}

		nodes[i] = childNode
	}

	nodeFileSize = node.FileSize()
	filledNode, err = node.Commit()
	if err != nil {
		return nil, 0, err
	}

	if node.NumChildren() == db.Maxlinks() {
		filledNode, err = r.Encode(context.TODO(), filledNode)
		if err != nil {
			return nil, 0, err
		}
	}

	return filledNode, nodeFileSize, nil
}
