package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type ByteWriter interface {
	Write(b []byte) (n int, err error)
}

type TreeNode struct {
	File   os.DirEntry
	Childs []TreeNode
}

func (node *TreeNode) Empty() bool {
	return node.File == nil
}

func (node *TreeNode) FormattedName() string {
	if node.Empty() {
		return ""
	}

	if node.File.IsDir() {
		return node.File.Name()
	} else {
		return fmt.Sprintf("%s (%s)", node.File.Name(), node.Size())
	}
}

func (node *TreeNode) Size() string {
	if node.Empty() {
		return ""
	}

	fileInfo, _ := node.File.Info()
	if fileInfo.Size() > 0 {
		return fmt.Sprintf("%db", fileInfo.Size())
	} else {
		return "empty"
	}
}

func getDirPanics(path string) []os.DirEntry {
	files, err := os.ReadDir(path)
	if err != nil {
		panic(err.Error())
	}

	return files
}

func makeDirTree(path string, isPrintFiles bool) []TreeNode {
	files := getDirPanics(path)
	var tree []TreeNode
	for _, file := range files {
		if !file.IsDir() && !isPrintFiles {
			continue
		}

		node := TreeNode{file, nil}
		if file.IsDir() {
			node.Childs = makeDirTree(filepath.Join(path, file.Name()), isPrintFiles)
		}
		tree = append(tree, node)
	}

	return tree
}

func printDirTree(out ByteWriter, tree []TreeNode, pathPrefix string) {
	if tree == nil {
		return
	}
	const (
		levelSymbol        = "│\t"
		dirOfOneElemSymbol = "\t"
		leafSymbol         = "├───"
		lastLeafSymbol     = "└───"
	)

	for i := 0; i < len(tree); i++ {
		printLeafSymbol := leafSymbol
		lvlSymbol := levelSymbol
		if i == len(tree)-1 {
			printLeafSymbol = lastLeafSymbol
			lvlSymbol = dirOfOneElemSymbol
		}

		fmt.Fprint(out, pathPrefix, printLeafSymbol, tree[i].FormattedName(), "\n")
		//out.Write([]byte(pathPrefix + printLeafSymbol + tree[i].FormattedName() + "\n"))
		printDirTree(out, tree[i].Childs, pathPrefix+lvlSymbol)
	}
}

func dirTree(treeResults ByteWriter, path string, isPrintFiles bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprint(treeResults, "Error occured: ", err)
		}
	}()
	tree := makeDirTree(path, isPrintFiles)
	printDirTree(treeResults, tree, "")
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	dirTree(out, path, printFiles)
	fmt.Printf("%v", out)
}
