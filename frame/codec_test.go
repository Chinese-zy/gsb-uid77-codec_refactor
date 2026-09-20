package frame

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func f(id uint64, v any) Field {
	return Field{ID: id, Name: NameOf(id), Value: v}
}

// goldenCases 是写死的字节对照表：输入字节 -> 期望解析结果。
// 同样的表在 web/codec.test.js 里独立写了一份。
type goldenCase struct {
	name string
	hex  string
	want Frame
	// valid 为 false 表示该字节含截断/非法字段；按 SPEC.md 只要求“安全停下、
	// 保留完整字段”，不要求 encode 能还原原始坏字节。
	valid bool
}

func goldenCases() []goldenCase {
	return []goldenCase{
		{
			name:  "不足两字节",
			hex:   "01",
			want:  Frame{Version: 0, Fields: []Field{}},
			valid: false,
		},
		{
			name:  "空帧只带版本",
			hex:   "0100",
			want:  Frame{Version: 1, Fields: []Field{}},
			valid: true,
		},
		{
			name:  "标量小端",
			hex:   "0200" + "030104000000",
			want:  Frame{Version: 2, Fields: []Field{f(3, uint32(4))}},
			valid: true,
		},
		{
			name:  "版本大小端区分",
			hex:   "0001",
			want:  Frame{Version: 256, Fields: []Field{}},
			valid: true,
		},
		{
			name:  "标量高位字节",
			hex:   "0000" + "0501fcfdfcfc",
			want:  Frame{Version: 0, Fields: []Field{f(5, uint32(0xfcfcfdfc))}},
			valid: true,
		},
		{
			name:  "重复字段保序不去重",
			hex:   "0000" + "010101000000" + "010102000000",
			want:  Frame{0, []Field{f(1, uint32(1)), f(1, uint32(2))}},
			valid: true,
		},
		{
			name:  "LEB128 温度 id 150",
			hex:   "0000" + "96010107000000",
			want:  Frame{0, []Field{f(150, uint32(7))}},
			valid: true,
		},
		{
			name:  "LEB128 多续位",
			hex:   "0000" + "8080080101000000",
			want:  Frame{0, []Field{f(0x20000, uint32(1))}},
			valid: true,
		},
		{
			name:  "成组含两字段",
			hex:   "0100" + "02030c" + "030104000000" + "040105000000",
			want:  Frame{1, []Field{f(2, []Field{f(3, uint32(4)), f(4, uint32(5))})}},
			valid: true,
		},
		{
			name:  "成组后还可接标量",
			hex:   "0100" + "020300" + "030109000000",
			want:  Frame{1, []Field{f(2, []Field{}), f(3, uint32(9))}},
			valid: true,
		},
		{
			name:  "截断标量丢弃当前字段",
			hex:   "0000" + "03010400",
			want:  Frame{0, []Field{}},
			valid: false,
		},
		{
			name:  "悬垂 varint 不炸",
			hex:   "0000" + "9680",
			want:  Frame{0, []Field{}},
			valid: false,
		},
		{
			name:  "未知 kind 停止",
			hex:   "0000" + "0102",
			want:  Frame{0, []Field{}},
			valid: false,
		},
		{
			name:  "成组长度越界",
			hex:   "0000" + "020310" + "030104000000",
			want:  Frame{0, []Field{}},
			valid: false,
		},
		{
			name:  "成组内层截断整体丢弃",
			hex:   "0000" + "020303" + "0301",
			want:  Frame{0, []Field{}},
			valid: false,
		},
	}
}

func TestGolden(t *testing.T) {
	for _, tc := range goldenCases() {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := hex.DecodeString(tc.hex)
			if err != nil {
				t.Fatal(err)
			}
			got := Decode(raw)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("decode 不一致\n got: %#v\nwant: %#v", got, tc.want)
			}
			if tc.valid {
				enc, err := Encode(got)
				if err != nil {
					t.Fatalf("encode: %v", err)
				}
				if hex.EncodeToString(enc) != tc.hex {
					t.Fatalf("encode 不对称\n got: %s\nwant: %s",
						hex.EncodeToString(enc), tc.hex)
				}
			}
		})
	}
}

