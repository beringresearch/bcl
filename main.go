package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/beringresearch/bcl/parser"
	"gopkg.in/yaml.v2"
)

func main() {

	example := `//BCL File example
base {
	image: 		""
	location: 	""
}

system {
    apt: 		["bash", "python3"]
}

copy {
	FileName {
	source:		"/file/or/directory"
	target: 	"/file/or/directory"
	action: 	 "chmod 0700 /file/or/directory"
	}
}

run {
	echo: 		"Hello World"
}

service {
	image: ""
	docker: "no"	
	name:		""
	version:	"1.0"
	ip: 		""
	ports:		""
	postdeploy {
		copy {
			FileName {
			source:		"/file/or/directory"
			target: 	"/file/or/directory"
			action: 	 "chmod 0700 /file/or/directory"
			}
		}
		run {
			echo: 		"Hello World"
		}	
	}
	resources {
		ram: 	"4GB"
		cpu: 	2
		gpu:	"no"
	}
}`
	if len(os.Args) < 2 {
		fmt.Println(example)
		return
	}

	path := os.Args[1]
	absPath, _ := filepath.Abs(path)
	file, err := ioutil.ReadFile(absPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	grammar, err := parser.Scan(string(file))
	if err != nil {
		log.Fatal(err.Error())
	}

	bravefile, err := parser.Parse(grammar)
	if err != nil {
		log.Fatal(err.Error())
	}

	output, err := yaml.Marshal(bravefile)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = ioutil.WriteFile("Bravefile", output, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
}
