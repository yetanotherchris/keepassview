# keepassview

A read-only KeePass (.kdbx) database viewer that runs a local web server and opens the vault in your browser. It does not edit, create, or write to the database in any way.

## Build

Requires Go 1.25+.

```
git clone https://github.com/yetanotherchris/keepassview
cd keepassview
go build -o keepassview .
```

### Using Task

If you have [Task](https://taskfile.dev) installed:

| Command | Description |
|---|---|
| `task build` | Compile Tailwind CSS then build the binary |
| `task build-no-css` | Build without regenerating CSS |
| `task css` | Regenerate CSS only (requires `build/tailwindcss` standalone binary) |
| `task vet` | Run `go vet` |
| `task tidy` | Run `go mod tidy` |

`task build` depends on the [Tailwind CSS standalone CLI](https://tailwindcss.com/blog/standalone-cli) placed at `build/tailwindcss` (or `build/tailwindcss.exe` on Windows). Use `task build-no-css` to skip that step if the CSS is already up to date.

## Setup

Run once to configure the database path, password mode, and idle timeout:

```
./keepassview init
```

Then launch with:

```
./keepassview
```

## Password storage

keepassview never stores your full master password. During `init` you choose how many leading characters you will type at each launch (the *prefix*). The remainder (the *suffix*) is saved in the OS credential store (Windows Credential Manager; not yet supported on Linux/macOS).

At launch you type the prefix from memory. The app fetches the suffix from the credential store, combines them in memory, opens the vault, then zeros both immediately. An attacker who extracts the credential store gets only the suffix — useless without the prefix you keep in your head.

If the OS credential store is unavailable, keepassview falls back to prompting for the full password each time.

## Idle timeout

The server shuts itself down after a configurable period of inactivity (default 300 s). You can also quit via the button in the UI or Ctrl-C in the terminal.
