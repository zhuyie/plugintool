package main

import (
	"debug/elf"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

func main() {
	var filename string
	flag.StringVar(&filename, "file", "plugin1.so", "ELF file name")
	flag.Parse()

	elf, err := elf.Open(filename)
	if err != nil {
		fmt.Printf("Open ELF file error: %v\n", err)
		return
	}

	var pluginPath string
	symbols, err := elf.Symbols()
	if err == nil {
		for _, sym := range symbols {
			if sym.Name == "go.link.thispluginpath" {
				path := make([]byte, int(sym.Size))
				_, err = readSection(elf, int(sym.Section), int64(sym.Value), path)
				if err == nil {
					pluginPath = string(path)
				}
			}
		}
	}
	if pluginPath != "" {
		fmt.Printf("PluginPath: %v\n\n", pluginPath)
	}

	symbols, err = elf.DynamicSymbols()
	if err != nil {
		fmt.Printf("Read DynamicSymbols error: %v\n", err)
		return
	}
	var i int
	for _, sym := range symbols {
		if strings.HasPrefix(sym.Name, "go.link.pkghashbytes.") {
			pkgName := sym.Name[len("go.link.pkghashbytes."):]

			hash := make([]byte, int(sym.Size))
			_, err = readSection(elf, int(sym.Section), int64(sym.Value), hash)
			if err == nil {
				fmt.Printf("%4d:  %v  %v\n", i, string(hash), pkgName)
			} else {
				fmt.Printf("%4d:  err=%v  %v\n", i, err, pkgName)
			}

			i++
		}
	}
}

func readSection(f *elf.File, sectionIndex int, offset int64, p []byte) (n int, err error) {
	if sectionIndex <= 0 || sectionIndex >= len(f.Sections) {
		return 0, errors.New("Invalid SectionIndex")
	}
	section := f.Sections[sectionIndex]
	startAddr := section.Addr
	if startAddr == 0 {
		startAddr = section.Offset
	}
	reader := section.Open()
	_, err = reader.Seek(offset-int64(startAddr), io.SeekStart)
	if err != nil {
		return 0, err
	}
	return reader.Read(p)
}
