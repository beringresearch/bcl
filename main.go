package main

import (
	"fmt"
	"github.com/beringresearch/bcl/parser"
)

// PackagesConf ...
type PackagesConf struct {
	Manager string
	System  []string
}

// RunConf ...
type RunConf struct {
	Command []string
}

// Environ ...
type Environ struct {
	Base     string
	Packages PackagesConf
	Run      RunConf
}

func main() {
	conf := parser.New("Bravefile.bcl")
	env := &Environ{}
	err := conf.Unmarshal(env)
	if err == nil {
		fmt.Println(env)
	} else {
		fmt.Println(err)
	}

	//fmt.Println(env.Run.Command[1])
}
