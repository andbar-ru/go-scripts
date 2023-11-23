package main

import (
	"bufio"
	"cmp"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var stat = make(map[string]int)

func parseLine(line string, grouped bool) error {
	fields := strings.Fields(line)
	var pkg string

	switch len(fields) {
	case 1:
		if !grouped {
			return fmt.Errorf("incorrect import string %q", line)
		}
		pkg = fields[0]
	case 2:
		pkg = fields[1]
	case 3:
		if grouped {
			return fmt.Errorf("incorrect import string %q", line)
		}
		pkg = fields[2]
	default:
		return fmt.Errorf("incorrect import string %q", line)
	}

	pkg = strings.Trim(pkg, "\"")
	stat[pkg]++
	return nil
}

func parseGoFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var insideImports, grouped bool
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !insideImports {
			if strings.HasPrefix(line, "import") {
				insideImports = true
				grouped = strings.HasPrefix(line, "import (")
				if !grouped {
					err := parseLine(line, grouped)
					if err != nil {
						return fmt.Errorf("file %q: %s", path, err.Error())
					}
				}
			}
		} else {
			if grouped {
				if line == ")" {
					break
				}
				err := parseLine(line, grouped)
				if err != nil {
					return fmt.Errorf("file %q: %s", path, err.Error())
				}
			} else {
				if !strings.HasPrefix(line, "import") {
					break
				}
				err := parseLine(line, grouped)
				if err != nil {
					return fmt.Errorf("file %q: %s", path, err.Error())
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "You must specify directory as the first argument!")
		os.Exit(1)
	}
	dir := args[0]
	info, err := os.Stat(dir)
	if err != nil {
		log.Fatal(err)
	}
	if !info.IsDir() {
		log.Fatalf("%q is not a directory\n", dir)
	}

	err = filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" {
			err = parseGoFile(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("error walking the dir %q: %v\n", dir, err)
	}

	keys := make([]string, 0, len(stat))
	for key := range stat {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b string) int {
		if n := cmp.Compare(stat[b], stat[a]); n != 0 { // reversed
			return n
		}
		return cmp.Compare(a, b)
	})
	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, stat[key])
	}
}
