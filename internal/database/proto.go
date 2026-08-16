package database

import (
	"encoding/binary"
	"unicode/utf8"
)

// protoField is one decoded protobuf key/value.
type protoField struct {
	num     int
	wire    int
	varint  uint64
	bytes   []byte
	fixed32 uint32
	fixed64 uint64
}

func decodeProtoFields(buf []byte) []protoField {
	var fields []protoField
	i := 0
	for i < len(buf) {
		key, n := binary.Uvarint(buf[i:])
		if n <= 0 {
			return fields
		}
		i += n
		num := int(key >> 3)
		wire := int(key & 7)
		if num <= 0 {
			return fields
		}
		f := protoField{num: num, wire: wire}
		switch wire {
		case 0:
			v, n2 := binary.Uvarint(buf[i:])
			if n2 <= 0 {
				return fields
			}
			i += n2
			f.varint = v
		case 1:
			if i+8 > len(buf) {
				return fields
			}
			f.fixed64 = binary.LittleEndian.Uint64(buf[i : i+8])
			i += 8
		case 2:
			ln, n2 := binary.Uvarint(buf[i:])
			if n2 <= 0 || int(ln) < 0 || i+n2+int(ln) > len(buf) {
				return fields
			}
			i += n2
			f.bytes = buf[i : i+int(ln)]
			i += int(ln)
		case 5:
			if i+4 > len(buf) {
				return fields
			}
			f.fixed32 = binary.LittleEndian.Uint32(buf[i : i+4])
			i += 4
		default:
			return fields
		}
		fields = append(fields, f)
	}
	return fields
}

func protoStringField(fields []protoField, num int) string {
	for _, f := range fields {
		if f.num == num && f.wire == 2 && utf8.Valid(f.bytes) {
			return string(f.bytes)
		}
	}
	return ""
}

func protoIntField(fields []protoField, num int) (int, bool) {
	for _, f := range fields {
		if f.num == num && f.wire == 0 {
			return int(f.varint), true
		}
	}
	return 0, false
}

func protoBoolField(fields []protoField, num int) bool {
	v, ok := protoIntField(fields, num)
	return ok && v != 0
}

func protoRepeatedBytes(fields []protoField, num int) [][]byte {
	var out [][]byte
	for _, f := range fields {
		if f.num == num && f.wire == 2 {
			out = append(out, f.bytes)
		}
	}
	return out
}

func encodeProtoVarint(v uint64) []byte {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], v)
	return buf[:n]
}

func encodeProtoKey(num, wire int) []byte {
	return encodeProtoVarint(uint64(num<<3 | wire))
}

func encodeProtoString(num int, s string) []byte {
	return encodeProtoBytes(num, []byte(s))
}

func encodeProtoBytes(num int, payload []byte) []byte {
	out := encodeProtoKey(num, 2)
	out = append(out, encodeProtoVarint(uint64(len(payload)))...)
	return append(out, payload...)
}

func encodeProtoVarintField(num int, v uint64) []byte {
	out := encodeProtoKey(num, 0)
	return append(out, encodeProtoVarint(v)...)
}
