package record

import (
	"fmt"
	"io"
	"time"
	"unsafe"
)

var (
	_ = unsafe.Sizeof(0)
	_ = io.ReadFull
	_ = time.Now()
)

// GenCodeSize returns the size of the gencode marshalled byte slice
func (d *Meta) GenCodeSize() (s int) {
	s += 34
	return
}

// GenCodeMarshal gencode marshalls Meta into the given byte array, or a new one if its too small.
func (d *Meta) GenCodeMarshal(buf []byte) ([]byte, error) {
	size := d.GenCodeSize()
	{
		if cap(buf) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{

		buf[0+0] = byte(d.Created >> 0)

		buf[1+0] = byte(d.Created >> 8)

		buf[2+0] = byte(d.Created >> 16)

		buf[3+0] = byte(d.Created >> 24)

		buf[4+0] = byte(d.Created >> 32)

		buf[5+0] = byte(d.Created >> 40)

		buf[6+0] = byte(d.Created >> 48)

		buf[7+0] = byte(d.Created >> 56)

	}
	{

		buf[0+8] = byte(d.Modified >> 0)

		buf[1+8] = byte(d.Modified >> 8)

		buf[2+8] = byte(d.Modified >> 16)

		buf[3+8] = byte(d.Modified >> 24)

		buf[4+8] = byte(d.Modified >> 32)

		buf[5+8] = byte(d.Modified >> 40)

		buf[6+8] = byte(d.Modified >> 48)

		buf[7+8] = byte(d.Modified >> 56)

	}
	{

		buf[0+16] = byte(d.Expires >> 0)

		buf[1+16] = byte(d.Expires >> 8)

		buf[2+16] = byte(d.Expires >> 16)

		buf[3+16] = byte(d.Expires >> 24)

		buf[4+16] = byte(d.Expires >> 32)

		buf[5+16] = byte(d.Expires >> 40)

		buf[6+16] = byte(d.Expires >> 48)

		buf[7+16] = byte(d.Expires >> 56)

	}
	{

		buf[0+24] = byte(d.Deleted >> 0)

		buf[1+24] = byte(d.Deleted >> 8)

		buf[2+24] = byte(d.Deleted >> 16)

		buf[3+24] = byte(d.Deleted >> 24)

		buf[4+24] = byte(d.Deleted >> 32)

		buf[5+24] = byte(d.Deleted >> 40)

		buf[6+24] = byte(d.Deleted >> 48)

		buf[7+24] = byte(d.Deleted >> 56)

	}
	{
		if d.secret {
			buf[32] = 1
		} else {
			buf[32] = 0
		}
	}
	{
		if d.cronjewel {
			buf[33] = 1
		} else {
			buf[33] = 0
		}
	}
	return buf[:i+34], nil
}

// GenCodeUnmarshal gencode unmarshalls Meta and returns the bytes read.
func (d *Meta) GenCodeUnmarshal(buf []byte) (uint64, error) {
	if len(buf) < d.GenCodeSize() {
		return 0, fmt.Errorf("insufficient data: got %d out of %d bytes", len(buf), d.GenCodeSize())
	}

	i := uint64(0)

	{

		d.Created = 0 | (int64(buf[0+0]) << 0) | (int64(buf[1+0]) << 8) | (int64(buf[2+0]) << 16) | (int64(buf[3+0]) << 24) | (int64(buf[4+0]) << 32) | (int64(buf[5+0]) << 40) | (int64(buf[6+0]) << 48) | (int64(buf[7+0]) << 56)

	}
	{

		d.Modified = 0 | (int64(buf[0+8]) << 0) | (int64(buf[1+8]) << 8) | (int64(buf[2+8]) << 16) | (int64(buf[3+8]) << 24) | (int64(buf[4+8]) << 32) | (int64(buf[5+8]) << 40) | (int64(buf[6+8]) << 48) | (int64(buf[7+8]) << 56)

	}
	{

		d.Expires = 0 | (int64(buf[0+16]) << 0) | (int64(buf[1+16]) << 8) | (int64(buf[2+16]) << 16) | (int64(buf[3+16]) << 24) | (int64(buf[4+16]) << 32) | (int64(buf[5+16]) << 40) | (int64(buf[6+16]) << 48) | (int64(buf[7+16]) << 56)

	}
	{

		d.Deleted = 0 | (int64(buf[0+24]) << 0) | (int64(buf[1+24]) << 8) | (int64(buf[2+24]) << 16) | (int64(buf[3+24]) << 24) | (int64(buf[4+24]) << 32) | (int64(buf[5+24]) << 40) | (int64(buf[6+24]) << 48) | (int64(buf[7+24]) << 56)

	}
	{
		d.secret = buf[32] == 1
	}
	{
		d.cronjewel = buf[33] == 1
	}
	return i + 34, nil
}
