package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"nixdevkit/internal/indexer"
)

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}
	rootDir, _ = filepath.Abs(rootDir)

	idx := indexer.NewIndexer(rootDir)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		idx.Stop()
		os.Exit(0)
	}()

	start := time.Now()
	fmt.Fprintf(os.Stderr, "[INFO] Starting indexer for %s\n", rootDir)
	if err := idx.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "[INFO] Indexer ready in %s\n", time.Since(start).Round(time.Millisecond))
	fmt.Println("idle")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		cmd := parts[0]

		switch cmd {
		case "health":
			fmt.Println(idx.HandleHealth())

		case "reindex":
			t := time.Now()
			if err := idx.HandleReindex(); err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] reindex: %v\n", err)
				fmt.Printf("error: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "[INFO] Reindex completed in %s\n", time.Since(t).Round(time.Millisecond))
				fmt.Println("ok")
			}

		case "retrieve":
			if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
				fmt.Println("error: missing query text")
				continue
			}
			query := parts[1]
			t := time.Now()
			results, err := idx.HandleRetrieve(query)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] retrieve: %v\n", err)
				fmt.Printf("error: %v\n", err)
				continue
			}
			data, _ := json.Marshal(results)
			fmt.Println(string(data))
			fmt.Fprintf(os.Stderr, "[INFO] retrieve completed in %s (%d results)\n", time.Since(t).Round(time.Millisecond), len(results))

		default:
			fmt.Printf("error: unknown command: %s\n", cmd)
		}
	}

	idx.Stop()
}
