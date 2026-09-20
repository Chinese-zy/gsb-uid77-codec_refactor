package frame

import (
	"encoding/binary"
	"errors"
	"math"
)

// SPEC.md 是前后端共同遵守的线格式：
// 小端多字节整数 + LEB128 变长整数，字段严格按出现顺序保留。

const (
	KindScalar byte = 1 // u32le
	KindGroup  byte = 3 // varint(byte_len) + 内层字段
)

type Field struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Value any    `json:"value"` // uint32 或 []Field
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

var ErrValue = errors.New("frame: unsupported field value")

func NameOf(id uint64) string {
	if n, ok := names[id]; ok {
		return n
	}
	return "未知"
}

func Decode(b []byte) Frame {
	out := Frame{Fields: []Field{}}
	if len(b) < 2 {
		return out
	}
	out.Version = binary.LittleEndian.Uint16(b[:2])
	fields, _ := decodeFields(b, 2, len(b))
	out.Fields = fields
	return out
}

// decodeFields 解析 [start,end) 内的字段序列。
// 第二个返回值为 false 表示途中遇到截断或非法字段；此时已解析字段仍返回，
// 调用方应按 SPEC.md 第 4 节丢弃当前不完整字段并停止。
func decodeFields(b []byte, start, end int) ([]Field, bool) {
	fields := make([]Field, 0)
	o := start
	for o < end {
		id, next, ok := readVarint(b, o, end)
		if !ok || next >= end {
			return fields, false
		}
		o = next
		kind := b[o]
		o++
		switch kind {
		case KindScalar:
			if o+4 > end {
				return fields, false
			}
			val := binary.LittleEndian.Uint32(b[o : o+4])
			o += 4
			fields = append(fields, Field{ID: id, Name: NameOf(id), Value: val})
		case KindGroup:
			n, next2, ok := readVarint(b, o, end)
			if !ok {
				return fields, false
			}
			o = next2
			innerEnd := o + int(n)
			if innerEnd > end || n > math.MaxInt32 {
				return fields, false
			}
			inner, ok := decodeFields(b, o, innerEnd)
			if !ok {
				return fields, false
			}
			fields = append(fields, Field{ID: id, Name: NameOf(id), Value: inner})
			o = innerEnd
		default:
			return fields, false
		}
	}
	return fields, true
}

func Encode(f Frame) ([]byte, error) {
	out := make([]byte, 0, 8)
	out = binary.LittleEndian.AppendUint16(out, f.Version)
	var err error
	out, err = encodeFields(out, f.Fields)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func encodeFields(out []byte, fields []Field) ([]byte, error) {
	for _, field := range fields {
		out = appendVarint(out, field.ID)
		switch v := field.Value.(type) {
		case uint32:
			out = append(out, KindScalar)
			out = binary.LittleEndian.AppendUint32(out, v)
		case []Field:
			inner := make([]byte, 0, 8)
			var err error
			inner, err = encodeFields(inner, v)
			if err != nil {
				return nil, err
			}
			out = append(out, KindGroup)
			out = appendVarint(out, uint64(len(inner)))
			out = append(out, inner...)
		default:
			return nil, ErrValue
		}
	}
	return out, nil
}

// readVarint 读 LEB128。续位为 0 才结束；最长 10 字节（uint64 上限）。
func readVarint(b []byte, start, end int) (uint64, int, bool) {
	var v uint64
	o := start
	for i := 0; ; i++ {
		if o >= end || i >= 10 {
			return 0, start, false
		}
		c := b[o]
		o++
		if i == 9 && c&0x7f > 0x01 {
			return 0, start, false
		}
		v |= uint64(c&0x7f) << uint(7*i)
		if c&0x80 == 0 {
			return v, o, true
		}
	}
}

func appendVarint(out []byte, v uint64) []byte {
	for v >= 0x80 {
		out = append(out, byte(v)|0x80)
		v >>= 7
	}
	return append(out, byte(v))
}
