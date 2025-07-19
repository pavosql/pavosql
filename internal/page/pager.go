package page

import (
	"sync"

	"github.com/gkits/pavosql/pkg/atomic"
)

const PageSize = 8192

type PageType uint8

const (
	PointerPage PageType = iota + 1
	LeafPage
)

type Pager struct {
	rw       atomic.ReadWriterAt
	freeList int64
	mu       sync.RWMutex
	end      int64
}

type (
	readFn   = func(int64) ([]byte, error)
	commitFn = func(map[int64][]byte) error

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

func (p *Pager) read(off int64) ([]byte, error) {
	page := make([]byte, 99)
	if _, err := p.rw.ReadAt(page, int64(off)); err != nil {
		return nil, err
	}
	return page, nil
}

func (p *Pager) commit(changes map[int64][]byte) error {
	for off, d := range changes {
		if _, err := p.rw.WriteAt(d, off); err != nil {
			return err
		}
	}
	return p.rw.Commit()
}
