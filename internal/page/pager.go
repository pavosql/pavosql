package page

import (
	"sync"

	"github.com/gkits/pavosql/pkg/atomic"
)

const Size = 8192

type Type uint8

const (
	Pointer Type = iota + 1
	Leaf
)

type Pager struct {
	rw       atomic.ReadWriterAt
	freeList int64
	mu       sync.RWMutex
	end      int64
}

type (
	readFn   = func(int64) ([Size]byte, error)
	commitFn = func(map[int64][Size]byte) error

	set[T comparable] = map[T]struct{}
)

func (p *Pager) NewReader() (*Reader, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	reader := newReader(p.read)

	return reader, nil
}

func (p *Pager) NewWriter() (*Writer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	writer := newWriter(newReader(p.read), make(set[int64]), p.end, p.commit)

	return writer, nil
}

func (p *Pager) read(off int64) ([Size]byte, error) {
	page := [Size]byte{}
	if _, err := p.rw.ReadAt(page[:], int64(off)); err != nil {
		return page, err
	}
	return page, nil
}

func (p *Pager) commit(changes map[int64][Size]byte) error {
	for off, d := range changes {
		if _, err := p.rw.WriteAt(d[:], off); err != nil {
			return err
		}
	}
	return p.rw.Commit()
}
