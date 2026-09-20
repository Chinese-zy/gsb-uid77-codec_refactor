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

func Decode(b []byte) Frame {
	out := Frame{}
	if len(b) < 2 {
		return out
	}
	out.Version = binary.BigEndian.Uint16(b[:2])
	o := 2
	bag := map[uint64]Field{}
	for o < len(b) {
		key, o2 := readVarint(b, o)
		o = o2
		if o >= len(b) {
			break
		}
		kind := b[o]
		o++
		switch kind {
		case 1:
			if o+4 > len(b) {
				return emit(out, bag)
			}
			val := binary.BigEndian.Uint32(b[o : o+4])
			o += 4
			bag[key] = Field{Name: nameOf(key), Value: val}
		case 3:
			n, o2 := readVarint(b, o)
			o = o2
			if o+int(n) > len(b) {
				return emit(out, bag)
			}
			inner := decodeInner(b[o : o+int(n)])
			o += int(n)
			bag[key] = Field{Name: nameOf(key), Value: inner}
		default:
			return emit(out, bag)
		}
	}
	return emit(out, bag)
}

func decodeInner(b []byte) []Field {
	bag := map[uint64]Field{}
	o := 0
	for o < len(b) {
		key, o2 := readVarint(b, o)
		o = o2
		if o >= len(b) {
			break
		}
		kind := b[o]
		o++
		if kind != 1 || o+4 > len(b) {
			break
		}
		val := binary.BigEndian.Uint32(b[o : o+4])
		o += 4
		bag[key] = Field{Name: nameOf(key), Value: val}
	}
	fields := make([]Field, 0, len(bag))
	for _, field := range bag {
		fields = append(fields, field)
	}
	return fields
}

func emit(out Frame, bag map[uint64]Field) Frame {
	for _, field := range bag {
		out.Fields = append(out.Fields, field)
	}
	return out
}

func readVarint(b []byte, o int) (uint64, int) {
	if o >= len(b) {
		return 0, o
	}
	first := b[o]
	if first < 128 {
		return uint64(first), o + 1
	}
	n := int(first & 0x7f)
	o++
	var v uint64
	for i := 0; i < n && o < len(b); i++ {
		v = (v << 8) | uint64(b[o])
		o++
	}
	return v, o
}
