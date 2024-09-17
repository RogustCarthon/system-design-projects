package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func main() {
	for {
		body := map[string]any{"message": "Hello, World!"}
		b, _ := json.Marshal(body)
		req, err := http.NewRequest("POST", "http://localhost:8080/hi", bytes.NewBuffer(b))
		if err != nil {
			panic(err)
		}
		req.Close = true
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		_ = resp.Body.Close()
		slog.Info("got response", "resp", resp.Status)
		time.Sleep(time.Second)
	}
}
