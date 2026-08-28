package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
)

var (
	xbindkeysrcPath = path.Join(os.Getenv("HOME"), ".xbindkeysrc")
	menuPath        = path.Join(os.Getenv("HOME"), ".fluxbox/mod3-menu")
)

func main() {
	xbindkeysrcFile, err := os.Open(xbindkeysrcPath)
	if err != nil {
		panic(err)
	}
	defer xbindkeysrcFile.Close()

	var menu strings.Builder

	menu.WriteString("[submenu] (Mod3)")

	scanner := bufio.NewScanner(xbindkeysrcFile)
	var command string
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) > 0 && text[0] == '"' {
			command = strings.Split(text, "\"")[1]
			continue
		} else if command != "" && slices.Contains(strings.Fields(text), "Mod3") {
			menu.WriteString("\t[exec] (" + text + ") {" + command + "}\n")
		}

		command = ""
	}

	if err = scanner.Err(); err != nil {
		panic(err)
	}

	menu.WriteString("[end]")

	menuFile, err := os.Create(menuPath)
	if err != nil {
		panic(err)
	}
	defer menuFile.Close()

	_, err = menuFile.WriteString(menu.String())
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s written\n", menuPath)
}
