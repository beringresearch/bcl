package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// IConfigLog ...
type IConfigLog interface {
}

//Config ...
type Config struct {
	filename       string
	queue          []reflect.Value
	current        reflect.Value
	searchVal      bool
	searchKey      bool
	inBlock        int
	inInclude      int
	canSkip        bool
	skip           bool
	bkQueue        []bool
	bkMulti        bool
	mapKey         reflect.Value
	setVar         bool
	searchVar      bool
	searchVarBlock bool
	vars           []map[string]string
	currentVar     map[string]string
}

// New ...
func New(filename string) *Config {
	return &Config{filename: filename}
}

// Unmarshal ...
func (conf *Config) Unmarshal(v interface{}) error {
	rev := reflect.ValueOf(v)
	if rev.Kind() != reflect.Ptr {
		err := errors.New("non-pointer passed to Unmarshal")
		return err
	}
	conf.current = rev.Elem()
	conf.inSearchKey()
	conf.inBlock = 0
	conf.inInclude = 0
	conf.currentVar = make(map[string]string)
	conf.searchVar = false
	conf.setVar = false
	conf.searchVarBlock = false
	return conf.parse()
}

// Reload ...
func (conf *Config) Reload() error {
	return conf.parse()
}

func (conf *Config) parse() error {
	var err error
	var s bytes.Buffer
	var vs bytes.Buffer
	var vsb bytes.Buffer
	if _, err = os.Stat(conf.filename); os.IsNotExist(err) {
		return err
	}

	var fp *os.File
	fp, err = os.Open(conf.filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)
	var b byte
	for err == nil {
		b, err = reader.ReadByte()
		if err == bufio.ErrBufferFull {
			return nil
		}

		if conf.canSkip && b == '#' {
			reader.ReadLine()
			continue
		}
		if conf.canSkip && b == '/' {
			if conf.skip {
				reader.ReadLine()
				conf.skip = false
				continue
			}
			conf.skip = true
			continue
		}
		if conf.searchKey {
			if conf.delimiter(b) {
				if s.Len() > 0 {
					conf.inSearchVal()
					if strings.Compare(s.String(), "include") == 0 {
						s.Reset()
						conf.inInclude++
						if conf.inInclude > 100 {
							return errors.New("too many include, exceeds 100 limit")
						}
						continue
					} else if strings.Compare(s.String(), "set") == 0 {
						s.Reset()
						conf.setVar = true
						continue
					}
					conf.getElement(s.String())
					s.Reset()
				}
				continue
			}
		}

		if b == '{' && !conf.searchVar && vs.Len() == 0 {
			if err := conf.createBlock(&s); err != nil {
				return err
			}
			continue
		}

		if conf.searchKey && b == '}' && conf.inBlock > 0 {
			conf.closeBlock(&s)
			continue
		}

		if conf.searchVal {
			if b == '$' {
				conf.searchVar = true
				vs.Reset()
				vsb.Reset()
				vsb.WriteByte(b)
				continue
			}

			if conf.searchVar {
				if b == '{' {
					if vs.Len() == 0 {
						conf.searchVarBlock = true
						vsb.WriteByte(b)
						continue
					}
					if !conf.searchVarBlock {
						if !conf.replace(&s, &vs) {
							s.Write(vsb.Bytes())
						}

						// is block?
						if err := conf.createBlock(&s); err != nil {
							return err
						}
					}
				}
				// Is space
				if conf.delimiter(b) {
					if !conf.replace(&s, &vs) {
						s.Write(vsb.Bytes())
					}
				}
			}

			if conf.searchVarBlock && b == '}' {
				conf.searchVarBlock = false
				// replace $???
				conf.searchVar = false
				if !conf.replace(&s, &vs) {
					vsb.WriteByte(b)
					s.Write(vsb.Bytes())
				}
				continue
			}

			if b == ';' {
				//	copy to this.current
				conf.inSearchKey()
				if conf.searchVar {
					if conf.searchVarBlock {
						return errors.New(vsb.String() + " is not terminated by }")
					}
					conf.searchVar = false
					conf.searchVarBlock = false
					if !conf.replace(&s, &vs) {
						s.Write(vsb.Bytes())
					}
				}

				// set to map
				if conf.setVar {
					sf := strings.Fields(s.String())
					if len(sf) != 2 {
						return errors.New("Invalid Config")
					}
					conf.currentVar[sf[0]] = sf[1]
					conf.setVar = false
					s.Reset()
					continue
				}

				if conf.inInclude > 0 {
					conf.filename = strings.TrimSpace(s.String())
					s.Reset()
					conf.inInclude--
					files, err := filepath.Glob(conf.filename)
					if err != nil {
						return err
					}
					for _, file := range files {
						conf.filename = file
						if err := conf.parse(); err != nil {
							return err
						}
					}
					continue
				}

				err := conf.set(s.String())
				if err != nil {
					return err
				}

				s.Reset()
				conf.popElement()
				continue
			} else if conf.searchVar { // if b == ';'
				vs.WriteByte(b)
				vsb.WriteByte(b)
				if !conf.searchVarBlock {
					conf.replace(&s, &vs)
				}
				continue
			}
		}

		s.WriteByte(b)
	}

	if !conf.searchKey && conf.inBlock > 0 {
		return fmt.Errorf("Invalid config file")
	}

	return nil
}

