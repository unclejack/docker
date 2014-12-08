package ioutils

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

func TestBufReader(t *testing.T) {
	reader, writer := io.Pipe()
	bufreader := NewBufReader(reader)

	// Write everything down to a Pipe
	// Usually, a pipe should block but because of the buffered reader,
	// the writes will go through
	done := make(chan bool)
	go func() {
		writer.Write([]byte("hello world"))
		writer.Close()
		done <- true
	}()

	// Drain the reader *after* everything has been written, just to verify
	// it is indeed buffering
	<-done
	output, err := ioutil.ReadAll(bufreader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, []byte("hello world")) {
		t.Errorf("expected 'hello world', got %q", string(output))
	}
}

type repeatedReader struct {
	readCount int
	maxReads  int
}

func newRepeatedReader(max int) *repeatedReader {
	return &repeatedReader{0, max}
}

func (r *repeatedReader) Read(p []byte) (int, error) {
	if r.readCount >= r.maxReads {
		return 0, io.EOF
	}
	r.readCount++
	copy([]byte{'b', 'a', 'r', 'b', 'a', 'r', 'b', 'a', 'r', 'b'}, p)
	return 10, nil
}

func Benchmark1M3BytesReads(b *testing.B) {
	reads := 1000000
	b.SetBytes(10 * int64(reads))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := newRepeatedReader(reads)
		bufReader := NewBufReader(reader)
		io.Copy(ioutil.Discard, bufReader)
	}

}
