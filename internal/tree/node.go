package tree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"iter"

	"github.com/gkits/pavosql/internal/page"
)

const (
	nOff    = 1
	wCurOff = nOff + 2
	dataOff = wCurOff + 2
)

var (
	ErrIndexOutOfBounds = errors.New("index is out of bounds")
)

/*
A node is an array of bytes representing the data stored in a single node of a B+Tree.
Nodes are stored in a custom byte format that is structured as follows:

	Description | Header | Data area
	------------+--------+----------
	Size in B   | 5      | ?

The Header contains specific meta data about the current state of the node and is structured as
follows:

	Description | Type | N | W-Cursor
	------------+------+---+-------------
	Size in B   | 1    | 2 | 2

N is the number of cells currently stored in the node and W-Cursor the write cursor represented
by an uint16 offset to the next position data can be written to.

The Data area is dynamically sized an contains the actual data in form of cells as well as the
offset references to those cells. Those two parts are separated by an empty space called the
void. The Data area is structured as follows:

	Description | Offsets | Void                      | Cells
	------------+---------+---------------------------+------
	Size in B   | N * 2   | W-Cursor - (N*2 + Header) | ?

The offsets are just an ordered list of uint16 pointing to the position of cell they are
referencing inside this node. The void is the empty space between the end of the offset list and
the W-Cursor which always points to the beginning of the cells.

The data stored in the cells is formatted as follows:

	Description | KeyLen | ValLen | Key    | Val
	------------+--------+--------+--------+-------
	Size in B   | 2      | 2      | KeyLen | ValLen
*/
type node [page.Size]byte

func newNode(typ page.Type) node {
	var n node
	n[0] = byte(typ)
	n.setWCursor(page.Size)
	return n
}

// Returns the type of n.
func (n *node) Type() page.Type {
	return page.Type(n[0])
}

// Returns the number of cells currently stored on n.
func (n *node) N() uint16 {
	return binary.LittleEndian.Uint16(n[nOff:])
}

// Returns the key of the i'th cell stored in n.
//
// Panics if i is greater or equal than the length of n.
func (n *node) Key(i uint16) []byte {
	if !n.indexInBounds(i) {
		panic(ErrIndexOutOfBounds)
	}
	off := n.offset(i)
	kLen := binary.LittleEndian.Uint16(n[off:])
	return n[off+4 : off+4+kLen]
}

// Returns the value of the i'th cell stored in n.
//
// Panics if i is greater or equal than the length of n.
func (n *node) Val(i uint16) []byte {
	if !n.indexInBounds(i) {
		panic(ErrIndexOutOfBounds)
	}
	off := n.offset(i)
	kLen := binary.LittleEndian.Uint16(n[off:])
	vLen := binary.LittleEndian.Uint16(n[off+2:])
	return n[off+4+kLen : off+4+kLen+vLen]
}

func (n *node) Pointer(i uint16) int64 {
	if !n.indexInBounds(i) {
		panic(ErrIndexOutOfBounds)
	}
	off := n.offset(i)
	kLen := binary.LittleEndian.Uint16(n[off:])
	return int64(binary.LittleEndian.Uint64(n[off+2+kLen : off+2+kLen+8]))
}

// Binary searches the target key inside n and returns its position and weither it exists.
func (n *node) Search(target []byte) (uint16, bool) {
	l := n.N()
	left, right := uint16(0), l

	for left < right {
		cur := uint16(uint(left+right) >> 1) // #nosec G115 // right shift stops overflow
		if cmp := bytes.Compare(n.Key(cur), target); cmp == -1 {
			left = cur + 1
		} else if cmp == 1 {
			right = cur
		} else {
			return cur, true
		}
	}
	return left, left < l && bytes.Equal(n.Key(left), target)
}

// Returns a copy of n with k-v set at position i. If the key at i is equal to k it will be
// overwritten an n.N will stay the same otherwise k-v will be inserted as a new cell and n.N will
// be increased by 1.
//
// Overwriting a k-v pair does not overwrite the data stored in the original cell, it mereley
// overwrites the reference to it. To free up the space taken up by unreferenced cells use Vacuum.
//
// WARNING: No additional check is performed weither i is the correct position for k. Meaning it is
// the callers responsibility to ensure that k belongs at position i to ensure the order of the keys
// will not break. Always use Search and CanSet before using Set to ensure that n has enough space
// for the k-v pair and that the value of i is correct.
func (n *node) Set(i uint16, k, v []byte) node {
	l := n.N()
	cell := makeCell(k, v)

	wCur := n.wCursor()
	off := wCur - uint16(len(cell))

	var res node
	copy(res[:], n[:])
	copy(res[off:wCur], cell)

	res.setWCursor(off)

	if bytes.Equal(k, n.Key(i)) {
		res.setOffset(i, off)
		return res
	}

	trailingOffs := n[offPos(i) : offPos(l)+2]
	copy(res[offPos(i+1):], trailingOffs)
	res.setOffset(i, off)
	res.setN(l + 1)

	return res
}

