package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIdentifySelf(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	if err := identify(exe, false); err != nil {
		t.Fatalf("identify(%q) = %v, want nil", exe, err)
	}
}

func TestHostFormat(t *testing.T) {
	parse := map[string]func(string) (string, error){
		"darwin":  parseMac,
		"linux":   parseElf,
		"windows": parsePE,
	}[runtime.GOOS]
	if parse == nil {
		t.Skipf("no expected format for %s", runtime.GOOS)
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := parse(exe); err != nil {
		t.Fatalf("test binary on %s not parsed as native format: %v", runtime.GOOS, err)
	}
}

func TestIdentifyUnrecognized(t *testing.T) {
	name := filepath.Join(t.TempDir(), "plain.txt")
	if err := os.WriteFile(name, []byte("hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := identify(name, false); !errors.Is(err, errUnrecognized) {
		t.Fatalf("identify(%q) = %v, want %v", name, err, errUnrecognized)
	}
}

func TestIdentifyMissing(t *testing.T) {
	name := filepath.Join(t.TempDir(), "missing")

	if err := identify(name, false); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("identify(%q) = %v, want fs.ErrNotExist", name, err)
	}
}
