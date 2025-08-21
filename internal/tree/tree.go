package tree

import (
	"errors"
	"fmt"

	"github.com/gkits/pavosql/internal/tree/node"
)

type pageReader interface {
	Read(off uint64) ([]byte, error)
}

type pageWriter interface {
	Alloc(d []byte) uint64
	Free(off uint64)
	Commit() error
	Abort()
}

type pageReadWriter interface {
	pageReader
	pageWriter
}

type Tree struct {
	root uint64
}

func New() *Tree {
	return &Tree{}
}

func (t *Tree) Get(r pageReader, k []byte) ([]byte, error) {
	cur, err := deserializeNodePage(r, t.root)
	if err != nil {
		return nil, fmt.Errorf("tree: failed to deserialize root node: %w", err)
	}

	for {
		i, exists := cur.Search(k)

		switch n := cur.(type) {
		case *node.Internal:
			ptr, err := n.Ptr(i)
			if err != nil {
				return nil, fmt.Errorf("tree: failed to get pointer: %w", err)
			}
			cur, err = deserializeNodePage(r, ptr)
			if err != nil {
				return nil, fmt.Errorf("tree: failed to deserialize root node: %w", err)
			}
		case *node.Leaf:
			if !exists {
				return nil, errors.New("tree: key does not exists on leaf node")
			}
			v, err := n.Val(i)
			if err != nil {
				return nil, fmt.Errorf("tree: failed to get value: %w", err)
			}
			return v, nil
		default:
			return nil, fmt.Errorf("tree: invalid node")
		}
	}
}

func (t *Tree) Insert(rw pageReadWriter, k []byte, v []byte) error {
	if err := t.set(rw, k, v, false); err != nil {
		return fmt.Errorf("tree: failed to insert key-value pair: %w", err)
	}
	return nil
}

func (t *Tree) Update(rw pageReadWriter, k []byte, v []byte) error {
	if err := t.set(rw, k, v, false); err != nil {
		return fmt.Errorf("tree: failed to update key-value pair: %w", err)
	}
	return nil
}

func (t *Tree) Delete(rw pageReadWriter, k []byte) error {
	return nil
}

type keyPtr struct {
	key []byte
	ptr uint64
}

func (t *Tree) set(rw pageReadWriter, k []byte, v []byte, shouldExist bool) error {
	root, err := deserializeNodePage(rw, t.root)
	if err != nil {
		return fmt.Errorf("failed to deserialize root node: %w", err)
	}

	kPs, err := t.setRecursive(rw, root, k, v, shouldExist)
	if err != nil {
		return fmt.Errorf("failed to set key-value pair: %w", err)
	}
	if len(kPs) == 1 {
		t.root = kPs[0].ptr
		return nil
	}

	newRoot := node.NewInternal()
	for i, kP := range kPs {
		if err := newRoot.Set(i, kP.key, kP.ptr); err != nil {
			return fmt.Errorf("failed to setup new root node: %w", err)
		}
	}
	if t.root, err = allocateNode(rw, newRoot); err != nil {
		return fmt.Errorf("failed to allocate new root node: %w", err)
	}
	return nil
}

func (t *Tree) setRecursive(rw pageReadWriter, n node.Node, k []byte, v []byte, shouldExist bool) ([]keyPtr, error) {
	i, exists := n.Search(k)

	switch cur := n.(type) {
	case *node.Internal:
		ptr, err := cur.Ptr(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get pointer: %w", err)
		}
		child, err := deserializeNodePage(rw, ptr)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize root node: %w", err)
		}
		rw.Free(ptr)
		ptrs, err := t.setRecursive(rw, child, k, v, shouldExist)
		if err != nil {
			return nil, fmt.Errorf("failed to set child node: %w", err)
		}

		for _, kP := range ptrs {
			if err := cur.Set(i, kP.key, kP.ptr); err != nil {
				return nil, fmt.Errorf("failed to set key-pointer pair: %w", err)
			}
		}

		var newPtrs []keyPtr
		for _, n := range child.Split(99) {
			newPtr, err := allocateNode(rw, n)
			if err != nil {
				return nil, fmt.Errorf("failed to allocate split node: %w", err)
			}
			newK, err := n.Key(0)
			if err != nil {
				return nil, fmt.Errorf("failed to get first node key: %w", err)
			}
			newPtrs = append(newPtrs, keyPtr{key: newK, ptr: newPtr})
		}
		return newPtrs, nil

	case *node.Leaf:
		if exists != shouldExist {
			return nil, errors.New("key existing state does not match")
		}
		if err := cur.Set(i, k, v); err != nil {
			return nil, fmt.Errorf("failed to set key-value pair: %w", err)
		}

		var newPtrs []keyPtr
		for _, n := range cur.Split(99) {
			newPtr, err := allocateNode(rw, n)
			if err != nil {
				return nil, fmt.Errorf("failed to allocate split node: %w", err)
			}
			newK, err := n.Key(0)
			if err != nil {
				return nil, fmt.Errorf("failed to get first node key: %w", err)
			}
			newPtrs = append(newPtrs, keyPtr{key: newK, ptr: newPtr})
		}
		return newPtrs, nil

	default:
		return nil, fmt.Errorf("invalid node")
	}
}

func deserializeNodePage(r pageReader, ptr uint64) (node.Node, error) {
	pg, err := r.Read(ptr)
	if err != nil {
		return nil, fmt.Errorf("failed to read page: %w", err)
	}
	n, err := node.Deserialize(pg)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize page: %w", err)
	}
	return n, nil
}

func allocateNode(w pageWriter, n node.Node) (uint64, error) {
	d, err := node.Serialize(n)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize node: %w", err)
	}
	return w.Alloc(d), nil
}