// Returns true if n has enough space left in its void to add the given k-v pair. CanSet always
// assumes that k does not exist.
func (n *node) CanSet(k, v []byte) bool {
	return n.voidSize() >= 6+len(k)+len(v)
}

// Returns a copy of n with the k-v pair at index i deleted.
//
// Delete mereley deletes the reference to the cell and does not free up the cells space. To free up
// the space taken up by unreferenced cells use Vacuum.
func (n *node) Delete(i uint16) node {
	if !n.indexInBounds(i) {
		panic(ErrIndexOutOfBounds)
	}

	l := n.N()

	var res node
	copy(res[:], n[:])

	trailingOffs := n[offPos(i) : offPos(l)+2]
	copy(res[offPos(i):], trailingOffs[2:])
	res.setOffset(l, 0)
	res.setN(l - 1)

	return res
}

// Splits n into two separate nodes.
func (n *node) Split() (left node, right node) {
	var addToRight bool
	var i uint16
	var wc uint16 = page.Size

	left, right = newNode(n.Type()), newNode(n.Type())

	thresh := (page.Size - wc) / 2

	addToNode := func(addTo *node, i uint16, cell []byte, wCursor *uint16) {
		*wCursor -= uint16(len(cell))
		addTo.setOffset(i, *wCursor)
		copy(addTo[*wCursor:], cell)
		addTo.setWCursor(*wCursor)
	}

	for k, v := range n.All() {
		cell := makeCell(k, v)

		if addToRight {
			addToNode(&right, i, cell, &wc)
			i++
			continue
		}

		addToNode(&left, i, cell, &wc)
		i++

		if wc < page.Size-thresh {
			addToRight = true
		}
	}
	return left, right
}

// Returns a resorted and reduced copy of n by freeing up space used by unreferenced cells.
func (n *node) Vacuum() node {
	var vacuumed node
	vacuumed[0] = byte(n.Type())
	vacuumed.setN(n.N())

	var wc uint16 = page.Size
	var i uint16
	for k, v := range n.All() {
		cell := makeCell(k, v)
		wc -= uint16(len(cell))
		vacuumed.setOffset(i, wc)
		copy(vacuumed[wc:], cell)

		i++
	}
	vacuumed.setWCursor(wc)

	return vacuumed
}

// An iterator over all key-value pairs of n.
func (n *node) All() iter.Seq2[[]byte, []byte] {
	return n.AllFrom(0)
}

// An iterator over all key-value pairs of n starting from position i.
func (n *node) AllFrom(i uint16) iter.Seq2[[]byte, []byte] {
	return func(yield func([]byte, []byte) bool) {
		for ; i < n.N(); i++ {
			k, v := n.Key(i), n.Val(i)
			if !yield(k, v) {
				return
			}
		}
	}
}

func (n *node) setN(nc uint16) {
	binary.LittleEndian.PutUint16(n[nOff:], nc)
}

func (n *node) offset(i uint16) uint16 {
	if n.indexInBounds(i) {
		panic(ErrIndexOutOfBounds)
	}
	return binary.LittleEndian.Uint16(n[offPos(i):])
}

func (n *node) setOffset(i, off uint16) {
	if n.indexInBounds(i) {
		panic(ErrIndexOutOfBounds)
	}
	binary.LittleEndian.PutUint16(n[offPos(i):], off)
}

func (n *node) indexInBounds(i uint16) bool {
	return i >= n.N()
}

func (n *node) wCursor() uint16 {
	return binary.LittleEndian.Uint16(n[wCurOff:])
}

func (n *node) setWCursor(wc uint16) {
	binary.LittleEndian.PutUint16(n[wCurOff:], wc)
}

func (n *node) voidSize() int {
	return int(n.wCursor()) - int(offPos(n.N())+2)
}

// Returns the calculated position to the offset inside the offset list. This does NOT return the
// offset itself only the reference to the offset.
func offPos(i uint16) uint16 { return dataOff + 2*i }

func makeCell(k, v []byte) []byte {
	cell := make([]byte, 4+len(k)+len(v))
	binary.LittleEndian.PutUint16(cell[0:], uint16(len(k)))
	binary.LittleEndian.PutUint16(cell[2:], uint16(len(v)))
	copy(cell[4:], k)
	copy(cell[4+len(k):], v)
	return cell
}
