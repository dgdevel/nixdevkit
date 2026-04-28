# nixdevkit

A minimal MCP server exposing Unix-inspired file tools. Designed for low token usage and sandboxed file access.

## Usage

```
./nixdevkit [--stdio|--http] [--address host:port] [--ignore pattern] [--show tools] [--hide tools] [rootdirectory]
```

All paths are virtual — `/` maps to the root directory. Path traversal is blocked.

- Default transport is stdio.
- `--http` starts a streamable HTTP server on the given `--address` (default `localhost:8080`).
- `--ignore` accepts a regular expression; files and directories whose relative path matches are hidden from all tools. Traversal tools (`ls`, `grep`, `sed`) skip entire matched directories.
- `--show` accepts a comma-separated list of tool names to expose (whitelist). Only the listed tools are available. Mutually exclusive with `--hide`. Proxied tools (from `mcps.yml`) are always included regardless of this flag.
- `--hide` accepts a comma-separated list of tool names to hide (blacklist). All other tools remain available. Mutually exclusive with `--show`. Proxied tools are always included regardless of this flag.
- If no root directory is given, the current working directory is used.

## Tools

### `file_create` — Create a file

| Argument | Description |
|----------|-------------|
| `path` | File path |
| `content` | File content |

Creates a new file. Errors if the file already exists.

### `ls` — List directory content

| Argument | Description |
|----------|-------------|
| `pathspec` | Glob expression for file names |

Recursively walks the root matching the glob pattern. Supports `*` and `**` (globstar) syntax. Directories end with `/`.

### `fread` — Read file content

| Argument | Description |
|----------|-------------|
| `path` | File to read |
| `line_range` | Line range `[from]:[to]`, 1-indexed |

Reads a file and outputs the raw content in blocks, with no transformation (no line numbers, no tab/trailing-space visualization). Output is split into blocks of 30 lines (configurable via `core.fread_block_size`). Each block is preceded by a header:

```
----- $path - line from X to Y -----
```

At the end, an EOF marker is emitted:

```
----- $path - EOF -----
```

### `mv` — Move files

| Argument | Description |
|----------|-------------|
| `source` | File path |
| `dest` | File path |

Moves a file or directory. Fails if destination already exists or source not found.

### `grep` — Print lines matching pattern with context

| Argument | Description |
|----------|-------------|
| `pattern` | Regular expression |
| `pathspec` | Glob expression for file names |

Output format: `filepath:linenumber:linecontent`. Shows 3 context lines before and after each match. Non-adjacent match groups are separated by `--`. Supports `**` globstar. Line numbers are 1-indexed. Output is limited to 500 content lines.

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

### `diff_strings` — Helper that format unified diff from two strings

| Argument | Description |
|----------|-------------|
| `string1` | First string |
| `string2` | Second string |

Like `diff` but operates on raw strings instead of files. Output is compatible with the `patch` tool. Returns empty string if strings are identical.

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

### `run_command` — Run the command

| Argument | Description |
|----------|-------------|
| `name` | Name of the command to run |
| `arguments` | Array of strings to pass to the command line |
| `timeout` | Timeout in seconds |

Validates the command name and argument count against the configuration, sanitizes input, and executes the command. Arguments are only accepted when the command defines an `arguments` list; passing arguments to a command that takes none is an error. Stdout and stderr are merged and returned untouched. On timeout, the process is sent SIGTERM, then SIGKILL after 5 seconds. If a timeout occurs, the output is prefixed with `Command timed out. Partial output.`.

For example, with `build_cmdline=make` and `build_arguments=target`, calling `run_command` with `name="build"` and `arguments=["clean"]` executes `make clean`.

### `examples` — Show usage examples for a tool

| Argument | Description |
|----------|-------------|
| `tool_name` | Name of the tool to get examples for |

Returns at least 3 examples of request/response pairs for the given tool. If the tool name is unknown, returns an error listing all available tool names.

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

### `task_create` — Append a task to the task list

| Argument | Required | Description |
|----------|----------|-------------|
| `description` | Yes | Task description |
| `parent` | No | ID of the parent task |

Returns the assigned ID in the format `Created ID: $ID`. When `parent` is provided, the new task becomes a child (e.g. `parent="2"` → new ID `2.1`).

### `task_set_status` — Change status of a task

| Argument | Description |
|----------|-------------|
| `ID` | Task ID |
| `status` | One of: `created`, `in_progress`, `completed` |

### `task_delete` — Delete a task

| Argument | Description |
|----------|-------------|
| `ID` | Task ID |

Deletes the task and all its children.

### `tasks_clear` — Clear all tasks

No arguments.

## Configuration

`nixdevkit` reads configuration from two locations, merged with local overriding global:

1. **Global**: `$XDG_CONFIG_HOME/nixdevkit/config.ini` (or `$HOME/.config/nixdevkit/config.ini` if `XDG_CONFIG_HOME` is unset)
2. **Local**: `[root]/.nixdevkit/config.ini`

