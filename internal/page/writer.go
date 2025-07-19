package page

type Writer struct {
	freelist set[int64]
	freed    set[int64]
	new      map[int64][Size]byte
	nextPage int64
	pageSize int64
	commit   commitFn
	*Reader
}

func newWriter(r *Reader, freelist set[int64], nextPage int64, commitCallback commitFn) *Writer {
	return &Writer{
		Reader: r,

		freelist: freelist,
		freed:    make(set[int64]),
		new:      make(map[int64][Size]byte),

		commit:   commitCallback,
		nextPage: nextPage,
	}
}

func (w *Writer) Alloc(d [Size]byte) int64 {
	off := w.nextPage
	switch {
	case len(w.freed) > 0:
		for off := range w.freed {
			delete(w.freed, off)
			break
		}
	case len(w.freelist) > 0:
		for off := range w.freelist {
			delete(w.freelist, off)
			break
		}
	default:
		w.nextPage += w.pageSize
	}
	w.new[off] = d
	return off
}

func (w *Writer) Free(off int64) {
	if _, ok := w.freelist[off]; ok {
		return
	}
	if _, ok := w.freed[off]; ok {
		return
	}
	w.freed[off] = struct{}{}
}

func (w *Writer) Commit() error {
	return w.commit(w.new)
}

func (w *Writer) Abort() {}
