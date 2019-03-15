package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

func processEndpoint(g *Generator, f *File) {
	gopath := os.Getenv("GOPATH")
	var buf bytes.Buffer

	tmpl, err := template.ParseFiles(filepath.Join(gopath, "src", "github.com", "fish2016", "go-kit-middlewarer", "tmpl", "endpoint.tmpl"))
	if err != nil {
		log.Fatalf("Template Parse Error: %s", err)
	}

	convertedPath := filepath.ToSlash(f.pkg.dir)

	endpointPackage := createImportWithPath(path.Join(convertedPath, "endpoint"))
	basePackage := createImportWithPath(convertedPath)

	for _, interf := range f.interfaces {
		err := tmpl.Execute(&buf, createTemplateBase(basePackage, endpointPackage, interf, f.imports))
		if err != nil {
			log.Fatalf("Template execution failed: %s\n", err)
		}
	}

	filename := "defs_gen.go"

	file := openFile(filepath.Join(".", "endpoint"), filename)
	defer file.Close()

	fmt.Fprint(file, string(formatBuffer(buf, filename)))
}

func init() {
	registerProcess("endpoint", processEndpoint)
}
