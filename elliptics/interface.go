package elliptics

import (
	"fmt"
	"time"
)

// implements Reader and Seeker interfaces
type ReadSeeker struct {
	session		*Session
	key		*Key

	total_size	uint64
	record_flags	uint64

	// updated only when transferring data to caller, not when reading from elliptics
	offset		int64

	read_offset	uint64
	read_size	uint64
	chunk		[]byte

	Mtime		time.Time
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

	_, err = r.ReadInternal(r.chunk)
	if err != nil {
		key.Free()

		return nil, err
	}

	return r, nil
}

func (r *ReadSeeker) Free() {
	r.key.Free()
}

func (r *ReadSeeker) ReadInternal(buf []byte) (n int, err error) {
	errors := make([]error, 0)

	// if we have already read at least some data and this object doesn't have chunked checksum
	// disable checksum verification, since the first call has already checked the whole file
	if r.total_size != 0 && (r.record_flags & DNET_RECORD_FLAGS_CHUNKED_CSUM) == 0 {
		r.session.SetIOflags(r.session.GetIOflags() | DNET_IO_FLAGS_NOCSUM)
	}

	r.read_offset = uint64(r.offset)
	for rd := range r.session.ReadInto(r.key, r.read_offset, buf) {
		err = rd.Error()
		if err != nil {
			errors = append(errors, err)
			continue
		}

		r.read_size = rd.IO().Size
		r.record_flags = rd.IO().RecordFlags
		r.total_size = rd.IO().TotalSize
		r.Mtime = rd.IO().Timestamp

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

func (r *ReadSeeker) Read(p []byte) (n int, err error) {
	ioflags := r.session.GetIOflags()
	defer r.session.SetIOflags(ioflags)

	offset := uint64(r.offset)

	for {
		if r.read_size != 0 && len(r.chunk) != 0 && r.read_offset <= offset {
			// we have read and cached enough data (---) to satisfy client's request (+++)
			// |-------------+++++++++++++++--------------------------------|
			// ^             ^             ^                                ^
			// |-read_offset |-offset      |-offset + len(p)                |-read_offset + read_size
			if r.read_offset + r.read_size >= offset + uint64(len(p)) {
				n = copy(p, r.chunk[offset - r.read_offset :])
				r.offset += int64(n)
				return n, nil
			}

			// we have read and cached end of the file, but client request
			// spans past the end of the file - we still can satisfy client's request
			// we have to return what we have
			// |-------------++++++++++++|++++++++++++++++++++++++++++++++|
			// ^             ^           ^                                ^
			// |-read_offset |-offset    |-read_offset + read_size        |-offset + len(p)
			//                           |-end of the file
			if r.read_offset + r.read_size == r.total_size {
				n = copy(p, r.chunk[offset - r.read_offset :])
				r.offset += int64(n)
				return n, nil
			}
		}

		if len(p) > len(r.chunk) {
			n, err = r.ReadInternal(p)
			if err != nil {
				return 0, err
			}

			r.offset += int64(n)
			return n, nil
		}

		n, err = r.ReadInternal(r.chunk)
		if err != nil {
			return 0, nil
		}
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
