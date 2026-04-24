# nixdevkit

A minimal MCP server exposing Unix-inspired file tools. Designed for low token usage and sandboxed file access.

## Usage

```
./nixdevkit [--stdio|--http] [--address host:port] [--ignore pattern] [rootdirectory]
```

All paths are virtual — `/` maps to the root directory. Path traversal is blocked.

- Default transport is stdio.
- `--http` starts a streamable HTTP server on the given `--address` (default `localhost:8080`).
- `--ignore` accepts a regular expression; files and directories whose relative path matches are hidden from all tools. Traversal tools (`find`, `grep`, `sed`) skip entire matched directories.
- If no root directory is given, the current working directory is used.

## Tools

### `create` — Create a file

| Argument | Description |
|----------|-------------|
| `path` | File path |
| `content` | File content |

Creates a new file. Errors if the file already exists.

### `ls` — List directory content

| Argument | Description |
|----------|-------------|
| `path` | Directory path |

Returns newline-separated entries with full relative paths. Directories end with `/`.

### `find` — Find files

| Argument | Description |
|----------|-------------|
| `pattern` | Glob expression |

Recursively walks the root. Supports `*` and `**` (globstar) syntax. Directories end with `/`.

### `read` — Read a file

| Argument | Description |
|----------|-------------|
| `path` | File path |
| `line_range` | Line range `[from]:[to]`, 0-indexed |

Returns file content. Use `":"` for the full file, `"2:"` from line 2 onward, `":3"` for lines 1–3. Invalid numbers default to full range.

### `replace_range` — Replace a file section

| Argument | Description |
|----------|-------------|
| `path` | File path |
| `line_range` | Line range `[from]:[to]`, 0-indexed |
| `content` | New content |

Replaces the specified line range with the new content. Use `"0:0"` to prepend, empty content to delete lines.

### `grep` — Print lines matching pattern

| Argument | Description |
|----------|-------------|
| `pattern` | Regular expression |
| `pathspec` | Glob expression for file names |

Output format: `filepath:linenumber:linecontent`. Supports `**` globstar. Line numbers are 1-indexed.

### `sed` — Search and replace in files

| Argument | Description |
|----------|-------------|
| `pattern` | Regular expression |
| `replacement` | Replacement string |
| `pathspec` | Glob expression for file names |

In-place match and replace (no capturing groups). Returns list of changed files. Supports `**` globstar.

### `diff` — Compare files, output unified diff

| Argument | Description |
|----------|-------------|
| `path1` | First file path |
| `path2` | Second file path |

Output is compatible with the `patch` tool. Returns empty string if files are identical.

### `patch` — Apply a unified diff

| Argument | Description |
|----------|-------------|
| `patch` | Unified diff to apply |

Applies the diff to the target file in-place. Designed to consume output from the `diff` tool.

### `rm` — Delete a file or a directory

| Argument | Description |
|----------|-------------|
| `path` | Path to delete |

Recursive delete (`rm -rf`). Returns ok for nonexistent paths.

### `mv` — Move files

| Argument | Description |
|----------|-------------|
| `source` | File path |
| `dest` | File path |

Moves a file or directory. Fails if destination already exists or source not found.

### `stat` — Various info on files and directories

| Argument | Description |
|----------|-------------|
| `path` | File or directory path |

Returns:

```
Type: [file|directory]
Size: [bytes], [human readable]
Permissions: [read|write|execute]
Owner: [username](uid=[uid])
Group: [groupname](gid=[gid])
Access: [ISO8601 timestamp]
Modify: [ISO8601 timestamp]
Change: [ISO8601 timestamp]
Birth: [ISO8601 timestamp]
```

Birth time uses `statx` when available, falls back to change time otherwise. Permissions are relative to the current user.

### `w3m-dump` — Fetch a webpage text

| Argument | Description |
|----------|-------------|
| `url` | URL to fetch |

Fetches a webpage, extracts the main readable content using Mozilla Readability, and converts it to Markdown. Only `http` and `https` URLs are supported. Connect timeout is 5 seconds, read timeout is 20 seconds. Response body is limited to 5 MB. If the resulting Markdown exceeds 200 KB, it is truncated and prefixed with `# PAGE TOO LONG - PARTIAL OUTPUT`.

### `online_search` — Search topic online

| Argument | Description |
|----------|-------------|
| `search_query` | Search query string |

Searches DuckDuckGo and returns results with title, URL, and description. Returns `No results found` if nothing matches. Example output:

