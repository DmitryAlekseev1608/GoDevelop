package main

import (
	"fmt"
	"io"
	"os"
	"log"
)

func getTreeDirFull(pre_text string, path string) (result string) {
	var pre_text_cur string
	result = ""
	tree, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for i, file := range tree {
		path_child := fmt.Sprintf("%s/%s", path, file.Name())
		info, _ := os.Stat(path_child)
		if i != len(tree)-1 {
			result += fmt.Sprintf("%s├───%s\n", pre_text, file.Name())
		} else {
			result += fmt.Sprintf("%s└───%s\n", pre_text, file.Name())
		}
		if 	info.IsDir() {
			if i != len(tree)-1 {
				pre_text_cur = pre_text + "│	"
			} else {
				pre_text_cur = pre_text + " 	"
			}
			result += getTreeDirFull(pre_text_cur, path_child)
		}
	}
	return
}

func dirTree(out io.Writer, path string, printFiles bool) (err error) {
	if printFiles {
		fmt.Println(getTreeDirFull("", path))
	}
	return
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
