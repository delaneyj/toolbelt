package bytebufferpool

import "io"

// ByteBuffer provides a byte buffer that minimizes allocations.
//
// Use Get for obtaining an empty byte buffer.
type ByteBuffer struct {
	buf []byte
}

// Len returns the size of the byte buffer.
func (b *ByteBuffer) Len() int {
	return len(b.buf)
}

// ReadFrom implements io.ReaderFrom by appending all data read from r.
func (b *ByteBuffer) ReadFrom(r io.Reader) (int64, error) {
	p := b.buf
	nStart := int64(len(p))
	nMax := int64(cap(p))
	n := nStart
	if nMax == 0 {
		nMax = 64
		p = make([]byte, nMax)
	} else {
		p = p[:nMax]
	}
	for {
		if n == nMax {
			nMax *= 2
			bNew := make([]byte, nMax)
			copy(bNew, p)
			p = bNew
		}
		nn, err := r.Read(p[n:])
		n += int64(nn)
		if err != nil {
			b.buf = p[:n]
			n -= nStart
			if err == io.EOF {
				return n, nil
			}
			return n, err
		}
	}
}

// WriteTo implements io.WriterTo.
func (b *ByteBuffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b.buf)
	return int64(n), err
}

// Bytes returns the accumulated bytes in the buffer.
//
// The returned slice aliases the internal buffer.
func (b *ByteBuffer) Bytes() []byte {
	return b.buf
}

// Write implements io.Writer by appending p to the buffer.
func (b *ByteBuffer) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteByte appends the byte c to the buffer.
//
// The function always returns nil.
func (b *ByteBuffer) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// WriteString appends s to the buffer.
func (b *ByteBuffer) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// Set replaces the buffer contents with p.
func (b *ByteBuffer) Set(p []byte) {
	b.buf = append(b.buf[:0], p...)
}

// SetString replaces the buffer contents with s.
func (b *ByteBuffer) SetString(s string) {
	b.buf = append(b.buf[:0], s...)
}

// String returns the string representation of the buffer contents.
func (b *ByteBuffer) String() string {
	return string(b.buf)
}

// Reset makes the buffer empty.
func (b *ByteBuffer) Reset() {
	b.buf = b.buf[:0]
}