func (conf *Config) set(s string) error {
	s = strings.TrimSpace(s)
	if conf.current.Kind() == reflect.Ptr {
		if conf.current.IsNil() {
			conf.current.Set(reflect.New(conf.current.Type().Elem()))
		}
		conf.current = conf.current.Elem()
	}

	switch conf.current.Kind() {
	case reflect.String:
		conf.current.SetString(conf.clearQuoted(s))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		itmp, err := strconv.ParseInt(s, 10, conf.current.Type().Bits())
		if err != nil {
			return err
		}
		conf.current.SetInt(itmp)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		itmp, err := strconv.ParseUint(s, 10, conf.current.Type().Bits())
		if err != nil {
			return err
		}
		conf.current.SetUint(itmp)
	case reflect.Float32, reflect.Float64:
		ftmp, err := strconv.ParseFloat(s, conf.current.Type().Bits())
		if err != nil {
			return err
		}
		conf.current.SetFloat(ftmp)
	case reflect.Bool:
		if s == "yes" || s == "on" {
			conf.current.SetBool(true)
		} else if s == "no" || s == "off" {
			conf.current.SetBool(false)
		} else {
			btmp, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			conf.current.SetBool(btmp)
		}
	case reflect.Slice:
		sf, err := conf.splitQuoted(s)
		if err != nil {
			return err
		}
		for _, sv := range sf {
			n := conf.current.Len()
			conf.current.Set(reflect.Append(conf.current, reflect.Zero(conf.current.Type().Elem())))
			conf.pushElement(conf.current.Index(n))
			conf.set(sv)
			conf.popElement()
		}
	case reflect.Map:
		if conf.current.IsNil() {
			conf.current.Set(reflect.MakeMap(conf.current.Type()))
		}

		sf, err := conf.splitQuoted(s)
		if err != nil {
			return err
		}
		if len(sf) != 2 {
			return errors.New("Invalid Config")
		}
		var v reflect.Value
		v = reflect.New(conf.current.Type().Key())
		conf.pushElement(v)
		conf.set(sf[0])
		key := conf.current
		conf.popElement()
		v = reflect.New(conf.current.Type().Elem())
		conf.pushElement(v)
		conf.set(sf[1])
		val := conf.current
		conf.popElement()

		conf.current.SetMapIndex(key, val)
	default:
		return fmt.Errorf("Invalid Type:%s", conf.current.Kind())
	}
	return nil
}

func (conf *Config) replace(s *bytes.Buffer, vs *bytes.Buffer) bool {
	if vs.Len() == 0 {
		return false
	}

	for k, v := range conf.currentVar {
		if strings.Compare(k, vs.String()) == 0 {
			// found
			conf.searchVar = false
			s.WriteString(v)
			vs.Reset()
			return true
		}
	}

	for i := len(conf.vars) - 1; i >= 0; i-- {
		for k, v := range conf.vars[i] {
			if strings.Compare(k, vs.String()) == 0 {
				s.WriteString(v)
				conf.searchVar = false
				vs.Reset()
				return true
			}
		}
	}

	return false
}

