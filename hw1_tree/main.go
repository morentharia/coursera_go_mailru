package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

func readDirNames(path string, printFiles bool) ([]os.FileInfo, error) {
	dir, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return []os.FileInfo{}, err
	}

	infos, err := dir.Readdir(0)

	var res []os.FileInfo
	for idx := range infos {
		if infos[idx].IsDir() || printFiles {
			res = append(res, infos[idx])
		}
	}
	return res, nil
}

func createLine(isLast bool, prefix string, info os.FileInfo) (lines string) {
	line := prefix
	if isLast {
		line += "└───"
	} else {
		line += "├───"
	}
	line += info.Name()

	if !info.IsDir() {
		var size string
		if info.Size() > 0 {
			size = strconv.FormatInt(info.Size(), 10) + "b"
		} else {
			size = "empty"
		}
		line += " (" + size + ")"
	}
	return line
}

func dirTreeLevel(output io.Writer, path string, printFiles bool, prefix string) error {
	infos, err := readDirNames(path, printFiles)
	if err != nil {
		return err
	}

	sort.Slice(infos, func(a, b int) bool {
		return infos[a].Name() < infos[b].Name()
	})

	for idx, info := range infos {
		isLast := idx == len(infos)-1

		fmt.Fprintln(output, createLine(
			isLast,
			prefix,
			info,
		))

		if info.IsDir() {
			var subPrefix string
			if isLast {
				subPrefix = prefix + "\t"
			} else {
				subPrefix = prefix + "│\t"
			}
			err := dirTreeLevel(
				output,
				path+string(os.PathSeparator)+info.Name(),
				printFiles,
				subPrefix,
			)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}

	return nil
}

func dirTree(output io.Writer, path string, printFiles bool) error {
	return dirTreeLevel(output, path, printFiles, "")
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
