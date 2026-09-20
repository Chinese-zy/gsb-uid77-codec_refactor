package main

import (
	"encoding/json"
	"net/http"
	"os"

	"codec/frame"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/frame", func(w http.ResponseWriter, r *http.Request) {
		raw, err := os.ReadFile("web/sample.bin")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(frame.Decode(raw))
	})
	mux.Handle("/", http.FileServer(http.Dir("web")))
	_ = http.ListenAndServe("127.0.0.1:8777", mux)
}