func (conf *Config) createBlock(s *bytes.Buffer) error {
	// fixed { be close to key like server{
	if conf.searchKey && s.Len() > 0 {
		conf.getElement(s.String())
		s.Reset()
		conf.inSearchVal()
	}

	// vars
	vars := make(map[string]string)
	conf.vars = append(conf.vars, conf.currentVar)
	conf.currentVar = vars

	conf.inBlock++
	//	slice or map?
	conf.bkQueue = append(conf.bkQueue, conf.bkMulti)
	conf.bkMulti = false
	if conf.searchVal && s.Len() > 0 && conf.current.Kind() == reflect.Map {
		conf.bkMulti = true
		if conf.current.IsNil() {
			conf.current.Set(reflect.MakeMap(conf.current.Type()))
		}
		var v reflect.Value
		v = reflect.New(conf.current.Type().Key())
		conf.pushElement(v)
		err := conf.set(s.String())
		if err != nil {
			return err
		}
		conf.mapKey = conf.current
		conf.popElement()
		val := reflect.New(conf.current.Type().Elem())
		conf.pushElement(val)
	}

	if conf.current.Kind() == reflect.Slice {
		conf.pushMultiBlock()
		n := conf.current.Len()
		if conf.current.Type().Elem().Kind() == reflect.Ptr {
			conf.current.Set(reflect.Append(conf.current, reflect.New(conf.current.Type().Elem().Elem())))
		} else {
			conf.current.Set(reflect.Append(conf.current, reflect.Zero(conf.current.Type().Elem())))
		}
		conf.pushElement(conf.current.Index(n))
	}
	conf.inSearchKey()
	s.Reset()
	return nil
}

func (conf *Config) closeBlock(s *bytes.Buffer) {
	if conf.bkMulti {
		val := conf.current
		conf.popElement()
		if conf.current.Kind() == reflect.Map {
			conf.current.SetMapIndex(conf.mapKey, val)
		}
	}
	conf.popMultiBlock()

	// vars
	conf.currentVar = conf.vars[len(conf.vars)-1]
	conf.vars = conf.vars[:len(conf.vars)-1]

	conf.inBlock--
	conf.popElement()
	conf.inSearchKey()
}

func (conf *Config) inSearchKey() {
	conf.searchVal = false
	conf.searchKey = true
	conf.canSkip = true
}

func (conf *Config) inSearchVal() {
	conf.searchKey = false
	conf.searchVal = true
	conf.canSkip = false
}

func (conf *Config) getElement(s string) {
	s = strings.TrimSpace(s)
	if conf.current.Kind() == reflect.Ptr {
		if conf.current.IsNil() {
			conf.current.Set(reflect.New(conf.current.Type().Elem()))
		}
		conf.current = conf.current.Elem()
	}
	conf.queue = append(conf.queue, conf.current)
	conf.current = conf.current.FieldByName(strings.Title(s))
}

func (conf *Config) pushElement(v reflect.Value) {
	conf.queue = append(conf.queue, conf.current)
	conf.current = v
}

func (conf *Config) popElement() {
	conf.current = conf.queue[len(conf.queue)-1]
	conf.queue = conf.queue[:len(conf.queue)-1]
}

func (conf *Config) delimiter(b byte) bool {
	return unicode.IsSpace(rune(b))
}

func (conf *Config) pushMultiBlock() {
	conf.bkQueue = append(conf.bkQueue, conf.bkMulti)
	conf.bkMulti = true
}

func (conf *Config) popMultiBlock() {
	conf.bkMulti = conf.bkQueue[len(conf.bkQueue)-1]
	conf.bkQueue = conf.bkQueue[:len(conf.bkQueue)-1]
}

func (conf *Config) clearQuoted(s string) string {
	s = strings.TrimSpace(s)
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}

	return s
}

func (conf *Config) splitQuoted(s string) ([]string, error) {
	var sq []string
	s = strings.TrimSpace(s)
	var lastSpace bool = true
	var needSpace bool = true
	var dQuote bool = false
	var sQuote bool = false
	var quote bool = false
	var ch byte
	var vs bytes.Buffer

	for i := 0; i < len(s); i++ {
		ch = s[i]

		if quote {
			quote = false
			vs.WriteByte(ch)
			continue
		}

		if ch == '\\' {
			quote = true
			lastSpace = false
			continue
		}

		if lastSpace {
			lastSpace = false
			switch ch {
			case '"':
				dQuote = true
				needSpace = false
				continue
			case '\'':
				sQuote = true
				needSpace = false
				continue
			case ' ':
				lastSpace = true
				continue
			}
			vs.WriteByte(ch)
		} else {
			if needSpace && conf.delimiter(ch) {
				if vs.Len() > 0 {
					sq = append(sq, vs.String())
				}
				vs.Reset()
				lastSpace = true
				continue
			}

			if dQuote {
				if ch == '"' {
					dQuote = false
					needSpace = true
					continue
				}
			} else if sQuote {
				if ch == '\'' {
					sQuote = false
					needSpace = true
					continue
				}
			}

			vs.WriteByte(ch)
		}
	}

	if quote || sQuote || dQuote {
		return nil, fmt.Errorf("Invalid value: %v", s)
	}

	if vs.Len() > 0 {
		sq = append(sq, vs.String())
	}

	return sq, nil
}
