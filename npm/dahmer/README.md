# dahmer

Find and kill the process bound to a TCP port. macOS, Linux, and Windows.

```sh
npm install -g dahmer

dahmer 3000              # show what's holding port 3000
dahmer 3000 3001 8080    # show multiple
dahmer 3000-3010         # range
dahmer 3000 kill         # SIGTERM (taskkill on Windows)
dahmer 3000 kill -f      # SIGKILL (taskkill /F on Windows)
dahmer ls                # list every listening TCP port
```

The right prebuilt binary for your platform is installed automatically via npm
`optionalDependencies` (no `postinstall` download).
