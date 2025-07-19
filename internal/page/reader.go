package page

type Reader struct {
	pages map[int64][Size]byte
	read  readFn
}

func newReader(callbackRead readFn) *Reader {
	return &Reader{make(map[int64][Size]byte), callbackRead}
}

func (r *Reader) Read(off int64) ([Size]byte, error) {
	if page, ok := r.pages[off]; ok {
		return page, nil
	}

	page, err := r.read(off)
	if err != nil {
		return page, err
	}
	r.pages[off] = page
	return page, nil
}
