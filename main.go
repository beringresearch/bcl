package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/repr"
)

type Bool bool

func (b *Bool) Capture(v []string) error { *b = v[0] == "true"; return nil }

type Value struct {
	Boolean    *Bool    `  @("true"|"false")`
	Identifier *string  `| @Ident { @"." @Ident }`
	String     *string  `| @(String|Char|RawString)`
	Number     *float64 `| @(Float|Int)`
	Array      []*Value `| "[" { @@ [ "," ] } "]"`
}

func (l *Value) GoString() string {
	switch {
	case l.Boolean != nil:
		return fmt.Sprintf("%v", *l.Boolean)
	case l.Identifier != nil:
		return fmt.Sprintf("`%s`", *l.Identifier)
	case l.String != nil:
		return fmt.Sprintf("%q", *l.String)
	case l.Number != nil:
		return fmt.Sprintf("%v", *l.Number)
	case l.Array != nil:
		out := []string{}
		for _, v := range l.Array {
			out = append(out, v.GoString())
		}
		return fmt.Sprintf("[]*Value{ %s }", strings.Join(out, ", "))
	}
	panic("??")
}

type Entry struct {
	Key   string `@Ident`
	Value *Value `( ":" @@`
	Block *Block `| @@ )`
}

type Block struct {
	Parameters []*Value `{ @@ }`
	Entries    []*Entry `"{" { @@ } "}"`
}

type Config struct {
	Entries []*Entry `{ @@ }`
}

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

	parser, err := participle.Build(&Config{})
	if err != nil {
		panic(err.Error())
	}

	expr := &Config{}
	err = parser.ParseString(string(file), expr)
	if err != nil {
		panic(err.Error())
	}

	repr.Println(expr)
}
