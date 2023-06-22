package main

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"log"
	"os"
)

// Based on the example program found here:
// https://www.jvt.me/posts/2023/05/15/go-parse-binary-architecture/
func main() {
	file, err := os.OpenFile("gofile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	logger := log.New(file, "", log.LstdFlags)

	if len(os.Args) != 2 {
		log.Fatal("Need more args")
	}
	command := os.Args[1]

	err = parseMac(command)
	if err != nil {
		logger.Println("Doesn't look like a Mach-O file:", err, command)
	}

	err = parseMacUniversalBinary(command)
	if err != nil {
		logger.Println("Doesn't look like a Mach-O Universal Binary:", err, command)
	}

	err = parseElf(command)
	if err != nil {
		logger.Println("Doesn't look like an ELF file:", err, command)
	}

	err = parsePE(command)
	if err != nil {
		logger.Println("Doesn't look like a PE file:", err, command)
	}
}

func parseMac(command string) error {
	f, err := macho.Open(command)
	if err != nil {
		return err
	}

	fmt.Printf("%s is a Mach-O binary with CPU architecture %v\n", command, f.Cpu.String())

	return nil
}

func parseMacUniversalBinary(command string) error {
	f, err := macho.OpenFat(command)
	if err != nil {
		return err
	}
	fmt.Printf("%s is a Mach-O universal binary with architectures:", command)
	for _, fa := range f.Arches {
		fmt.Printf(" %v", fa.Cpu.String())
	}
	fmt.Println()

	return nil
}

func parseElf(command string) error {
	f, err := elf.Open(command)
	if err != nil {
		return err
	}

	fmt.Printf("%s is an Executable and Linked Format (ELF) binary with CPU architecture %v\n", command, f.Machine.String())

	return nil
}

func parsePE(command string) error {
	f, err := pe.Open(command)
	if err != nil {
		return err
	}

	fmt.Printf("%s is a PE binary with CPU architecture 0x%x\n", command, f.Machine)

	return nil
}