```
Title: This is a page
Url: http://....
Description:
The page is about being
a page. The description text can be
multiline.

Title: This is another page
Url: http://....
Description: A short description
```

### `available_commands` — List available commands

No arguments.

Lists all user-defined commands from the configuration, including their descriptions and expected arguments. Example output:

```
Command: build
Arguments: target

Command: test
Arguments: no arguments are taken, invoke without arguments
Description: Run tests

Command: run
Description: Run the main executable; target_folder is the directory to work with, config_file is the reference configuration to use.
```

### `exec_command` — Run the command

| Argument | Description |
|----------|-------------|
| `name` | Name of the command to run |
| `arguments` | Array of strings to pass to the command line |
| `timeout` | Timeout in seconds |

Validates the command name and argument count against the configuration, sanitizes input, and executes the command. Arguments are only accepted when the command defines an `arguments` list; passing arguments to a command that takes none is an error. Stdout and stderr are merged and returned untouched. On timeout, the process is sent SIGTERM, then SIGKILL after 5 seconds. If a timeout occurs, the output is prefixed with `Command timed out. Partial output.`.

For example, with `build_cmdline=make` and `build_arguments=target`, calling `exec_command` with `name="build"` and `arguments=["clean"]` executes `make clean`.

## Task Management

Tasks are stored in `[root]/.nixdevkit/tasks.txt`. Each task has a system-assigned hierarchical ID, a status, and a description.

Status markers:

| Status | Marker |
|--------|--------|
| `created` | `[ ]` |
| `in_progress` | `[_]` |
| `completed` | `[X]` |

Example file content:

```
1. [X] Design the API
2. [_] Implement features
2.1 [X] Add config loading
2.2 [ ] Add error handling
3. [ ] Write documentation
```

### `tasks_list` — List all tasks

No arguments.

Returns the content of the tasks file.

### `tasks_create` — Append a task to the task list

| Argument | Required | Description |
|----------|----------|-------------|
| `description` | Yes | Task description |
| `parent` | No | ID of the parent task |

Returns the assigned ID in the format `ID: $ID`. When `parent` is provided, the new task becomes a child (e.g. `parent="2"` → new ID `2.1`).

### `tasks_set_status` — Change status of a task

| Argument | Description |
|----------|-------------|
| `ID` | Task ID |
| `status` | One of: `created`, `in_progress`, `completed` |

### `tasks_delete` — Delete a task

| Argument | Description |
|----------|-------------|
| `ID` | Task ID |

Deletes the task and all its children.

### `tasks_clear` — Clear all tasks

No arguments.

## Configuration

`nixdevkit` reads an INI-style configuration file at `[root]/.nixdevkit/config.ini`. The entire `.nixdevkit` directory is invisible to all MCP tools — it cannot be listed, read, created, edited, or deleted through the server.

The configuration is re-read on every request, so changes take effect without restarting the server.

### `nixdevkit-config` — Manage the configuration file

```
./nixdevkit-config <get|set> <namespace.key> [value]
./nixdevkit-config <root> <get|set> <namespace.key> [value]
```

Examples:

```
./nixdevkit-config set core.readonly true
./nixdevkit-config get core.readonly
./nixdevkit-config /path/to/project set core.readonly yes
```

### `core.readonly`

When set to `true` (or `1` / `yes`), the write tools are hidden from the server:

- `create`
- `replace_range`
- `sed`
- `patch`
- `rm`
- `mv`

Read-only tools (`ls`, `find`, `read`, `grep`, `diff`, `stat`, `w3m-dump`, `online_search`, `available_commands`) remain available.

### `commands` — User-defined commands

The `commands` section lets you define named commands that can be listed and executed through the `available_commands` and `exec_command` tools. Each command requires a `cmdline` and can optionally have a `description` and an `arguments` list.

| Key | Required | Description |
|-----|----------|-------------|
| `commands.list` | Yes | Comma-separated list of command names |
| `commands.<name>_cmdline` | Yes | The command line to execute |
| `commands.<name>_description` | No | Human-readable description of the command |
| `commands.<name>_arguments` | No | Comma-separated list of argument names the command accepts |

Example configuration:

```
./nixdevkit-config set commands.list build,test,run
./nixdevkit-config set commands.build_cmdline "make"
./nixdevkit-config set commands.build_arguments "target"
./nixdevkit-config set commands.test_cmdline "make test"
./nixdevkit-config set commands.test_description "Run tests"
```
