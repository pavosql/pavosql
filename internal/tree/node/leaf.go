package node

import (
	"bytes"
	"errors"
	"slices"
)

type Leaf struct {
	head       Header
	keys, vals [][]byte
}

func (n *Leaf) Type() Type {
	return n.head.typ
}

func (n *Leaf) Key(i int) ([]byte, error) {
	if i < 0 || i >= int(n.head.count) {
		return nil, errors.New("node: index out of range")
	}
	return n.keys[i], nil
}

func (n *Leaf) Val(i int) ([]byte, error) {
	if i < 0 || i >= int(n.head.count) {
		return nil, errors.New("node: index out of range")
	}
	return n.vals[i], nil
}

func (n *Leaf) Set(i int, key, val []byte) error {
	if i < 0 || i > int(n.head.count) {
		return errors.New("node: index out of range")
	}
	if bytes.Equal(n.keys[i], key) {
		n.vals[i] = val
		return nil
	}
	n.keys = slices.Insert(n.keys, i, key)
	n.vals = slices.Insert(n.vals, i, val)
	return nil
}

func (n *Leaf) Del(i int) error {
	if i < 0 || i >= int(n.head.count) {
		return errors.New("node: index out of range")
	}
	n.keys = slices.Delete(n.keys, i, i)
	n.vals = slices.Delete(n.vals, i, i)
	return nil
}

func (n *Leaf) Search(key []byte) (int, bool) {
	return slices.BinarySearchFunc(n.keys, key, bytes.Compare)
}

func (n *Leaf) Split(maxSize int) []Node {
	return nil
}

func (n *Leaf) Merge(right *Leaf) (*Leaf, error) {
	return nil, nil
}
