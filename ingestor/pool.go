package ingestor

import (
	"strconv"
	"sync"
)

const _size = 1024

type pool struct {
	p *sync.Pool
}

var newPool = func() pool {
	return pool{
		p: &sync.Pool{
			New: func() interface{} {
				return &buffer{bs: make([]byte, 0, _size)}
			},
		},
	}
}

func (p pool) get() *buffer {
	buf := p.p.Get().(*buffer)
	buf.reset()
	buf.pool = p
	return buf
}

func (p pool) put(buf *buffer) {
	p.p.Put(buf)
}

type buffer struct {
	bs   []byte
	pool pool
}

func (b *buffer) appendByte(v byte) {
	b.bs = append(b.bs, v)
}

func (b *buffer) appendString(s string) {
	b.bs = append(b.bs, s...)
}

func (b *buffer) appendInt(i int64) {
	b.bs = strconv.AppendInt(b.bs, i, 10)
}

func (b *buffer) appendUInt(i uint64) {
	b.bs = strconv.AppendUint(b.bs, i, 10)
}

func (b *buffer) appendBool(v bool) {
	b.bs = strconv.AppendBool(b.bs, v)
}

func (b *buffer) appendFloat(f float64, bitSize int) {
	b.bs = strconv.AppendFloat(b.bs, f, 'f', -1, bitSize)
}

func (b *buffer) length() int {
	return len(b.bs)
}

func (b *buffer) capacity() int {
	return cap(b.bs)
}

func (b *buffer) byteSlice() []byte {
	return b.bs
}

func (b *buffer) stringByteSlice() string {
	return string(b.bs)
}

func (b *buffer) reset() {
	b.bs = b.bs[:0]
}

func (b *buffer) write(bs []byte) (int, error) {
	b.bs = append(b.bs, bs...)
	return len(bs), nil
}

func (b *buffer) free() {
	b.pool.put(b)
}
