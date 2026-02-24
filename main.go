package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CmdProbe runs the probe command
func CmdProbe(args []string) int {
	fs := flag.NewFlagSet("probe", flag.ExitOnError)
	outdir := fs.String("outdir", "./out", "输出目录")
	ratePerMin := fs.Float64("rate-per-min", 45.0, "本地节流上限")
	maxBatch := fs.Int("max-batch", 50, "domains.check 单次最大域名数")
	tldMode := fs.String("tld-mode", "all", "遍历哪些 TLD: all/api-registerable-only/mainstream-only")
	cacheTTLHours := fs.Float64("cache-ttl-hours", 24.0, "缓存 TTL（小时）")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: probe <word> [options]")
		return 1
	}

	word := fs.Arg(0)
	outdirAbs, _ := filepath.Abs(*outdir)

	var tldModeVal TldMode
	switch *tldMode {
	case "api-registerable-only":
		tldModeVal = TldModeApiRegisterableOnly
	case "mainstream-only":
		tldModeVal = TldModeMainstreamOnly
	default:
		tldModeVal = TldModeAll
	}

	ctx := context.Background()
	outputChan := make(chan interface{}, 100)

	go func() {
		err := ProbeWord(ctx, word, outdirAbs, tldModeVal, *cacheTTLHours, *ratePerMin, *maxBatch, outputChan)
		if err != nil {
			errObj := map[string]string{"type": "error", "message": err.Error()}
			data, _ := json.Marshal(errObj)
			fmt.Fprintln(os.Stderr, string(data))
			os.Exit(1)
		}
		close(outputChan)
	}()

	for obj := range outputChan {
		data, _ := json.Marshal(obj)
		fmt.Println(string(data))
	}

	return 0
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string
	Data  string
}

// String returns the SSE formatted string
func (e SSEEvent) String() string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", e.Event, e.Data)
}

// CmdServe runs the HTTP server
func CmdServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	host := fs.String("host", "127.0.0.1", "监听地址")
	port := fs.Int("port", 8000, "监听端口")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	http.HandleFunc("/probe/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/probe/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 || parts[0] == "" {
			http.Error(w, "Missing word parameter", http.StatusBadRequest)
			return
		}
		word := parts[0]

		ratePerMin := 45.0
		if rpm := r.URL.Query().Get("rate_per_min"); rpm != "" {
			if val, err := strconv.ParseFloat(rpm, 64); err == nil {
				ratePerMin = val
			}
		}

		tldMode := TldModeAll
		if tm := r.URL.Query().Get("tld_mode"); tm != "" {
			switch tm {
			case "api-registerable-only":
				tldMode = TldModeApiRegisterableOnly
			case "mainstream-only":
				tldMode = TldModeMainstreamOnly
			}
		}

		outdir := filepath.Join("./out", word)
		outdirAbs, _ := filepath.Abs(outdir)

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ctx := r.Context()
		outputChan := make(chan interface{}, 100)

		go func() {
			_ = ProbeWord(ctx, word, outdirAbs, tldMode, 24.0, ratePerMin, 50, outputChan)
			close(outputChan)
		}()

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		for obj := range outputChan {
			data, _ := json.Marshal(obj)
			event := "message"
			if m, ok := obj.(map[string]interface{}); ok {
				if t, ok := m["type"].(string); ok {
					event = t
				}
			} else if p, ok := obj.(ProgressEvent); ok {
				event = p.Type
			} else if s, ok := obj.(SummaryEvent); ok {
				event = s.Type
			}

			sse := SSEEvent{Event: event, Data: string(data)}
			fmt.Fprint(w, sse.String())
			flusher.Flush()
		}
	})

	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Printf("Server starting on http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
	return 0
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: namecheap-domain-probe <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  probe <word>    探测一个单词的所有 TLD 组合")
		fmt.Println("  serve           启动 SSE 服务")
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var exitCode int
	switch cmd {
	case "probe":
		exitCode = CmdProbe(args)
	case "serve":
		exitCode = CmdServe(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}

	os.Exit(exitCode)
}
