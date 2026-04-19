package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	stdio := flag.Bool("stdio", false, "use stdio transport")
	http := flag.Bool("http", false, "use HTTP transport")
	addr := flag.String("address", "localhost:8080", "HTTP listen address")
	ignore := flag.String("ignore", "", "regex pattern to ignore files/directories")
	flag.Parse()

	if *ignore != "" {
		re, err := regexp.Compile(*ignore)
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

	s := server.NewMCPServer("nixdevkit", "0.1.0",
		server.WithToolCapabilities(true),
		server.WithToolFilter(func(ctx context.Context, tools []mcp.Tool) []mcp.Tool {
			if !isReadonly() {
				return tools
			}
			hidden := map[string]bool{
				"create": true, "edit": true, "sed": true, "patch": true, "rm": true,
			}
			var filtered []mcp.Tool
			for _, t := range tools {
				if !hidden[t.Name] {
					filtered = append(filtered, t)
				}
			}
			return filtered
		}),
	)

	s.AddTool(mcp.NewTool("ls",
		mcp.WithDescription("List directory content"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Directory path"),
		),
	), lsHandler)

	s.AddTool(mcp.NewTool("find",
		mcp.WithDescription("Find files"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("Glob expression"),
		),
	), findHandler)

	s.AddTool(mcp.NewTool("read",
		mcp.WithDescription("Read a file"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("File path"),
		),
		mcp.WithString("line_range",
			mcp.Required(),
			mcp.Description("Line range [from]:[to], 0-indexed"),
		),
	), readHandler)

	s.AddTool(mcp.NewTool("create",
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

	s.AddTool(mcp.NewTool("edit",
		mcp.WithDescription("Replace a file section"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("File path"),
		),
		mcp.WithString("line_range",
			mcp.Required(),
			mcp.Description("Line range [from]:[to], 0-indexed"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("New content"),
		),
	), editHandler)

	s.AddTool(mcp.NewTool("grep",
		mcp.WithDescription("Print lines matching pattern"),
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

	s.AddTool(mcp.NewTool("tasks_create",
		mcp.WithDescription("Append a task to the task list"),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Task description"),
		),
		mcp.WithString("parent",
			mcp.Description("ID of the parent task, optional"),
		),
	), tasksCreateHandler)

	s.AddTool(mcp.NewTool("tasks_set_status",
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

	s.AddTool(mcp.NewTool("tasks_delete",
		mcp.WithDescription("Delete a task"),
		mcp.WithString("ID",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
	), tasksDeleteHandler)

	s.AddTool(mcp.NewTool("tasks_clear",
		mcp.WithDescription("Clear all tasks"),
	), tasksClearHandler)

	s.AddTool(mcp.NewTool("available_commands",
		mcp.WithDescription("List available commands"),
	), availableCommandsHandler)

	s.AddTool(mcp.NewTool("exec_command",
		mcp.WithDescription("Run the command from available_commands"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the command to run"),
		),
		mcp.WithArray("arguments",
			mcp.Required(),
			mcp.Description("Array of strings to pass to the command line"),
			mcp.WithStringItems(),
		),
		mcp.WithNumber("timeout",
			mcp.Required(),
			mcp.Description("Timeout in seconds"),
		),
	), execCommandHandler)

	if *http && !*stdio {
		srv := server.NewStreamableHTTPServer(s)
		fmt.Fprintf(os.Stderr, "nixdevkit: HTTP on %s, root=%s\n", *addr, rootDir)
		if err := srv.Start(*addr); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		_ = stdio
		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
