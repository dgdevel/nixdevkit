package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"nixdevkit/internal/cfg"
	"nixdevkit/internal/indexer"
	"nixdevkit/internal/memory"

	"github.com/mark3labs/mcp-go/mcp"
)

var (
	memStore    *memory.Memory
	memEmbedder *indexer.LlamaServer
	memCancel   context.CancelFunc
	memMu       sync.Mutex
)

func startMemory(rootDir string) error {
	config := cfg.MergedRead(rootDir)
	llamaCfg := config["llama"]
	if llamaCfg == nil {
		return fmt.Errorf("missing [llama] config section for memory")
	}

	llamaPath := llamaCfg["path"]
	if llamaPath == "" {
		return fmt.Errorf("missing llama.path config for memory")
	}

	embedderRepo := llamaCfg["embedder"]
	if embedderRepo == "" {
		return fmt.Errorf("missing llama.embedder config for memory")
	}

	memCtx, cancel := context.WithCancel(context.Background())
	memCancel = cancel

	t := time.Now()
	fmt.Fprintf(os.Stderr, "[INFO] Starting memory embedder server...\n")
	srv, err := indexer.StartServer(memCtx, llamaPath, embedderRepo, "--embedding")
	if err != nil {
		cancel()
		return fmt.Errorf("starting memory embedder: %w", err)
	}
	memEmbedder = srv
	fmt.Fprintf(os.Stderr, "[INFO] Memory embedder on port %d (started in %s)\n", srv.Port(), time.Since(t).Round(time.Millisecond))

	embedFn := func(ctx context.Context, text string) ([]float32, error) {
		return srv.GetEmbeddingOpenAI(ctx, text)
	}

	memDir := memory.DirPath(rootDir)
	memStore, err = memory.NewMemory(memCtx, memDir, embedFn)
	if err != nil {
		srv.Stop()
		cancel()
		return fmt.Errorf("initializing memory: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[INFO] Memory store ready (%d facts)\n", memStore.Count())
	return nil
}

func stopMemory() {
	if memCancel != nil {
		memCancel()
	}
	if memEmbedder != nil {
		memEmbedder.Stop()
	}
}

func memoryPutHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fact, err := req.RequireString("fact")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	memMu.Lock()
	defer memMu.Unlock()

	if memStore == nil {
		return mcp.NewToolResultError("memory not initialized"), nil
	}

	if err := memStore.Put(ctx, fact); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to store fact: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Fact memorized (%d total facts)", memStore.Count())), nil
}

func relevantMemoryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt, err := req.RequireString("prompt")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	memMu.Lock()
	defer memMu.Unlock()

	if memStore == nil {
		return mcp.NewToolResultError("memory not initialized"), nil
	}

	facts, err := memStore.Retrieve(ctx, prompt, 10)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to retrieve: %v", err)), nil
	}

	if len(facts) == 0 {
		return mcp.NewToolResultText("No relevant memories found."), nil
	}

	data, _ := json.MarshalIndent(facts, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
