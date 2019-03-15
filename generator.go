package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Generator holds the state of the analysis.  Primarily used to buffer the
// output for format.Source.
type Generator struct {
	buf bytes.Buffer // Accumulated ouptu.
	pkg *Package
}

// Printf writes the given output to the internalized buffer.
func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) parsePackageDir(directory string) {
	pkg, err := build.Default.ImportDir(directory, 0)
	if err != nil {
		log.Fatalf("cannot process directory %s: %s", directory, err)
	}

	d, e := os.Getwd()
	gopath := os.Getenv("GOPATH")

	if e != nil {
		log.Fatalf("Error Grabbing WD: %s\n", e)
	}

	prefix := filepath.Join(gopath, "src") + string([]rune{filepath.Separator})

	fmt.Println("parsePackageDir prefix:", prefix, d)
	d, err = filepath.Rel(prefix, d)
	if err != nil {
		log.Fatalf("Unable to get a relative path: %s\n", err)
	}
	fmt.Println("parsePackageDir d:", d)
	var names []string
	names = append(names, pkg.GoFiles...) //指定目录下的go文件
	fmt.Println("parsePackageDir::names:", names)

	names = prefixDirectory(directory, names)
	g.parsePackage(d, names, nil)
}

func (g *Generator) parsePackageFiles(names []string) {
	g.parsePackage(".", names, nil)
}

// parsePackage analyzes the signle package constructed from the named files.
// If text is non-nil, it is a string to be used instead of the content of the file,
// to be used for testing.  parsePackage exists if there is an error.
func (g *Generator) parsePackage(directory string, names []string, text interface{}) {
	var files []*File
	var astFiles []*ast.File
	g.pkg = new(Package)
	fs := token.NewFileSet()

	fmt.Println("parsePackage:", directory, names)
	fmt.Println("parsePackage:log.Prefix()", log.Prefix())
	for _, name := range names {
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		parsedFile, err := parser.ParseFile(fs, name, text, parser.ParseComments)

		for _, v := range parsedFile.Comments {
			str := v.Text()
			fmt.Println("parsePackage.parsedFile.Comments str:", str)
			if strings.HasPrefix(str, SubCmdTempl) {
				lines := strings.Split(str, "\n")
				if len(lines) <= 0 {
					continue
				}
				var firstLine = lines[0]

				//typ := strings.TrimPrefix(firstLine, SubCmdExtra)
				lineTmp := strings.TrimPrefix(firstLine, SubCmdTempl)
				arrString := strings.Split(lineTmp, " ")
				typ := arrString[0]

				//check the validity of subcmd flag
				if (typ != SubCmdFlagLogging) && (typ != SubCmdFlagUrl) {
					log.Fatalf("%s: invalid subcmd flag: <%s>", directory, typ)
				}

				lineTmp = strings.TrimPrefix(lineTmp, typ)
				lineTmp = strings.TrimSpace(lineTmp)

				if len(lineTmp) > 0 {
					extras[typ] += (lineTmp + "\n")
				}

				if len(lines) > 1 {
					//extras[typ] = strings.Join(lines[1:], "\n")
					extras[typ] += strings.Join(lines[1:], "\n")
				}
			}
		}

		if err != nil {
			log.Fatalf("parsing package: %s: %s", name, err)
		}
		astFiles = append(astFiles, parsedFile)
		files = append(files, &File{
			file:     parsedFile,
			pkg:      g.pkg,
			path:     directory,
			fileName: name,
		})
	}

	if len(astFiles) == 0 {
		log.Fatalf("%s: no buildable Go files", directory)
	}
	g.pkg.name = astFiles[0].Name.Name
	g.pkg.files = files
	g.pkg.dir = directory
	g.pkg.check(fs, astFiles)
	fmt.Println("parsePackage g.pkg.name:", g.pkg.name)
}

// generate does 'things'
func (g *Generator) generate(typeName string) {
	// pre-process
	for _, file := range g.pkg.files {
		// Set the state for this run of the walker.
		fmt.Println("generate140:", file.fileName)
		if file.file != nil {
			ast.Inspect(file.file, file.genImportsAndTypes)
		}
	}

	if *summarize != "" {
		g.pkg.Summarize()
		return
	}

	var targetFile *File

	for _, file := range g.pkg.files {
		for _, i := range file.interfaces {
			fmt.Println("Generator range:", i.name)
			if i.name == typeName {
				targetFile = file
				break
			}
		}
	}

	if targetFile == nil {
		log.Fatalf("Unable to fine the type specified: %s\n", typeName)
	}

	// begin generation
	list := strings.Split(*middlewaresToGenerate, ",")
	//list = append(list, "endpoint")
	//list = append(list, "default")
	for _, l := range list {
		if bindings[l] != nil {
			bindings[l](g, targetFile)
		}
	}
}

// format returns gofmt-ed contents of the Generator's buffer.
func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the pacakge to analyze the error")
		return g.buf.Bytes()
	}
	return src
}
