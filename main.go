package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"

	"keepassview/internal/config"
	"keepassview/internal/creds"
	"keepassview/internal/server"
	"keepassview/internal/vault"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "init: %v\n", err)
			os.Exit(1)
		}
		return
	}

	var port int
	flag.IntVar(&port, "port", 0, "port to listen on (0 = ephemeral)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	pwBytes, err := resolvePassword(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "password: %v\n", err)
		os.Exit(1)
	}

	v, vaultErr := vault.Open(cfg.DBPath, string(pwBytes))
	zeroBytes(pwBytes)
	if vaultErr != nil {
		fmt.Fprintln(os.Stderr, vaultErr)
		os.Exit(1)
	}

	if err := server.Run(v, cfg, port, assets); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}

// resolvePassword obtains the full master password.
// In Mode A (credential store) the user types the N-char prefix; the suffix is
// fetched from the store and appended. In Mode B the user types the full password.
func resolvePassword(cfg *config.Config) ([]byte, error) {
	store := creds.New()

	if cfg.UseCredentialStore && store.Available() {
		key := credentialKey(cfg.DBPath)
		suffix, err := store.Get(key)
		if err != nil {
			return nil, fmt.Errorf("no stored credential for this database; run keepassview init (%w)", err)
		}

		fmt.Printf("Enter first %d character(s) of master password: ", cfg.PromptCharCount)
		prefixBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return nil, fmt.Errorf("read password: %w", err)
		}

		full := append(prefixBytes, []byte(suffix)...)
		zeroBytes(prefixBytes)
		return full, nil
	}

	// Mode B — full password.
	fmt.Print("Master password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("read password: %w", err)
	}
	return pw, nil
}

// runInit interactively configures the tool and (optionally) enrolls the
// credential store. Re-running it replaces any existing configuration.
func runInit() error {
	store := creds.New()
	cfg := config.Default()

	reader := bufio.NewReader(os.Stdin)

	// ── DB path ──────────────────────────────────────────────────────────────
	fmt.Print("Path to .kdbx database file: ")
	dbPath, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	dbPath = strings.TrimSpace(dbPath)
	if dbPath == "" {
		return fmt.Errorf("database path is required")
	}
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("cannot access %q: %w", dbPath, err)
	}
	cfg.DBPath = dbPath

	// ── Credential store ─────────────────────────────────────────────────────
	if store.Available() {
		fmt.Print("Store password suffix in OS credential store? [Y/n]: ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		cfg.UseCredentialStore = answer == "" || answer == "y" || answer == "yes"
	} else {
		fmt.Printf("Note: OS credential store not available on %s; using full-password mode.\n", runtime.GOOS)
		cfg.UseCredentialStore = false
	}

	// ── Prompt char count (only relevant for Mode A) ─────────────────────────
	if cfg.UseCredentialStore {
		fmt.Printf("How many leading characters to type at launch? [%d]: ", cfg.PromptCharCount)
		nStr, _ := reader.ReadString('\n')
		nStr = strings.TrimSpace(nStr)
		if nStr != "" {
			n := 0
			if _, err := fmt.Sscan(nStr, &n); err != nil || n < 3 {
				fmt.Println("Defaulting to 3.")
				n = 3
			}
			cfg.PromptCharCount = n
		}
	}

	// ── Idle timeout ─────────────────────────────────────────────────────────
	fmt.Printf("Idle timeout in seconds? [%d]: ", cfg.IdleTimeoutSeconds)
	tStr, _ := reader.ReadString('\n')
	tStr = strings.TrimSpace(tStr)
	if tStr != "" {
		t := 0
		if _, err := fmt.Sscan(tStr, &t); err != nil || t < 30 {
			fmt.Println("Minimum 30 s; using 30.")
			t = 30
		}
		cfg.IdleTimeoutSeconds = t
	}

	// ── Master password ───────────────────────────────────────────────────────
	fmt.Print("Master password (not echoed): ")
	pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read password: %w", err)
	}

	// Verify the password actually decrypts the database.
	fmt.Println("Verifying password…")
	_, verifyErr := vault.Open(cfg.DBPath, string(pwBytes))
	if verifyErr != nil {
		zeroBytes(pwBytes)
		return fmt.Errorf("password verification failed: %w", verifyErr)
	}
	fmt.Println("Password verified.")

	// ── Enroll credential store (Mode A) ─────────────────────────────────────
	if cfg.UseCredentialStore {
		runes := []rune(string(pwBytes))
		n := cfg.PromptCharCount
		if len(runes) <= n {
			fmt.Printf("Password has only %d rune(s), which is not longer than N=%d. Falling back to full-password mode.\n", len(runes), n)
			cfg.UseCredentialStore = false
		} else {
			suffix := string(runes[n:])
			key := credentialKey(cfg.DBPath)
			if err := store.Set(key, suffix); err != nil {
				zeroBytes(pwBytes)
				return fmt.Errorf("store credential: %w", err)
			}
			fmt.Println("Password suffix stored in credential store.")
		}
	}

	zeroBytes(pwBytes)

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Println("Configuration saved. Run keepassview to launch the viewer.")
	return nil
}

// credentialKey returns the credential store key for the given database path.
func credentialKey(dbPath string) string {
	h := sha256.Sum256([]byte(dbPath))
	return fmt.Sprintf("keepassview:%x", h[:])
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
