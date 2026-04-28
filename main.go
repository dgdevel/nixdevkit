package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"nixdevkit/internal/mcps"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var (
		stdio          bool
		http           bool
		addr           string
		ignore         string
		showTools      string
		hideTools      string
		enableIndexer  bool
	)
	{
		stdioF := flag.Bool("stdio", false, "use stdio transport")
		httpF := flag.Bool("http", false, "use HTTP transport")
		addrF := flag.String("address", "localhost:8080", "HTTP listen address")
		ignoreF := flag.String("ignore", "", "regex pattern to ignore files/directories")
		showF := flag.String("show", "", "comma-separated whitelist of tool names (mutually exclusive with -hide)")
		hideF := flag.String("hide", "", "comma-separated blacklist of tool names (mutually exclusive with -show)")
		indexerF := flag.Bool("enable-indexer", false, "start code indexer subprocess")
		flag.Parse()
		stdio, http, addr, ignore, showTools, hideTools, enableIndexer = *stdioF, *httpF, *addrF, *ignoreF, *showF, *hideF, *indexerF
	}

	if showTools != "" && hideTools != "" {
		fmt.Fprintln(os.Stderr, "nixdevkit: --show and --hide are mutually exclusive")
		os.Exit(1)
	}

	if ignore != "" {
		re, err := regexp.Compile(ignore)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nixdevkit: invalid ignore pattern: %v\n", err)
			os.Exit(1)
		}
		ignoreRe = re
	}

	args := flag.Args()
	if len(args) > 0 {
		rootDir = args[0]
	} else {
		rootDir, _ = os.Getwd()
	}
	rootDir, _ = filepath.Abs(rootDir)

	proxiedTools := map[string]bool{}

	s := server.NewMCPServer("nixdevkit", "0.1.0",
		server.WithToolCapabilities(true),
		server.WithToolFilter(func(ctx context.Context, tools []mcp.Tool) []mcp.Tool {
			readonlyHidden := map[string]bool{
				"file_create": true, "sed": true, "patch": true, "rm": true, "mv": true,
			}

			var showSet map[string]bool
			var hideSet map[string]bool
			if showTools != "" {
				showSet = make(map[string]bool)
				for _, n := range strings.Split(showTools, ",") {
					showSet[strings.TrimSpace(n)] = true
				}
			}
			if hideTools != "" {
				hideSet = make(map[string]bool)
				for _, n := range strings.Split(hideTools, ",") {
					hideSet[strings.TrimSpace(n)] = true
				}
			}

			var filtered []mcp.Tool
			for _, t := range tools {
				if proxiedTools[t.Name] {
					filtered = append(filtered, t)
					continue
				}
				if isReadonly() && readonlyHidden[t.Name] {
					continue
				}
				if showSet != nil && !showSet[t.Name] {
					continue
				}
				if hideSet != nil && hideSet[t.Name] {
					continue
				}
				filtered = append(filtered, t)
			}
			return filtered
		}),
	)

	s.AddTool(mcp.NewTool("ls",
		mcp.WithDescription("List directory content"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("Glob expression"),
		),
	), lsHandler)

	s.AddTool(mcp.NewTool("cat-b",
		mcp.WithDescription("Read a file with line numbers (like `cat -b`, max 500 lines, → = tab, · = trailing space)"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("File path"),
		),
		mcp.WithString("line_range",
			mcp.Required(),
			mcp.Description("Line range [from]:[to], 0-indexed"),
		),
	), catbHandler)

	s.AddTool(mcp.NewTool("fread",
		mcp.WithDescription("Read file content"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("File to read"),
		),
		mcp.WithString("line_range",
			mcp.Required(),
			mcp.Description("Line range [from]:[to], 1-indexed"),
		),
	), freadHandler)

	s.AddTool(mcp.NewTool("file_create",
		mcp.WithDescription("Create a file"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("File path"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("File content"),
		),
	), createHandler)

	s.AddTool(mcp.NewTool("mv",
		mcp.WithDescription("Move files"),
		mcp.WithString("source",
			mcp.Required(),
			mcp.Description("File path"),
		),
		mcp.WithString("dest",
			mcp.Required(),
			mcp.Description("File path"),
		),
	), mvHandler)

	s.AddTool(mcp.NewTool("grep",
		mcp.WithDescription("Print lines matching pattern with 3 context lines (→ = tab, · = trailing space)"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("Regular expression"),
		),
		mcp.WithString("pathspec",
			mcp.Required(),
			mcp.Description("Glob expression for file names"),
		),
	), grepHandler)

	s.AddTool(mcp.NewTool("sed",
		mcp.WithDescription("Search and replace in files"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("Regular expression"),
		),
		mcp.WithString("replacement",
			mcp.Required(),
			mcp.Description("Replacement string"),
		),
		mcp.WithString("pathspec",
			mcp.Required(),
			mcp.Description("Glob expression for file names"),
		),
	), sedHandler)

	s.AddTool(mcp.NewTool("diff",
		mcp.WithDescription("Compare files, output unified diff"),
		mcp.WithString("path1",
			mcp.Required(),
			mcp.Description("First file path"),
		),
		mcp.WithString("path2",
			mcp.Required(),
			mcp.Description("Second file path"),
		),
	), diffHandler)

	s.AddTool(mcp.NewTool("diff_strings",
		mcp.WithDescription("Unified diff from two strings"),
		mcp.WithString("string1",
			mcp.Required(),
			mcp.Description("First string"),
		),
		mcp.WithString("string2",
			mcp.Required(),
			mcp.Description("Second string"),
		),
	), diffStringsHandler)

	s.AddTool(mcp.NewTool("patch",
		mcp.WithDescription("Apply a unified diff"),
		mcp.WithString("patch",
			mcp.Required(),
			mcp.Description("Unified diff to apply"),
		),
	), patchHandler)

	s.AddTool(mcp.NewTool("rm",
		mcp.WithDescription("Delete a file or a directory"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to delete"),
		),
	), rmHandler)

	s.AddTool(mcp.NewTool("stat",
		mcp.WithDescription("Various info on files and directories"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("File or directory path"),
		),
	), statHandler)

	s.AddTool(mcp.NewTool("tasks_list",
		mcp.WithDescription("List of tasks prefixed by ID and state ([ ] created, [_] in progress, [X] completed)"),
	), tasksListHandler)

	s.AddTool(mcp.NewTool("task_create",
		mcp.WithDescription("Append a task to the task list"),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Task description"),
		),
		mcp.WithString("parent",
			mcp.Description("ID of the parent task, optional"),
		),
	), tasksCreateHandler)

	s.AddTool(mcp.NewTool("task_set_status",
		mcp.WithDescription("Change status of a task"),
		mcp.WithString("ID",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
		mcp.WithString("status",
			mcp.Required(),
			mcp.Description("One of: created, in_progress, completed"),
		),
	), tasksSetStatusHandler)

	s.AddTool(mcp.NewTool("task_delete",
		mcp.WithDescription("Delete a task"),
		mcp.WithString("ID",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
	), tasksDeleteHandler)

	s.AddTool(mcp.NewTool("tasks_clear",
		mcp.WithDescription("Clear all tasks"),
	), tasksClearHandler)

	s.AddTool(mcp.NewTool("w3m-dump",
		mcp.WithDescription("Fetch a webpage text (like `w3m -dump`)"),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("URL to fetch"),
		),
	), w3mdumpHandler)

	s.AddTool(mcp.NewTool("online_search",
		mcp.WithDescription("Search topic online"),
		mcp.WithString("search_query",
			mcp.Required(),
			mcp.Description("Search query string"),
		),
	), onlineSearchHandler)

	s.AddTool(mcp.NewTool("examples",
		mcp.WithDescription("Show usage examples for a tool"),
		mcp.WithString("tool_name",
			mcp.Required(),
			mcp.Description("Name of the tool to get examples for"),
		),
	), examplesHandler)

	s.AddTool(mcp.NewTool("available_commands",
		mcp.WithDescription("List available commands"),
	), availableCommandsHandler)

	s.AddTool(mcp.NewTool("run_command",
		mcp.WithDescription("Run the command from available_commands"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the command to run"),
		),
		mcp.WithArray("arguments",
			mcp.Description("Array of strings to pass to the command line"),
			mcp.WithStringItems(),
		),
		mcp.WithNumber("timeout",
			mcp.Required(),
			mcp.Description("Timeout in seconds"),
		),
	), runCommandHandler)

	if enableIndexer {
		if err := startIndexer(rootDir); err != nil {
			fmt.Fprintf(os.Stderr, "nixdevkit: indexer: %v\n", err)
		} else {
			s.AddTool(mcp.NewTool("relevant_code",
				mcp.WithDescription("Find code relevant to a prompt using semantic search and reranking. Returns one result per line in the format: file_path:line_start-line_end:language:chunk_type:signature. Use cat-b to read the actual code at those lines."),
				mcp.WithString("prompt",
					mcp.Required(),
					mcp.Description("Description of the code you are looking for"),
				),
			), relevantCodeHandler)
		}
	}

	mcpsCfg, err := mcps.LoadConfig(mcps.ConfigPath(rootDir))
	if err != nil {
		fmt.Fprintf(os.Stderr, "nixdevkit: mcps config: %v\n", err)
		os.Exit(1)
	}
	if mcpsCfg != nil {
		proxiedNames, err := mcps.RegisterProxiedTools(context.Background(), s, mcpsCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nixdevkit: mcps: %v\n", err)
			os.Exit(1)
		}
		for _, n := range proxiedNames {
			proxiedTools[n] = true
		}
		fmt.Fprintf(os.Stderr, "nixdevkit: loaded %d upstream MCP servers\n", len(mcpsCfg.MCPS))
	}

	if http && !stdio {
		srv := server.NewStreamableHTTPServer(s)
		fmt.Fprintf(os.Stderr, "nixdevkit: HTTP on %s, root=%s\n", addr, rootDir)
		if err := srv.Start(addr); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
