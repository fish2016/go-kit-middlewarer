package main

import (
	"fmt"
	"go/ast"
	"path"
	// "path/filepath"
	"strings"
)

type Import struct {
	name      string
	path      string
	last      string
	isEmbeded bool
	isParam   bool
}

func createImportWithPath(p string) *Import {
	fmt.Println("createImportWithPath", p)
	last := path.Base(p)
	name := last
	if strings.Contains(last, "-") {
		lastPieces := strings.Split(last, "-")
		name = lastPieces[len(lastPieces)-1]
	}
	fmt.Println("createImportWithPath", name, p, last)
	return &Import{
		name: name,
		path: p,
		last: last,
	}
}

func createImport(imp *ast.ImportSpec) *Import {
	var name string
	pth := strings.TrimPrefix(strings.TrimSuffix(imp.Path.Value, "\""), "\"")
	last := path.Base(pth)
	if n := imp.Name; n == nil {
		name = last
	} else {
		name = n.String()
	}

	if strings.Contains(name, "-") {
		namePieces := strings.Split(name, "-")
		name = namePieces[len(namePieces)-1]
	}

	return &Import{
		name: name,
		path: pth,
		last: last,
	}
}

func (i Import) ImportSpec() string {
	if i.name == i.last {
		return fmt.Sprintf("\"%s\"", i.path)
	}

	return fmt.Sprintf("%s \"%s\"", i.name, i.path)
}
