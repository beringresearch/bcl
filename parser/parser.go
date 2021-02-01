package parser

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2"
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

// Scan accepts a string and applies a lexer
func Scan(file string) (*Config, error) {
	expr := &Config{}
	parser := participle.MustBuild(&Config{}, participle.Unquote())

	err := parser.ParseString("", file, expr)
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
			if len(entry.Block.Entries) > 1 {
				return nil, errors.New("To many blocks in the system entry")
			}
			if !stringInSlice(entry.Block.Entries[0].Key, []string{"apt", "apk"}) {
				return nil, errors.New("Unsupported package manager <" + entry.Block.Entries[0].Key + ">. Only <apt> and <apk> are supported")
			}

			packages.Manager = entry.Block.Entries[0].Key
			packages.System = getStringArray(entry.Block.Entries[0].Value.Array)
			bravefile.SystemPackages = packages

		case "copy":
			copyCommandArray, err := parseCopyBlock(entry)
			if err != nil {
				return nil, err
			}
			bravefile.Copy = copyCommandArray

		case "run":
			runCommandArray, err := parseRunBlock(entry)
			if err != nil {
				return nil, err
			}
			bravefile.Run = runCommandArray

		case "service":
			for _, block := range entry.Block.Entries {
				if block.Key == "image" {
					service.Image = *block.Value.String
				} else if block.Key == "name" {
					service.Name = *block.Value.String
				} else if block.Key == "docker" {
					service.Docker = *block.Value.String
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
				} else if block.Key == "postdeploy" {
					for _, postdeploy := range block.Block.Entries {
						if postdeploy.Key == "copy" {
							postDeployCopy, err := parseCopyBlock(postdeploy)
							if err != nil {
								return nil, err
							}
							service.Postdeploy.Copy = postDeployCopy
						} else if postdeploy.Key == "run" {
							postDeployRun, err := parseRunBlock(postdeploy)
							if err != nil {
								return nil, err
							}
							service.Postdeploy.Run = postDeployRun
						} else {
							return nil, errors.New("Unsupported postdeploy key " + postdeploy.Key)
						}
					}
				} else if block.Key == "resources" {
					for _, resource := range block.Block.Entries {
						if resource.Key == "ram" {
							resources.RAM = *resource.Value.String
						} else if resource.Key == "cpu" {
							resources.CPU = *resource.Value.Integer
						} else if resource.Key == "gpu" {
							resources.GPU = *resource.Value.String
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

func parseRunBlock(entry *Entry) ([]validate.RunCommand, error) {
	var runCommand validate.RunCommand
	var runCommandArray []validate.RunCommand

	for _, block := range entry.Block.Entries {
		runCommand.Command = "sh"
		var arg []string

		if block.Value.Array != nil {
			arg = getStringArray(block.Value.Array)
			runCommand.Args = append([]string{"-c", block.Key}, arg...)
		} else {
			argString := *block.Value.String
			if strings.Contains(argString, "\n") {
				argString = strings.Replace(argString, "\n", " ", -1)
			}

			argString = "-c[DLM]" + block.Key + " " + argString
			arg = strings.Split(argString, "[DLM]")

			runCommand.Args = arg
		}

		runCommandArray = append(runCommandArray, runCommand)
	}
	return runCommandArray, nil
}

func parseCopyBlock(entry *Entry) ([]validate.CopyCommand, error) {
	var copyCommand validate.CopyCommand
	var copyCommandArray []validate.CopyCommand

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
	return copyCommandArray, nil
}

func getStringArray(array []*Value) []string {
	a := make([]string, len(array))
	for i, v := range array {
		a[i] = *v.String
	}
	return a
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
