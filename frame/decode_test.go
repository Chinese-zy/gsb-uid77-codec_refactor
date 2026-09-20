package frame

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDecodeSample(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(file), "..", "web", "sample.bin")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := Decode(b)
	if got.Version == 0 || len(got.Fields) == 0 {
		t.Fatalf("empty %#v", got)
	}
}
