package ingestor

import (
	"bytes"
	"testing"
)

func Test_bufferMethods(t *testing.T) {
	b := &buffer{}

	b.appendByte(byte('a'))
	rec, exp := len(b.bs), 1
	if rec != exp {
		t.Errorf("appendByte byte string length %v, expected %v", rec, exp)
	}
	b.appendString("b")
	rec, exp = len(b.bs), 2
	if rec != exp {
		t.Errorf("appendString byte string length %v, expected %v", rec, exp)
	}
	b.appendInt(int64(1))
	rec, exp = len(b.bs), 3
	if rec != exp {
		t.Errorf("appendInt byte string length %v, expected %v", rec, exp)
	}
	b.appendUInt(uint64(1))
	rec, exp = len(b.bs), 4
	if rec != exp {
		t.Errorf("appendUInt byte string length %v, expected %v", rec, exp)
	}
	b.appendBool(true)
	rec, exp = len(b.bs), 8
	if rec != exp {
		t.Errorf("appendBool byte string length %v, expected %v", rec, exp)
	}
	b.appendFloat(0.5, 32)
	rec, exp = len(b.bs), 11
	if rec != exp {
		t.Errorf("appendFloat byte string length %v, expected %v", rec, exp)
	}
	out := b.length()
	exp = 11
	if out != exp {
		t.Errorf("length returning wrong length %v, expected %v", out, exp)
	}
	out = b.capacity()
	exp = 16
	if out != exp {
		t.Errorf("capacity returning incorrect cap %v, expected %v", out, exp)
	}
	bts := b.byteSlice()
	if !bytes.Equal(bts, []byte{97, 98, 49, 49, 116, 114, 117, 101, 48, 46, 53}) {
		t.Errorf("byteSlice returning incorrect byte array %v", bts)
	}
	sbts := b.stringByteSlice()
	if sbts != "ab11true0.5" {
		t.Errorf("stringByteSlice returning incorrect slice %v", sbts)
	}
}
