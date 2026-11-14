package main

import (
	"debug/macho"
	"debug/pe"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	desc, err := parse(exe)
	if err != nil {
		t.Fatalf("test binary on %s not parsed as native format: %v", runtime.GOOS, err)
	}

	for _, want := range []string{" executable ", " " + runtime.GOARCH} {
		if !strings.Contains(desc, want) {
			t.Errorf("description %q does not contain %q", desc, want)
		}
	}
}

func TestArchNames(t *testing.T) {
	tests := []struct {
		got, want string
	}{
		{machoArch(macho.CpuAmd64, 3), "amd64"},
		{machoArch(macho.CpuArm64, 0), "arm64"},
		{machoArch(macho.CpuArm64, 0x80000002), "arm64e"},
		{machoArch(macho.Cpu(0x1234), 0), "cpu 0x1234"},
		{peArch(pe.IMAGE_FILE_MACHINE_AMD64), "amd64"},
		{peArch(pe.IMAGE_FILE_MACHINE_I386), "386"},
		{peArch(0x1234), "machine 0x1234"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("got %q, want %q", tt.got, tt.want)
		}
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
