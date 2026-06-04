# keepassview

A read-only KeePass (.kdbx) database viewer that runs a local web server and opens the vault in your browser.

## Build

Requires Go 1.25+.

```
git clone https://github.com/yetanotherchris/keepassview
cd keepassview
go build -o keepassview .
```

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
