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
			fmt.Printf("%s: %s\n", name, desc)
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

	class := "32-bit"
	if f.Magic == macho.Magic64 {
		class = "64-bit"
	}

	desc := fmt.Sprintf("Mach-O %s %s %s", class, machoType(f.Type), machoArch(f.Cpu, f.SubCpu))
	return desc, f.Close()
}

func parseMacUniversalBinary(name string) (string, error) {
	f, err := macho.OpenFat(name)
	if err != nil {
		return "", err
	}

	arches := make([]string, 0, len(f.Arches))
	for _, fa := range f.Arches {
		arches = append(arches, machoArch(fa.Cpu, fa.SubCpu))
	}

	return "Mach-O universal binary [" + strings.Join(arches, " ") + "]", f.Close()
}

func machoType(t macho.Type) string {
	switch t {
	case macho.TypeObj:
		return "object"
	case macho.TypeExec:
		return "executable"
	case macho.TypeDylib:
		return "dynamically linked shared library"
	case macho.TypeBundle:
		return "bundle"
	}
	return fmt.Sprintf("type 0x%x", uint32(t))
}

// machoSubCPUArm64e is CPU_SUBTYPE_ARM64E; the high byte of a subtype holds
// capability flags, so it is masked off before comparing.
const machoSubCPUArm64e = 2

func machoArch(cpu macho.Cpu, subCPU uint32) string {
	switch cpu {
	case macho.Cpu386:
		return "386"
	case macho.CpuAmd64:
		return "amd64"
	case macho.CpuArm:
		return "arm"
	case macho.CpuArm64:
		if subCPU&0x00ffffff == machoSubCPUArm64e {
			return "arm64e"
		}
		return "arm64"
	case macho.CpuPpc:
		return "ppc"
	case macho.CpuPpc64:
		return "ppc64"
	}
	return fmt.Sprintf("cpu 0x%x", uint32(cpu))
}

func parseElf(name string) (string, error) {
	f, err := elf.Open(name)
	if err != nil {
		return "", err
	}

	class := "32-bit"
	if f.Class == elf.ELFCLASS64 {
		class = "64-bit"
	}

	desc := fmt.Sprintf("ELF %s %s %s", class, elfType(f), elfArch(f))
	return desc, f.Close()
}

func elfType(f *elf.File) string {
	switch f.Type {
	case elf.ET_REL:
		return "relocatable"
	case elf.ET_EXEC:
		return "executable"
	case elf.ET_DYN:
		// PIE executables are ET_DYN too; DF_1_PIE tells them apart from
		// shared libraries.
		if flags, err := f.DynValue(elf.DT_FLAGS_1); err == nil && len(flags) > 0 && flags[0]&uint64(elf.DF_1_PIE) != 0 {
			return "pie executable"
		}
		return "shared object"
	case elf.ET_CORE:
		return "core file"
	}
	return fmt.Sprintf("type 0x%x", uint16(f.Type))
}

func elfArch(f *elf.File) string {
	le := f.Data == elf.ELFDATA2LSB
	switch f.Machine {
	case elf.EM_386:
		return "386"
	case elf.EM_X86_64:
		return "amd64"
	case elf.EM_ARM:
		return "arm"
	case elf.EM_AARCH64:
		return "arm64"
	case elf.EM_RISCV:
		if f.Class == elf.ELFCLASS64 {
			return "riscv64"
		}
		return "riscv"
	case elf.EM_PPC64:
		if le {
			return "ppc64le"
		}
		return "ppc64"
	case elf.EM_MIPS:
		if le {
			return "mipsle"
		}
		return "mips"
	case elf.EM_S390:
		return "s390x"
	case elf.EM_LOONGARCH:
		return "loong64"
	}
	return strings.ToLower(strings.TrimPrefix(f.Machine.String(), "EM_"))
}

func parsePE(name string) (string, error) {
	f, err := pe.Open(name)
	if err != nil {
		return "", err
	}

	class := "32-bit"
	if _, ok := f.OptionalHeader.(*pe.OptionalHeader64); ok {
		class = "64-bit"
	}

	kind := "object"
	switch {
	case f.Characteristics&pe.IMAGE_FILE_DLL != 0:
		kind = "DLL"
	case f.Characteristics&pe.IMAGE_FILE_EXECUTABLE_IMAGE != 0:
		kind = "executable"
	}

	desc := fmt.Sprintf("PE %s %s %s", class, kind, peArch(f.Machine))
	return desc, f.Close()
}

func peArch(machine uint16) string {
	switch machine {
	case pe.IMAGE_FILE_MACHINE_I386:
		return "386"
	case pe.IMAGE_FILE_MACHINE_AMD64:
		return "amd64"
	case pe.IMAGE_FILE_MACHINE_ARMNT:
		return "arm"
	case pe.IMAGE_FILE_MACHINE_ARM64:
		return "arm64"
	case pe.IMAGE_FILE_MACHINE_RISCV64:
		return "riscv64"
	case pe.IMAGE_FILE_MACHINE_LOONGARCH64:
		return "loong64"
	}
	return fmt.Sprintf("machine 0x%x", machine)
}