Both the global and local `.nixdevkit` directories are invisible to all MCP tools — they cannot be listed, read, created, edited, or deleted through the server.

The configuration is re-read on every request, so changes take effect without restarting the server.

### `nixdevkit-config` — Manage the configuration file

```
./nixdevkit-config [--global] <get|set> <namespace.key> [value]
./nixdevkit-config <root> <get|set> <namespace.key> [value]
```

With `--global`, operations target the global configuration file instead of the local one. The `--global` flag cannot be combined with a root directory argument.

Examples:

```
./nixdevkit-config set core.readonly true
./nixdevkit-config --global set core.readonly yes
./nixdevkit-config get core.readonly
./nixdevkit-config /path/to/project set core.readonly yes
```

### `core.readonly`

When set to `true` (or `1` / `yes`), the write tools are hidden from the server:

- `file_create`
- `sed`
- `patch`
- `rm`
- `mv`

### `core.fread_block_size`

Block size (number of lines) for the `fread` tool. Default is `30`.

### `commands` — User-defined commands

The `commands` section lets you define named commands that can be listed and executed through the `available_commands` and `run_command` tools. Each command requires a `cmdline` and can optionally have a `description` and an `arguments` list.

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

### `mcps` — Upstream MCP server proxying

`nixdevkit` can proxy tools from upstream MCP servers, making them available as if they were built-in. Configuration is loaded from both global (`$XDG_CONFIG_HOME/nixdevkit/mcps.yml`) and local (`[root]/.nixdevkit/mcps.yml`), merged with local overriding global.

```yaml
mcps:
  myserver:
    url: http://localhost:9001/mcp
    headers:
      Authorization: Bearer token123
    tools:
      search:
        rename: my_search
        description: Search my database
        arguments:
          query:
            description: The search query
      get_item:
        keep_as_is: true
```

| Field | Required | Description |
|-------|----------|-------------|
| `mcps.<name>.url` | Yes | URL of the upstream MCP server (streamable HTTP transport) |
| `mcps.<name>.headers` | No | HTTP headers to send with each request |
| `mcps.<name>.tools` | Yes | Map of upstream tool names to their configuration |

For each tool entry:

| Field | Required | Description |
|-------|----------|-------------|
| `rename` | No | New name for the proxied tool |
| `description` | No | Override the tool description |
| `arguments` | No | Map of argument names to `{rename, description}` overrides |
| `keep_as_is` | No | If `true`, pass the tool through unchanged (no rename/description overrides) |

Only tools explicitly listed in the `tools` map are proxied. Other tools from the upstream server are ignored.

Proxied tools are excluded from `--show`/`--hide` filtering — they are always visible.

## Code Indexer

nixdevkit includes an optional code indexer that provides semantic code search using local embedding and reranking models powered by llama.cpp. It runs as a child process of the main nixdevkit server.

### Setup

```
./nixdevkit-setup-indexer [--global] [rootdirectory]
```

With `--global`, llama.cpp binaries and models are stored in the global config directory (`$XDG_CONFIG_HOME/nixdevkit/`), and the `[llama]` configuration is written there. This is recommended so that all projects share the same binaries and models. A root directory cannot be specified when using `--global`.

Downloads llama.cpp (CPU-only x86_64), an embedding model (nomic-embed-text-v1.5 Q4_K_M) and a reranking model (bge-reranker-v2-m3 Q4_K_M), then writes the configuration to the config file.

The index storage (vector database) is always local to each project at `[root]/.nixdevkit/index/`, since it is project-specific.

### First-time indexing

The initial index can take several minutes depending on project size. It is recommended to warm up the index before launching the MCP server:

```
echo "reindex" | ./nixdevkit-indexer [rootdirectory]
```

Wait for the `ok` response, then start the MCP server with `--enable-indexer`. Subsequent startups will only index changed files (incremental via mtime tracking).

### `relevant_code` — Semantic code search

| Argument | Description |
|----------|-------------|
| `prompt` | Description of the code you are looking for |

Requires `--enable-indexer`. Returns one result per line in the format:

```
file_path:line_start-line_end:language:chunk_type:signature
```

Use `fread` with the reported line range to read the actual code. Returns an empty string if the indexer is not ready or no results are found.

### Configuration

The indexer reads the `[llama]` section from the merged global+local configuration:

| Key | Description |
|-----|-------------|
| `llama.path` | Path to `llama-server` binary (may include extra flags) |
| `llama.embedder` | HuggingFace repo ID for the embedding model |
| `llama.reranker` | HuggingFace repo ID for the reranking model (not required when `llama.reranker_enabled` is `false`) |
| `llama.search_count` | Number of documents retrieved from the vector database (default: `50`) |
| `llama.result_count` | Number of final results returned after reranking (default: `10`) |
| `llama.reranker_enabled` | Set to `false`, `0`, `no`, `disabled`, or `off` to skip the reranker entirely. When disabled, results are scored by vector similarity only and the reranker server is not started (default: `true`) |
