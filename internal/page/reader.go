package page

type Reader struct {
	pages map[int64][]byte
	read  readFn
}

func newReader(callbackRead readFn) *Reader {
	return &Reader{make(map[int64][]byte), callbackRead}
}

func (r *Reader) Read(off int64) ([]byte, error) {
	if page, ok := r.pages[off]; ok {
		return page, nil
	}

	page, err := r.read(off)
	if err != nil {
		return nil, err
	}
	r.pages[off] = page
	return page, nil
}
