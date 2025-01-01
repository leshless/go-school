package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {	
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	
	out := os.Stdout

	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err)
	}
}


func dirTree(out io.Writer, path string, printFiles bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("attemted to get working directory path but got error: %w", err)
	}

	path = filepath.Join(wd, path)

	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("attemted to retrieve directory stat but got error: %w", err)
	}

	if !stat.IsDir(){
		return fmt.Errorf("second argument is not a directory")
	}

	return printContents(out, path, "", printFiles)
}

// HELPERS

func printContents(out io.Writer, path string, prefix string, printFiles bool) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("attempted to read directory contents but got error: %w", err)
	}

	if !printFiles {
		filtered := make([]os.DirEntry, 0)
		for _, file := range files {
			if file.IsDir(){
				filtered = append(filtered, file)
			}
		}	

		files = filtered
	}

	for i, file := range files {
		isLast := (i == len(files) - 1)

		if !file.IsDir() && !printFiles {
			continue
		}

		err := writeString(out, prefix, file, isLast)
		if err != nil {
			return fmt.Errorf("attempted to write data to output but got error: %w", err)
		}

		if !file.IsDir(){
			continue
		}

		name := file.Name()
		
		dirPath := filepath.Join(path, name)
		dirPrefix := incrementPrefix(prefix, isLast)

		err = printContents(out, dirPath, dirPrefix, printFiles)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeString(out io.Writer, prefix string, file fs.DirEntry, isLast bool) error {
	err := writePrefixString(out, prefix, isLast)
	if err != nil{
		return err
	}

	if file.IsDir() {
		name := file.Name()

		err := writeDirString(out, name)
		if err != nil{
			return err
		}
	}else {
		name := file.Name()

		info, err := file.Info()
		if err != nil {
			return err
		}

		size := info.Size()

		err = writeFileString(out, name, size)
		if err != nil{
			return err
		}
	}

	return nil
}

func writePrefixString(out io.Writer, prefix string, isLast bool) error {
	var err error	
	if isLast {
		_, err = out.Write([]byte(fmt.Sprintf("%s└───", prefix)))
	}else {
		_, err = out.Write([]byte(fmt.Sprintf("%s├───", prefix)))
	}
	
	return err
}

func writeDirString(out io.Writer, name string) error {
	_, err := out.Write([]byte(fmt.Sprintf("%s\n", name)))
	
	return err
}

func writeFileString(out io.Writer, name string, size int64) error {
	var err error
	if size == 0{
		_, err = out.Write([]byte(fmt.Sprintf("%s (empty)\n", name)))
	}else{
		_, err = out.Write([]byte(fmt.Sprintf("%s (%db)\n", name, size)))
	}

	return err
}

func incrementPrefix(prefix string, isLast bool) string {
	if !isLast {
		prefix += "│"
	}
	prefix += "\t"

	return prefix
}