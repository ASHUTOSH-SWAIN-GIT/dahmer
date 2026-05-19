# dahmer

Find and kill the process bound to a TCP port. Cross-platform (macOS, Linux, Windows), built with [Cobra](https://github.com/spf13/cobra) + [Bubble Tea](https://github.com/charmbracelet/bubbletea).

```text
dahmer 3000              show what's holding port 3000
dahmer 3000 3001 8080    show multiple ports
dahmer 3000-3010         show a range
dahmer 3000 kill         SIGTERM the process on 3000
dahmer 3000-3010 kill    SIGTERM every process in the range
dahmer 3000 kill -f      SIGKILL (taskkill /F on Windows)
dahmer ls                list every listening TCP port
```

## Install

```sh
go install github.com/ashutosh-swain-git/dahmer@latest
```

Requires Go 1.21+. The binary lands in `$(go env GOPATH)/bin` — make sure that's on your `PATH`.

## How it works

- **macOS / Linux** — uses `lsof` to enumerate listeners and `syscall.Kill` (SIGTERM / SIGKILL) to terminate.
- **Windows** — uses `netstat -ano` to find LISTENING TCP sockets, `tasklist` to resolve PIDs to image names, and `taskkill /PID <pid> [/F]` to terminate.

## License

MIT — see [LICENSE](./LICENSE).