func TestEncodeRejectsBadValue(t *testing.T) {
	_, err := Encode(Frame{Fields: []Field{f(1, 17)}}) // int 而非 uint32
	if err != ErrValue {
		t.Fatalf("got %v, want ErrValue", err)
	}
}

func TestSampleGolden(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(file), "..", "web", "sample.bin")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := Frame{1, []Field{
		f(2, []Field{
			f(3, uint32(5)),
			f(4, uint32(6)),
			f(5, uint32(7)),
			f(6, uint32(8)),
		}),
		f(1, uint32(1000)),
		f(8, uint32(3)),
		f(150, uint32(7)),
	}}
	got := Decode(raw)
	if !reflect.DeepEqual(got, want) {
		gotJSON, _ := json.Marshal(got)
		wantJSON, _ := json.Marshal(want)
		t.Fatalf("sample 不一致\n got: %s\nwant: %s", gotJSON, wantJSON)
	}
	enc, err := Encode(got)
	if err != nil || !reflect.DeepEqual(enc, raw) {
		t.Fatalf("sample 往返不对称: %v", err)
	}
}

// TestParityWithJS 直接跑页面侧编解码做对拍：同一张字节表两边结果必须一致，
// 且各自重新编码出的字节也必须一致。对不上就让 go test 失败。
func TestParityWithJS(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("未安装 node，跳过前后端对拍")
	}

	type caseIn struct {
		Name string `json:"name"`
		Hex  string `json:"hex"`
	}
	type caseOut struct {
		Name  string          `json:"name"`
		Frame json.RawMessage `json:"frame"`
		Reenc string          `json:"reenc"`
	}

	cases := make([]caseIn, 0, len(goldenCases())+1)
	for _, tc := range goldenCases() {
		cases = append(cases, caseIn{tc.name, tc.hex})
	}
	_, file, _, _ := runtime.Caller(0)
	sample, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "web", "sample.bin"))
	if err != nil {
		t.Fatal(err)
	}
	cases = append(cases, caseIn{"样例文件", hex.EncodeToString(sample)})

	inJSON, _ := json.Marshal(cases)
	bridge := filepath.Join(filepath.Dir(file), "..", "web", "parity_bridge.js")
	cmd := exec.Command("node", bridge)
	cmd.Stdin = bytes.NewReader(inJSON)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("JS 对拍失败: %v\n%s", err, ee.Stderr)
		}
		t.Fatal(err)
	}
	var jsOut []caseOut
	if err := json.Unmarshal(out, &jsOut); err != nil {
		t.Fatalf("解析 JS 结果失败: %v\n%s", err, out)
	}
	if len(jsOut) != len(cases) {
		t.Fatalf("用例数对不上: go=%d js=%d", len(cases), len(jsOut))
	}

	for i, tc := range cases {
		raw, _ := hex.DecodeString(tc.Hex)
		goFrame := Decode(raw)
		goJSON, _ := json.Marshal(goFrame)
		if string(jsOut[i].Frame) != string(goJSON) {
			t.Errorf("用例 %q 解析不一致\n js: %s\n go: %s",
				tc.Name, jsOut[i].Frame, goJSON)
		}
		goEnc, err := Encode(goFrame)
		if err != nil {
			t.Errorf("用例 %q go encode: %v", tc.Name, err)
			continue
		}
		if jsOut[i].Reenc != hex.EncodeToString(goEnc) {
			t.Errorf("用例 %q 重编码不一致\n js: %s\n go: %s",
				tc.Name, jsOut[i].Reenc, hex.EncodeToString(goEnc))
		}
	}
}
