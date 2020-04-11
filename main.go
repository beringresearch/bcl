package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/beringresearch/bcl/parser"
	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) < 2 {
		panic("no valid file name or path provided for file!")
	}

	path := os.Args[1]
	absPath, _ := filepath.Abs(path)
	file, err := ioutil.ReadFile(absPath)
	if err != nil {
		panic(err.Error())
	}

	grammar, err := parser.Scan(string(file))
	if err != nil {
		panic(err.Error())
	}

	bravefile, err := parser.Parse(grammar)
	output, err := yaml.Marshal(bravefile)
	if err != nil {
		panic(err.Error())
	}

	err = ioutil.WriteFile("Bravefile", output, 0644)
	if err != nil {
		panic(err.Error())
	}
}
