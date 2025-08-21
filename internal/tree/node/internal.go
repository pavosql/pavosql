package node

import (
	"bytes"
	"errors"
	"slices"
)

type Internal struct {
	head Header
	keys [][]byte
	ptrs []uint64
}

func NewInternal() *Internal {
	return &Internal{
		head: Header{
			typ:   TypeInternal,
			count: 0,
		},
		keys: make([][]byte, 0),
		ptrs: make([]uint64, 0),
	}
}

func (n *Internal) Type() Type {
	return n.head.typ
}

func (n *Internal) Key(i int) ([]byte, error) {
	return n.keys[i], nil
}

func (n *Internal) Ptr(i int) (uint64, error) {
	return n.ptrs[i], nil
}

func (n *Internal) Set(i int, key []byte, ptr uint64) error {
	if i < 0 || i > int(n.head.count) {
		return errors.New("node: index out of range")
	}
	if bytes.Equal(n.keys[i], key) {
		n.ptrs[i] = ptr
		return nil
	}
	n.keys = slices.Insert(n.keys, i, key)
	n.ptrs = slices.Insert(n.ptrs, i, ptr)
	return nil
}

func (n *Internal) Del(i int) error {
	if i < 0 || i >= int(n.head.count) {
		return errors.New("node: index out of range")
	}
	n.keys = slices.Delete(n.keys, i, i)
	n.ptrs = slices.Delete(n.ptrs, i, i)
	return nil
}

func (n *Internal) Search(key []byte) (int, bool) {
	return slices.BinarySearchFunc(n.keys, key, bytes.Compare)
}

func (n *Internal) Split(maxSize int) []Node {
	return nil
}

func (n *Internal) Merge(right *Internal) (*Internal, error) {
	return nil, nil
}
