package ioutils

import (
	"io"
)

type readCloserWrapper struct {
	io.Reader
	closer func() error
}

func (r *readCloserWrapper) Close() error {
	return r.closer()
}

func NewReadCloserWrapper(r io.Reader, closer func() error) io.ReadCloser {
	return &readCloserWrapper{
		Reader: r,
		closer: closer,
	}
}

type readerErrWrapper struct {
	reader io.Reader
	closer func()
}

func (r *readerErrWrapper) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err != nil {
		r.closer()
	}
	return n, err
}

func NewReaderErrWrapper(r io.Reader, closer func()) io.Reader {
	return &readerErrWrapper{
		reader: r,
		closer: closer,
	}
}

type buffer struct {
	data []byte
	size int
	err  error
}

type bufReader struct {
	returning chan *buffer
	outgoing  chan *buffer
	reader    io.Reader
	err       error
}

func NewBufReader(r io.Reader) *bufReader {
	reader := &bufReader{
		returning: make(chan *buffer, 32),
		outgoing:  make(chan *buffer, 32),
		reader:    r,
	}
	reader.seed(32)
	go reader.drain()
	return reader
}

func (r *bufReader) seed(n int) {
	var data []byte
	var buf *buffer
	for i := 0; i < n; i++ {
		data = make([]byte, 1024)
		buf = &buffer{
			data,
			0,
			nil,
		}
		r.returning <- buf
	}
}

func (r *bufReader) drain() {
	var buf *buffer
	for {
		buf = <-r.returning
		n, err := r.reader.Read(buf.data)
		buf.size = n
		buf.err = err
		r.outgoing <- buf
		if err != nil {
			break
		}
	}
}

func (r *bufReader) returnBuffer(buf *buffer) {
	r.returning <- buf
}

func (r *bufReader) Read(p []byte) (n int, err error) {
	var buf *buffer
	buf = <-r.outgoing
	n = buf.size
	err = buf.err
	if buf.size == 0 && buf.err != nil {
		return 0, buf.err
	}
	copy(p, buf.data[0:buf.size])
	r.returning <- buf
	return
}

func (r *bufReader) Close() error {
	close(r.outgoing)
	close(r.returning)
	closer, ok := r.reader.(io.ReadCloser)
	if !ok {
		return nil
	}
	return closer.Close()
}
