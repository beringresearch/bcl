package parser

import (
	"errors"
	"fmt"
	"github.com/alecthomas/participle"
	"github.com/beringresearch/bcl/validate"
	"strings"
)

// Bool value
type Bool bool

// Capture boolean value
func (b *Bool) Capture(v []string) error { *b = v[0] == "true"; return nil }

// Value structure
type Value struct {
	Boolean    *Bool    `  @("true"|"false")`
	Identifier *string  `| @Ident { @"." @Ident }`
	String     *string  `| @(String|Char|RawString)`
	Integer    *int64   `| @(Int)`
	Float      *float64 `| @(Float)`
	Array      []*Value `| "[" { @@ [ "," ] } "]"`
}

// GoString conversion
func (l *Value) GoString() string {
	switch {
	case l.Boolean != nil:
		return fmt.Sprintf("%v", *l.Boolean)
	case l.Identifier != nil:
		return fmt.Sprintf("`%s`", *l.Identifier)
	case l.String != nil:
		return fmt.Sprintf("%q", *l.String)
	case l.Integer != nil:
		return fmt.Sprintf("%v", *l.Integer)
	case l.Float != nil:
		return fmt.Sprintf("%v", *l.Float)
	case l.Array != nil:
		out := []string{}
		for _, v := range l.Array {
			out = append(out, v.GoString())
		}
		return fmt.Sprintf("[]*Value{ %s }", strings.Join(out, ", "))
	}
	panic("??")
}

// Entry is a BCL minimal functioning unit
type Entry struct {
	Key   string `@Ident`
	Value *Value `( ":" @@`
	Block *Block `| @@)`
}

// Block is BCL's minimal functioning structure
type Block struct {
	Parameters []*Value `{ @@ }`
	Entries    []*Entry `"{" { @@ } "}"`
}

// Config represents the global BCL configuration
type Config struct {
	Entries []*Entry `{ @@ }`
}

// NewConfig ..
func NewConfig() *Config {
	return &Config{}
}

func getStringArray(array []*Value) []string {
	a := make([]string, len(array))
	for i, v := range array {
		a[i] = *v.String
	}
	return a
}

// Scan accepts a string and applies a lexer
func Scan(file string) (*Config, error) {
	expr := &Config{}
	parser, err := participle.Build(&Config{})

	if err != nil {
		return &Config{}, err
	}

	err = parser.ParseString(file, expr)
	if err != nil {
		return &Config{}, err
	}
	return expr, nil
}

// Parse ..
func Parse(file *Config) (*validate.Bravefile, error) {
	bravefile := validate.NewBravefile()

	var imageDescription validate.ImageDescription
	var packages validate.Packages
	var service validate.Service
	var resources validate.Resources

	var copyCommandArray []validate.CopyCommand
	var runCommandArray []validate.RunCommand

	for _, entry := range file.Entries {
		switch entry.Key {
		case "base":
			for _, block := range entry.Block.Entries {
				if block.Key == "location" {
					imageDescription.Location = *block.Value.String
				} else if block.Key == "image" {
					imageDescription.Image = *block.Value.String
				} else {
					return nil, errors.New("Unsupported base key " + block.Key)
				}
				bravefile.Base = imageDescription
			}

		case "system":
			packages.Manager = entry.Block.Entries[0].Key
			packages.System = getStringArray(entry.Block.Entries[0].Value.Array)
			bravefile.SystemPackages = packages

		case "copy":
			var copyCommand validate.CopyCommand
			for _, block := range entry.Block.Entries {
				for _, file := range block.Block.Entries {
					if file.Key == "source" {
						copyCommand.Source = *file.Value.String
					} else if file.Key == "target" {
						copyCommand.Target = *file.Value.String
					} else if file.Key == "action" {
						copyCommand.Action = *file.Value.String
					} else {
						return nil, errors.New("Unsupported copy key " + file.Key)
					}
				}

				copyCommandArray = append(copyCommandArray, copyCommand)
			}
			bravefile.Copy = copyCommandArray

		case "run":
			var runCommand validate.RunCommand
			for _, block := range entry.Block.Entries {
				runCommand.Command = block.Key
				if block.Value.String == nil {
					arg := getStringArray(block.Value.Array)
					runCommand.Args = arg
				} else {
					if strings.Contains(*block.Value.String, "\n") {
						runCommand.Args = []string{*block.Value.String}
					} else {
						runCommand.Args = strings.Split(*block.Value.String, " ")
					}
				}

				runCommandArray = append(runCommandArray, runCommand)
			}
			bravefile.Run = runCommandArray

		case "service":
			for _, block := range entry.Block.Entries {
				if block.Key == "name" {
					service.Name = *block.Value.String
				} else if block.Key == "version" {
					service.Version = *block.Value.String
				} else if block.Key == "ip" {
					service.IP = *block.Value.String
				} else if block.Key == "ports" {
					if block.Value.String == nil {
						service.Ports = getStringArray(block.Value.Array)
					} else {
						service.Ports = []string{*block.Value.String}
					}
				} else if block.Key == "resources" {
					for _, resource := range block.Block.Entries {
						if resource.Key == "ram" {
							resources.RAM = *resource.Value.String
						} else if resource.Key == "cpu" {
							resources.CPU = *resource.Value.Integer
						} else if resource.Key == "gpu" {
							resources.GPU = bool(*resource.Value.Boolean)
						} else {
							return nil, errors.New("Unsupported resource key " + resource.Key)
						}
					}
					service.Resources = resources
				}
			}
			bravefile.PlatformService = service

		default:
			return nil, errors.New("Unsupported Entry key " + entry.Key)
		}
	}
	return bravefile, nil
}
