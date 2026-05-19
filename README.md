# dahmer

A small, fast CLI that finds — and kills — whatever is holding your TCP ports. Cross-platform (macOS, Linux, Windows), built in Go with [Cobra](https://github.com/spf13/cobra) and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

```text
$ dahmer 3000
╭─────────────────────────────╮
│ dahmer :: show port 3000    │
│                             │
│ PID       48213             │
│ Command   node              │
│ User      ashu              │
│                             │
│ run `dahmer 3000 kill` to   │
│ terminate it                │
╰─────────────────────────────╯
```

## Install

### Go

```sh
go install github.com/ashutosh-swain-git/dahmer@latest
```

Requires Go 1.21+. The binary is placed in `$(go env GOPATH)/bin` — make sure that's on your `PATH`.

### Build from source

```sh
git clone https://github.com/ASHUTOSH-SWAIN-GIT/dahmer.git
cd dahmer
go build -o dahmer .
./dahmer 3000
```

## Usage

### Inspect a port

```sh
dahmer 3000
```

Shows the PID, command, and user holding port 3000. If nothing is listening, you'll see a friendly "port is free" message.

### Inspect multiple ports

```sh
dahmer 3000 3001 8080
dahmer 3000-3010
```

You can mix individual ports and ranges. Results render as a table; free ports are dimmed.

### Kill a process

```sh
dahmer 3000 kill
```

Sends `SIGTERM` (on Windows, `taskkill /PID`). Give the process a moment to shut down cleanly.

### Force-kill

```sh
dahmer 3000 kill -f
```

Sends `SIGKILL` (on Windows, `taskkill /PID /F`). Use this when `dahmer 3000 kill` didn't take.

### Kill a range

```sh
dahmer 3000-3010 kill
dahmer 3000 3001 8080 kill -f
```

Kills every process in the range concurrently and reports a summary.

### List every listening port

```sh
dahmer ls
```

Table of every TCP socket currently in `LISTEN` state.

### Help

```sh
dahmer --help
dahmer kill --help
```

## Command reference

| Command                       | What it does                                          |
| ----------------------------- | ----------------------------------------------------- |
| `dahmer <port>`               | Show what's holding the port                          |
| `dahmer <port> <port> ...`    | Show multiple ports                                   |
| `dahmer <lo>-<hi>`            | Show a port range                                     |
| `dahmer <port> kill`          | Terminate the process (SIGTERM)                       |
| `dahmer <port> kill -f`       | Force-terminate (SIGKILL)                             |
| `dahmer ls`                   | List every listening TCP port                         |
| `dahmer --help`               | Show usage                                            |

## How it works

| Platform        | Lookup                | Kill                            |
| --------------- | --------------------- | ------------------------------- |
| macOS / Linux   | `lsof -iTCP -sTCP:LISTEN` | `syscall.Kill` (SIGTERM / SIGKILL) |
| Windows         | `netstat -ano` + `tasklist` | `taskkill /PID <pid> [/F]`     |

The platform split lives behind Go build tags in `internal/port/` — no runtime branching, no CGO, no external Go dependencies for the OS bits.

## License

MIT — see [LICENSE](./LICENSE).
