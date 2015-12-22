package elliptics

import (
	"fmt"
)

// implements Reader and Seeker interfaces
type ReadSeeker struct {
	session		*Session
	key		*Key

	total_size	uint64
	record_flags	uint64

	offset		int64

	read_offset	uint64
	read_size	uint64
	chunk		[]byte
}

func NewReadSeeker(session *Session, kstr string) (*ReadSeeker, error) {
	key, err := NewKey(kstr)
	if err != nil {
		return nil, err
	}

	r := &ReadSeeker {
		session:		session,
		key:			key,
		chunk:			make([]byte, 10 * 1024 * 1024),
	}

	_, err = r.Read(r.chunk)
	if err != nil {
		key.Free()

		return nil, err
	}

	return r, nil
}

func (r *ReadSeeker) Free() {
	r.key.Free()
}

func (r *ReadSeeker) Read(p []byte) (n int, err error) {
	errors := make([]error, 0)

	offset := uint64(r.offset)

	if r.read_size != 0 && len(r.chunk) != 0 && r.read_offset <= offset {
		// we have read and cached enough data (---) to satisfy client's request (+++)
		// |-------------+++++++++++++++--------------------------------|
		// ^             ^             ^                                ^
		// |-read_offset |-offset      |-offset + len(p)                |-read_offset + read_size
		if r.read_offset + r.read_size >= offset + uint64(len(p)) {
			copied := copy(p, r.chunk[offset - r.read_offset :])
			return copied, nil
		}

		// we have read and cached end of the file, but client request
		// spans past the end of the file - we still can satisfy client's request
		// we have to return what we have
		// |-------------++++++++++++|++++++++++++++++++++++++++++++++|
		// ^             ^           ^                                ^
		// |-read_offset |-offset    |-read_offset + read_size        |-offset + len(p)
		//                           |-end of the file
		if r.read_offset + r.read_size == r.total_size {
			copied := copy(p, r.chunk[offset - r.read_offset :])
			return copied, nil
		}
	}

	r.read_offset = uint64(r.offset)
	for rd := range r.session.ReadInto(r.key, r.read_offset, p) {
		err = rd.Error()
		if err != nil {
			errors = append(errors, err)
			continue
		}

		r.read_size = rd.IO().Size
		r.record_flags = rd.IO().RecordFlags
		r.total_size = rd.IO().TotalSize

		r.offset += int64(r.read_size)
		return int(r.read_size), nil
	}

	code := -6
	if len(errors) != 0 {
		for _, err = range errors {
			code = ErrorCode(err)
			if code != -110 {
				break
			}
		}
	}

	return 0, &DnetError {
		Code:  code,
		Flags: 0,
		Message: fmt.Sprintf(
			"read-seeker error: current-offset: %d, total-size: %d, errors: %v",
			r.offset, r.total_size, errors),
	}
}

func (r *ReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		r.offset = offset
	case 1:
		r.offset += offset
	case 2:
		r.offset = int64(r.total_size) + offset
	}

	return r.offset, nil
}
