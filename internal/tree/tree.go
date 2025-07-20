package tree

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gkits/pavosql/internal/page"
)

type pageReader interface {
	Read(off int64) ([page.Size]byte, error)
}

type pageWriter interface {
	Alloc(d [page.Size]byte) int64
	Free(off int64)
	Commit() error
	Abort()
}

type pageReadWriter interface {
	pageReader
	pageWriter
}

type Tree struct {
	root int64
}

func New() *Tree {
	return &Tree{}
}

func (t *Tree) Get(r pageReader, k []byte) ([]byte, error) {
	pg, err := r.Read(t.root)
	if err != nil {
		return nil, err
	}
	cur := node(pg)

	for {
		i, exists := cur.Search(k)

		switch page.GetType(cur) {
		case page.Pointer:
			ptr := cur.Pointer(i)
			pg, err = r.Read(ptr)
			if err != nil {
				return nil, fmt.Errorf("tree: failed to read page: %w", err)
			}
			cur = node(pg)
		case page.Leaf:
			if !exists {
				return nil, errors.New("key does not exists on leaf node")
			}
			return cur.Val(i), nil
		default:
			return nil, errors.New("invalid page type")
		}
	}
}

func (t *Tree) Set(wr pageReadWriter, k []byte, v []byte) error {
	pg, err := wr.Read(t.root)
	if err != nil {
		return fmt.Errorf("tree: failed to read root page: %w", err)
	}
	cur := node(pg)

	visited := []node{cur}
	for {
		i, exists := cur.Search(k)

		switch page.GetType(cur) {
		case page.Pointer:
			ptr := cur.Pointer(i)
			pg, err = wr.Read(ptr)
			if err != nil {
				return fmt.Errorf("tree: failed to read page: %w", err)
			}
			cur = node(pg)
			visited = append(visited, cur)
			continue

		case page.Leaf:
			if !exists {
				return errors.New("key does not exists on leaf node")
			}

			if cur.CanSet(k, v) {
				newNode := cur.Set(i, k, v)
				ptr := wr.Alloc(newNode)
				// TODO: pass ptr to parent node
				_ = ptr
				break
			}

			// TODO: handle splitting
			left, right := cur.Split()
			_, _ = left, right

		default:
			return errors.New("invalid page type")
		}
		break
	}

	for _, n := range slices.Backward(visited) {
		ptr := wr.Alloc(n)
		// TODO: pass ptr to parent nodes
		_ = ptr
	}

	return nil
}

func (t *Tree) Delete(k []byte) error {
	return nil
}
