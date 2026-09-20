package frame

import "encoding/binary"

type Field struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type Frame struct {
	Version uint16  `json:"version"`
	Fields  []Field `json:"fields"`
}

var names = map[uint64]string{
	1:   "电压",
	2:   "成组",
	3:   "湿度",
	4:   "气压",
	5:   "风速",
	6:   "云量",
	8:   "计数",
	150: "温度",
}

func nameOf(id uint64) string {
	if n, ok := names[id]; ok {
		return n
	}
	return "未知"
}

// Decode 按 SPEC.md 解一帧：小端、LEB128、字段保持上线顺序。
func Decode(b []byte) Frame {
	out := Frame{Fields: []Field{}}
	if len(b) < 2 {
		return out
	}
	out.Version = binary.LittleEndian.Uint16(b[:2])
	out.Fields = readFields(b, 2, len(b))
	return out
}

func readFields(b []byte, o, end int) []Field {
	fields := []Field{}
	for o < end {
		var key uint64
		key, o = readVarint(b, o, end)
		if o >= end {
			break
		}
		kind := b[o]
		o++
		switch kind {
		case 1:
			if o+4 > end {
				return fields
			}
			val := binary.LittleEndian.Uint32(b[o : o+4])
			o += 4
			fields = append(fields, Field{Name: nameOf(key), Value: val})
		case 3:
			var n uint64
			n, o = readVarint(b, o, end)
			if n > uint64(end-o) {
				return fields
			}
			inner := readFields(b, o, o+int(n))
			o += int(n)
			fields = append(fields, Field{Name: nameOf(key), Value: inner})
		default:
			return fields
		}
	}
	return fields
}

// readVarint 读 LEB128：低 7 位有效，最高位 1 表示还有后续字节。
func readVarint(b []byte, o, end int) (uint64, int) {
	var v uint64
	var shift uint
	for o < end {
		c := b[o]
		o++
		v |= uint64(c&0x7f) << shift
		if c&0x80 == 0 {
			break
		}
		shift += 7
	}
	return v, o
}
