package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func examplesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolName, err := req.RequireString("tool_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	examples, ok := toolExamples[toolName]
	if !ok {
		available := make([]string, 0, len(toolExamples))
		for k := range toolExamples {
			available = append(available, k)
		}
		sort.Strings(available)
		return mcp.NewToolResultError(fmt.Sprintf("unknown tool %q. Available: %s", toolName, strings.Join(available, ", "))), nil
	}
	return mcp.NewToolResultText(examples), nil
}

var toolExamples = map[string]string{
	"ls": `Example 1: List root directory

Request:
  tool: ls
  arguments: {"path": "/"}

Response:
  file1.txt
  subdir/
  README.md

Example 2: List a subdirectory

Request:
  tool: ls
  arguments: {"path": "/subdir"}

Response:
  nested.txt
  another_dir/

Example 3: Nonexistent directory returns error

Request:
  tool: ls
  arguments: {"path": "/no_such_dir"}

Response (error):
  open /no_such_dir: no such file or directory`,

	"find": `Example 1: Find all .go files recursively

Request:
  tool: find
  arguments: {"pattern": "**/*.go"}

Response:
  main.go
  server/handler.go
  server/routes.go

Example 2: Find files in root only

Request:
  tool: find
  arguments: {"pattern": "*.txt"}

Response:
  file1.txt
  notes.txt

Example 3: Find directories

Request:
  tool: find
  arguments: {"pattern": "src*"}

Response:
  src/`,

	"cat-b": `Example 1: Read full file with line numbers

Request:
  tool: cat-b
  arguments: {"path": "/main.go", "line_range": ":"}

Response:
         1	package main
       2
       3	func main() {
       4		fmt.Println("hello")
       5	}

Example 2: Read a specific line range (lines 3-5, 0-indexed)

Request:
  tool: cat-b
  arguments: {"path": "/main.go", "line_range": "2:5"}

Response:
       3	func main() {
       4		fmt.Println("hello")
       5	}

Example 3: Empty lines are printed without a number

Request:
  tool: cat-b
  arguments: {"path": "/data.txt", "line_range": ":"}

Response:
       1	first line
       2	third line`,

	"file_create": `Example 1: Create a new file

Request:
  tool: file_create
  arguments: {"path": "/hello.txt", "content": "Hello, world!"}

Response:
  ok

Example 2: File already exists returns error

Request:
  tool: file_create
  arguments: {"path": "/hello.txt", "content": "different content"}

Response (error):
  file already exists

Example 3: Create a file in a nested directory (directories are created automatically)

Request:
  tool: file_create
  arguments: {"path": "/src/pkg/util.go", "content": "package pkg"}

Response:
  ok`,

	"mv": `Example 1: Move a file

Request:
  tool: mv
  arguments: {"source": "/old.txt", "dest": "/new.txt"}

Response:
  ok

Example 2: Source not found returns error

Request:
  tool: mv
  arguments: {"source": "/missing.txt", "dest": "/somewhere.txt"}

Response (error):
  source not found

Example 3: Destination already exists returns error

Request:
  tool: mv
  arguments: {"source": "/a.txt", "dest": "/b.txt"}

Response (error):
  destination already exists`,

	"grep": `Example 1: Simple pattern match

Request:
  tool: grep
  arguments: {"pattern": "TODO", "pathspec": "*.go"}

Response:
  main.go:14:// TODO: handle error
  handler.go:7:// TODO: refactor

Example 2: Regex pattern

Request:
  tool: grep
  arguments: {"pattern": "^func ", "pathspec": "**/*.go"}

Response:
  main.go:5:func main() {
  handler.go:12:func handle() error {

Example 3: No match returns empty string

Request:
  tool: grep
  arguments: {"pattern": "NONEXISTENT_PATTERN", "pathspec": "*.go"}

Response:
  (empty string)`,

	"sed": `Example 1: Replace text in matching files

Request:
  tool: sed
  arguments: {"pattern": "old_name", "replacement": "new_name", "pathspec": "**/*.go"}

Response:
  main.go
  handler.go

Example 2: No matches returns empty string

Request:
  tool: sed
  arguments: {"pattern": "nothing_matches_this", "replacement": "x", "pathspec": "*.txt"}

Response:
  (empty string)

Example 3: Replace in a specific file

Request:
  tool: sed
  arguments: {"pattern": "v1.0", "replacement": "v2.0", "pathspec": "config.yaml"}

Response:
  config.yaml`,

	"diff": `Example 1: Two different files

Request:
  tool: diff
  arguments: {"path1": "/original.txt", "path2": "/modified.txt"}

Response:
  --- /original.txt
  +++ /modified.txt
  @@ -1,3 +1,3 @@
   line one
  -line two
  +line two changed
   line three

Example 2: Identical files return empty string

Request:
  tool: diff
  arguments: {"path1": "/file.txt", "path2": "/copy.txt"}

Response:
  (empty string)

Example 3: Adding lines at the end

Request:
  tool: diff
  arguments: {"path1": "/a.txt", "path2": "/b.txt"}

Response:
  --- /a.txt
  +++ /b.txt
  @@ -2,3 +2,5 @@
   second line
   third line
  +fourth line
  +fifth line`,

	"diff_strings": `Example 1: Two different strings

Request:
  tool: diff_strings
  arguments: {"string1": "hello\nworld", "string2": "hello\nearth"}

Response:
  --- string1
  +++ string2
  @@ -1,2 +1,2 @@
   hello
  -world
  +earth

Example 2: Identical strings return empty string

Request:
  tool: diff_strings
  arguments: {"string1": "same", "string2": "same"}

Response:
  (empty string)

Example 3: Multi-line diff

Request:
  tool: diff_strings
  arguments: {"string1": "a\nb\nc", "string2": "a\nx\ny\nc"}

Response:
  --- string1
  +++ string2
  @@ -1,3 +1,4 @@
   a
  -b
  +x
  +y
   c`,

	"patch": `Example 1: Apply a unified diff

Request:
  tool: patch
  arguments: {"patch": "--- /file.txt\n+++ /file.txt\n@@ -1,3 +1,3 @@\n line one\n-old line\n+new line\n line three"}

Response:
  ok

Example 2: Invalid patch format returns error

Request:
  tool: patch
  arguments: {"patch": "not a valid patch"}

Response (error):
  invalid patch format

Example 3: Adding lines via patch

Request:
  tool: patch
  arguments: {"patch": "--- /file.txt\n+++ /file.txt\n@@ -1,2 +1,4 @@\n first\n second\n+third\n+fourth"}

Response:
  ok`,

	"rm": `Example 1: Delete a file

Request:
  tool: rm
  arguments: {"path": "/old_file.txt"}

Response:
  ok

Example 2: Delete a directory recursively

Request:
  tool: rm
  arguments: {"path": "/old_directory"}

Response:
  ok

Example 3: Nonexistent path also returns ok

Request:
  tool: rm
  arguments: {"path": "/already_gone.txt"}

Response:
  ok`,

	"stat": `Example 1: Stat a file

Request:
  tool: stat
  arguments: {"path": "/main.go"}

Response:
  Type: file
  Size: 234, 234 B
  Permissions: read,write
  Owner: alice(uid=1000)
  Group: alice(gid=1000)
  Access: 2025-04-25T10:30:00
  Modify: 2025-04-25T09:15:00
  Change: 2025-04-25T09:15:00
  Birth: 2025-04-20T08:00:00

Example 2: Stat a directory

Request:
  tool: stat
  arguments: {"path": "/src"}

Response:
  Type: directory
  Size: 4096, 4.0 KB
  Permissions: read,write,execute
  Owner: alice(uid=1000)
  Group: alice(gid=1000)
  Access: 2025-04-25T10:30:00
  Modify: 2025-04-25T09:15:00
  Change: 2025-04-25T09:15:00
  Birth: 2025-04-20T08:00:00

Example 3: Nonexistent path returns error

Request:
  tool: stat
  arguments: {"path": "/no_such_file"}

Response (error):
  stat /no_such_file: no such file or directory`,

	"tasks_list": `Example 1: List existing tasks

Request:
  tool: tasks_list
  arguments: {}

Response:
  1. [X] Design the API
  2. [_] Implement features
  2.1 [X] Add config loading
  2.2 [ ] Add error handling
  3. [ ] Write documentation

Example 2: No tasks returns empty string

Request:
  tool: tasks_list
  arguments: {}

Response:
  (empty string)

Example 3: Tasks with mixed statuses

Request:
  tool: tasks_list
  arguments: {}

Response:
  1. [ ] Set up CI pipeline
  2. [_] Write unit tests
  3. [ ] Add integration tests`,

	"task_create": `Example 1: Create a top-level task

Request:
  tool: task_create
  arguments: {"description": "Set up the database schema"}

Response:
  Created ID: 1

Example 2: Create a sub-task under task 2

Request:
  tool: task_create
  arguments: {"description": "Write migration script", "parent": "2"}

Response:
  Created ID: 2.1

Example 3: Multiple sub-tasks get auto-incremented IDs

Request:
  tool: task_create
  arguments: {"description": "Add seed data", "parent": "2"}

Response:
  Created ID: 2.2`,

	"task_set_status": `Example 1: Mark a task as in progress

Request:
  tool: task_set_status
  arguments: {"ID": "1", "status": "in_progress"}

Response:
  true

Example 2: Mark a task as completed

Request:
  tool: task_set_status
  arguments: {"ID": "2.1", "status": "completed"}

Response:
  true

Example 3: Invalid status returns error

Request:
  tool: task_set_status
  arguments: {"ID": "1", "status": "done"}

Response (error):
  invalid status: done`,

	"task_delete": `Example 1: Delete a task

Request:
  tool: task_delete
  arguments: {"ID": "3"}

Response:
  true

Example 2: Deleting a parent also removes its children

Request:
  tool: task_delete
  arguments: {"ID": "2"}

Response:
  true
  (task 2, 2.1, 2.2 are all removed)

Example 3: Deleting a non-existent task still succeeds

Request:
  tool: task_delete
  arguments: {"ID": "99"}

Response:
  true`,

	"tasks_clear": `Example 1: Clear all tasks

Request:
  tool: tasks_clear
  arguments: {}

Response:
  true

Example 2: Clear when already empty

Request:
  tool: tasks_clear
  arguments: {}

Response:
  true

Example 3: After clearing, tasks_list returns empty

Request:
  tool: tasks_clear
  arguments: {}

Response:
  true
  (subsequent call to tasks_list will return an empty string)`,

	"w3m-dump": `Example 1: Fetch a webpage

Request:
  tool: w3m-dump
  arguments: {"url": "https://example.com/docs/api"}

Response:
  # API Documentation

  ## Getting Started
  Welcome to the API. This endpoint allows you to...

Example 2: Unsupported URL scheme returns error

Request:
  tool: w3m-dump
  arguments: {"url": "ftp://files.example.com/data"}

Response (error):
  only http and https URLs are supported

Example 3: Very long pages get truncated

Request:
  tool: w3m-dump
  arguments: {"url": "https://example.com/very-long-article"}

Response:
  # PAGE TOO LONG - PARTIAL OUTPUT

  # Article Title
  First part of the article content...`,

	"online_search": `Example 1: Search with results

Request:
  tool: online_search
  arguments: {"search_query": "golang mcp server"}

Response:
  Title: mark3labs/mcp-go
  Url: https://github.com/mark3labs/mcp-go
  Description:
  An MCP server implementation in Go.

  Title: Building MCP Servers
  Url: https://modelcontextprotocol.io
  Description:
  Learn how to build Model Context Protocol servers.

Example 2: No results found

Request:
  tool: online_search
  arguments: {"search_query": "xyzzyqwerty12345nonexistent"}

Response:
  No results found

Example 3: Multi-word query

Request:
  tool: online_search
  arguments: {"search_query": "how to parse json in go"}

Response:
  Title: encoding/json - Go Documentation
  Url: https://pkg.go.dev/encoding/json
  Description:
  Package json implements encoding and decoding of JSON.`,

	"available_commands": `Example 1: List configured commands

Request:
  tool: available_commands
  arguments: {}

Response:
  Command: build
  Arguments: target
  Description: Build the project

  Command: test
  Arguments: no arguments are taken, invoke without arguments
  Description: Run tests

Example 2: No commands configured returns empty string

Request:
  tool: available_commands
  arguments: {}

Response:
  (empty string)

Example 3: Single command with multiple arguments

Request:
  tool: available_commands
  arguments: {}

Response:
  Command: deploy
  Arguments: environment
  Arguments: version
  Description: Deploy to the specified environment`,

	"run_command": `Example 1: Run a command with arguments

Request:
  tool: run_command
  arguments: {"name": "build", "arguments": ["release"], "timeout": 60}

Response:
  Build complete: release binary at ./bin/app

Example 2: Run a command without arguments

Request:
  tool: run_command
  arguments: {"name": "test", "arguments": [], "timeout": 120}

Response:
  PASS
  ok  github.com/example/project  0.123s

Example 3: Unknown command returns error

Request:
  tool: run_command
  arguments: {"name": "deploy", "arguments": [], "timeout": 30}

Response (error):
  unknown command: deploy`,

	"examples": `Example 1: Get examples for a tool

Request:
  tool: examples
  arguments: {"tool_name": "ls"}

Response:
  Example 1: List root directory
  ...

Example 2: Unknown tool returns error with available names

Request:
  tool: examples
  arguments: {"tool_name": "nonexistent"}

Response (error):
  unknown tool "nonexistent". Available: cat-b, diff, ...

Example 3: Get examples for task tools

Request:
  tool: examples
  arguments: {"tool_name": "task_create"}

Response:
  Example 1: Create a top-level task
  ...`,
}
