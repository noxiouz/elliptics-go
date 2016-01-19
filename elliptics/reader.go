package elliptics

import (
	"fmt"
	"io"
	"time"
)

// implements Reader and Seeker interfaces
type ReadSeeker struct {
	session		*Session

	key		*Key
	// if true, Free() method will free the key
	// if ReadSeeker has been created with foreign key,
	// for example using NewReadSeekerKey(),
	// then we can not remove key, it belongs to the caller
	want_key_free	bool

	TotalSize	uint64
	RecordFlags	uint64

	// updated only when transferring data to caller, not when reading from elliptics
	offset		int64

	// offset to remote elliptics key which we use (have used) to read data into @chunk
	read_offset	uint64
	read_size	uint64
	chunk		[]byte

	Mtime		time.Time
}

func NewEmptyReadSeeker() (*ReadSeeker, error) {
	r := &ReadSeeker {
		session:		nil,
		key:			nil,
		want_key_free:		false,
		chunk:			make([]byte, 10 * 1024 * 1024),
	}

	return r, nil
}

func NewReadSeeker(session *Session, kstr string) (*ReadSeeker, error) {
	key, err := NewKey(kstr)
	if err != nil {
		return nil, err
	}

	rs, err := NewReadSeekerKeyOffsetSize(session, key, 0, 0)
	if err != nil {
		key.Free()
		return nil, err
	}

	rs.want_key_free = true

	return rs, nil
}

func NewReadSeekerOffsetSize(session *Session, kstr string, offset, size uint64) (*ReadSeeker, error) {
	key, err := NewKey(kstr)
	if err != nil {
		return nil, err
	}

	rs, err := NewReadSeekerKeyOffsetSize(session, key, offset, size)
	if err != nil {
		key.Free()
		return nil, err
	}

	rs.want_key_free = true

	return rs, nil
}

func NewReadSeekerKeyOffsetSize(session *Session, key *Key, offset, size uint64) (*ReadSeeker, error) {
	if size == 0 {
		size = 10 * 1024 * 1024
	}

	r := &ReadSeeker {
		session:		session,
		key:			key,
		want_key_free:		false,
		offset:			int64(offset),
		chunk:			make([]byte, size),
	}

	_, err := r.ReadInternal(r.chunk)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *ReadSeeker) Free() {
	if r.want_key_free {
		r.key.Free()
	}
}

func (r *ReadSeeker) ReadInternal(buf []byte) (n int, err error) {
	if r.session == nil || r.key == nil {
		return 0, fmt.Errorf("trying to read from empty interface")
	}

	ioflags := r.session.GetIOflags()
	defer r.session.SetIOflags(ioflags)

	errors := make([]error, 0)

	// if we have already read at least some data and this object doesn't have chunked checksum
	// disable checksum verification, since the first call has already checked the whole file
	if r.TotalSize != 0 && (r.RecordFlags & DNET_RECORD_FLAGS_CHUNKED_CSUM) == 0 {
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
		r.RecordFlags = rd.IO().RecordFlags
		r.TotalSize = rd.IO().TotalSize
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
			r.offset, r.TotalSize, errors),
	}
}

func (r *ReadSeeker) Read(p []byte) (n int, err error) {
	if r.session == nil || r.key == nil {
		return 0, fmt.Errorf("trying to read from empty interface")
	}

	offset := uint64(r.offset)
	if offset >= r.TotalSize {
		return 0, io.EOF
	}

	for {
		if r.read_size != 0 && len(r.chunk) != 0 && r.read_offset <= offset && offset - r.read_offset < uint64(len(r.chunk)) {
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
			if r.read_offset + r.read_size == r.TotalSize {
				n = copy(p, r.chunk[offset - r.read_offset :])
				//fmt.Printf("read_offset: %d, read_size: %d, total_size: %d, offset: %d, p-size: %d, n: %d\n",
				//	r.read_offset, r.read_size, r.TotalSize, offset, len(p), n)
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
			return 0, err
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
		r.offset = int64(r.TotalSize) + offset
	}

	return r.offset, nil
}

func (r *ReadSeeker) SetKey(session *Session, key *Key) error {
	r.Free()

	r.session = session
	r.key = key
	r.want_key_free = false

	r.TotalSize = 0
	r.RecordFlags = 0

	r.offset = 0
	r.read_offset = 0
	r.read_size = 0

	_, err := r.ReadInternal(r.chunk)
	if err != nil {
		return err
	}

	return nil
}
