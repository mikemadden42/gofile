// Package main implements gofile, a file(1)-like utility that reports the
// format and CPU architecture of Mach-O, ELF, and PE binaries.
package main

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Based on the example program found here:
// https://www.jvt.me/posts/2023/05/15/go-parse-binary-architecture/
// Docs for debug package can be found here:
// https://pkg.go.dev/debug

type parser struct {
	format string
	parse  func(name string) (string, error)
}

var parsers = []parser{
	{"Mach-O", parseMac},
	{"Mach-O universal", parseMacUniversalBinary},
	{"ELF", parseElf},
	{"PE", parsePE},
}

var errUnrecognized = errors.New("unrecognized format")

func main() {
	verbose := flag.Bool("v", false, "print why each format did not match")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-v] file...\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	status := 0
	for _, name := range flag.Args() {
		if err := identify(name, *verbose); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			status = 1
		}
	}
	os.Exit(status)
}

func identify(name string, verbose bool) error {
	if _, err := os.Stat(name); err != nil {
		return err
	}

	for _, p := range parsers {
		desc, err := p.parse(name)
		if err == nil {
			fmt.Printf("%s is %s\n", name, desc)
			return nil
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "%s: not %s: %v\n", name, p.format, err)
		}
	}

	return errUnrecognized
}

func parseMac(name string) (string, error) {
	f, err := macho.Open(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("a Mach-O binary with CPU architecture %v", f.Cpu), f.Close()
}

func parseMacUniversalBinary(name string) (string, error) {
	f, err := macho.OpenFat(name)
	if err != nil {
		return "", err
	}
	arches := make([]string, 0, len(f.Arches))
	for _, fa := range f.Arches {
		arches = append(arches, fa.Cpu.String())
	}

	return "a Mach-O universal binary with architectures: " + strings.Join(arches, " "), f.Close()
}

func parseElf(name string) (string, error) {
	f, err := elf.Open(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("an Executable and Linked Format (ELF) binary with CPU architecture %v", f.Machine), f.Close()
}

func parsePE(name string) (string, error) {
	f, err := pe.Open(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("a PE binary with CPU architecture 0x%x", f.Machine), f.Close()
}
