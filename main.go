package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var (
	flPackageName  = flag.String("package", "version", "name for the generated golang package")
	flVariableName = flag.String("variable", "VERSION", "variable name in the generated golang package")
	flOutputFile   = flag.String("output", "", "output filename (default stdout)")
)

func main() {
	flag.Parse()
	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Args()[0]
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	var output io.Writer
	if len(*flOutputFile) > 0 {
		fh, err := os.Create(*flOutputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer fh.Close()
		output = fh
	} else {
		output = os.Stdout
	}

	vers, err := GitDescribe(dir)
	if err != nil {
		log.Fatal(err)
	}
	vp := VersionPackage{
		Name:     *flPackageName,
		Path:     dir,
		Date:     time.Now(),
		Variable: *flVariableName,
		Version:  vers,
	}

	packageTemplate.Execute(output, vp)
}

// VersionPackage is the needed information to template a version package
type VersionPackage struct {
	Name     string
	Path     string
	Date     time.Time
	Variable string
	Version  string
}

var packageTemplate = template.Must(template.New("default").Parse(packageLayout))
var packageLayout = `package {{.Name}}
// AUTO-GENEREATED. DO NOT EDIT
// {{.Date}}

// {{.Variable}} is the generated version from {{.Path}}
var {{.Variable}} = "{{.Version}}"
 `

// GitDescribe calls `git describe` in the provided path
func GitDescribe(path string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// TODO check if this is a directory
	if err := os.Chdir(path); err != nil {
		return "", err
	}
	defer os.Chdir(cwd)

	buf, err := exec.Command("git", "describe").CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(buf)), nil
}
