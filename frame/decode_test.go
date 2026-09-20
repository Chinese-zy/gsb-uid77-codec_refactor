package frame

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type goldenCase struct {
	Name   string          `json:"name"`
	Hex    string          `json:"hex"`
	Expect json.RawMessage `json:"expect"`
}

func loadGolden(t *testing.T) []goldenCase {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(file), "..", "testdata", "golden.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		Cases []goldenCase `json:"cases"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	return golden.Cases
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	s = strings.ReplaceAll(s, " ", "")
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// 页面和接口共用 testdata/golden.json 这一份字节对照表，对不上就别交。
func TestDecodeGolden(t *testing.T) {
	for _, c := range loadGolden(t) {
		t.Run(c.Name, func(t *testing.T) {
			gotJSON, err := json.Marshal(Decode(mustHex(t, c.Hex)))
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if err := json.Unmarshal(gotJSON, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(c.Expect, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("字节 %q\n页面/接口应对上 %s\n实际解出   %s", c.Hex, c.Expect, gotJSON)
			}
		})
	}
}

// 演示页用的 sample.bin 必须和对照表里的同名字节保持一致。
func TestDecodeSample(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "web", "sample.bin"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range loadGolden(t) {
		if c.Name != "sample.bin 完整报文" {
			continue
		}
		if want := mustHex(t, c.Hex); !reflect.DeepEqual(raw, want) {
			t.Fatalf("sample.bin 是 %x，对照表里是 %x，两边得是同一份字节", raw, want)
		}
		return
	}
	t.Fatal("对照表里找不到 sample.bin 完整报文")
}
