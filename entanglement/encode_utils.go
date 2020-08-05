package entanglement

import (
	"fmt"
	"math"

	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
)

const (
	LH    = iota
	H     = iota
	RH    = iota
	alpha = 3
	s     = 5
	p     = 5
)

// Alpha = 3
func GetForwardNeighbours(i int) (r, h, l int) {
	// Check is it top, center or bottom in the lattice
	// 1 -> Top, 0 -> Bottom, else Center
	var nodePos = i % s

	if nodePos == 1 {
		r = i + s + 1
		h = i + s
		l = i + (s * p) - int(math.Pow(float64(s-1), 2))
	} else if nodePos == 0 {
		r = i + (s * p) - int(math.Pow(float64(s), 2)-1)
		h = i + s
		l = i + s - 1
	} else {
		r = i + s + 1
		h = i + s
		l = i + (s - 1)
	}
	return
}

// TODO: Fix underflow naming errors on the nodes on the extreme of the lattice.
func GetBackwardNeighbours(i int) (r, h, l int) {
	// Check is it top, center or bottom in the lattice
	// 1 -> Top, 0 -> Bottom, else Center
	var nodePos = i % s

	if nodePos == 1 {
		r = i - (s * p) + int((math.Pow(float64(s), 2) - 1))
		h = i - s
		l = i - (s - 1)
	} else if nodePos == 0 {
		r = i - (s + 1)
		h = i - s
		l = i - (s * p) + int(math.Pow(float64(s-1), 2))
	} else {
		r = i - (s + 1)
		h = i - s
		l = i - (s - 1)
	}
	return
}

func GetMemoryPosition(index int) (r, h, l int) {
	// Get the position in the ParityMemory array where the parity is located
	// For now this will recursively call the GetBackwardNeighbours function

	h = ((index - 1) % s) + s
	r, l = index, index

	for ; r > s; r, _, _ = GetBackwardNeighbours(r) {
	}

	switch r {
	case 1:
		r = 0
	case 2:
		r = 4
	case 3:
		r = 3
	case 4:
		r = 2
	case 5:
		r = 1
	}

	for ; l > s; _, _, l = GetBackwardNeighbours(l) {
	}

	switch l {
	case 1:
		l = 11
	case 2:
		l = 12
	case 3:
		l = 13
	case 4:
		l = 14
	case 5:
		l = 10
	}

	return
}

// NextEntangleNode specifies the rules for creating a helical structure
// with indexed nodes arranged by s, p
func nextEntangleNode(i, strand int) int {
	nodePos := i % s

	switch strand {
	case LH:
		switch nodePos {
		case -4: // special case for (3,5,5); Go doesn't like negative modulo
			fallthrough
		case 1:
			return i + s*p - (s-1)*(s-1)
		case 0:
			return i + s - 1
		default:
			return i + s - 1
		}
	case H:
		switch nodePos {
		case -4:
			fallthrough
		case 1:
			return i + s
		case 0:
			return i + s
		default:
			return i + s

		}
	case RH:
		switch nodePos {
		case -4:
			fallthrough
		case 1:
			return i + s + 1
		case 0:
			return i + s*p - (s*s - 1)
		default:
			return i + s + 1
		}
	default:
		// this shouldn't be hit
		return 0
	}
}

// ValidateNode checks whenever the given IPLD Node can be applied with Entanglement coding.
func ValidateNode(nd format.Node) error {
	_, ok := nd.(*merkledag.ProtoNode)
	if !ok {
		return fmt.Errorf("reedsolomon: node must be proto")
	}

	ls := nd.Links()
	// can relax need for links
	if len(ls) != 0 {

		size := ls[0].Size
		for _, l := range ls[1:] {
			if l.Size != size {
				return fmt.Errorf("reedsolomon: node's links must have equal size")
			}
		}
	}

	return nil
}

// XORByteSlice returns an XOR slice of 2 input slices
func XORByteSlice(a []byte, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("length of byte slices is not equivalent: %d != %d", len(a), len(b))
	}

	buf := make([]byte, len(a))

	for i, _ := range a {
		buf[i] = a[i] ^ b[i]
	}

	return buf, nil
}
