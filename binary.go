package toolbelt

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// WriteUint8 writes v in little-endian order.
func WriteUint8(w io.Writer, v uint8) error {
	_, err := w.Write([]byte{v})
	return err
}

// ReadUint8 reads a little-endian uint8.
func ReadUint8(r io.Reader) (uint8, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}

// WriteUint32 writes v in little-endian order.
func WriteUint32(w io.Writer, v uint32) error {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	_, err := w.Write(b[:])
	return err
}

// ReadUint32 reads a little-endian uint32.
func ReadUint32(r io.Reader) (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

// WriteUint64 writes v in little-endian order.
func WriteUint64(w io.Writer, v uint64) error {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	_, err := w.Write(b[:])
	return err
}

// ReadUint64 reads a little-endian uint64.
func ReadUint64(r io.Reader) (uint64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

// WriteInt32 writes v in little-endian order.
func WriteInt32(w io.Writer, v int32) error {
	return WriteUint32(w, uint32(v))
}

// ReadInt32 reads a little-endian int32.
func ReadInt32(r io.Reader) (int32, error) {
	u, err := ReadUint32(r)
	return int32(u), err
}

// WriteInt64 writes v in little-endian order.
func WriteInt64(w io.Writer, v int64) error {
	return WriteUint64(w, uint64(v))
}

// ReadInt64 reads a little-endian int64.
func ReadInt64(r io.Reader) (int64, error) {
	u, err := ReadUint64(r)
	return int64(u), err
}

// WriteFloat32 writes v in little-endian IEEE 754 binary form.
func WriteFloat32(w io.Writer, v float32) error {
	return WriteUint32(w, math.Float32bits(v))
}

// ReadFloat32 reads a little-endian IEEE 754 float32.
func ReadFloat32(r io.Reader) (float32, error) {
	u, err := ReadUint32(r)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(u), nil
}

// WriteString writes a length-prefixed string with a uint32 length.
func WriteString(w io.Writer, s string) error {
	if len(s) > int(^uint32(0)) {
		return fmt.Errorf("toolbelt: string too large: %d", len(s))
	}
	if err := WriteUint32(w, uint32(len(s))); err != nil {
		return err
	}
	_, err := io.WriteString(w, s)
	return err
}

// ReadString reads a length-prefixed string with a uint32 length.
func ReadString(r io.Reader) (string, error) {
	n, err := ReadUint32(r)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", nil
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
