package sargs

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/lennon-guan/smartime"
)

type parseArgFunc = func([]string) error

type parseTask struct {
	flags    *flag.FlagSet
	setArgs  []parseArgFunc
	required map[string]struct{}
}

type commonValue[T any] struct {
	p      *T
	parser func(string, unsafe.Pointer) error
}

func newCommonValue[T any](p *T, defaultValue T, parser func(string, unsafe.Pointer) error) *commonValue[T] {
	*p = defaultValue
	return &commonValue[T]{
		p:      p,
		parser: parser,
	}
}

func (v *commonValue[T]) String() string {
	return fmt.Sprint(*v.p)
}

func (v *commonValue[T]) Set(s string) error {
	return v.parser(s, unsafe.Pointer(v.p))
}

func parseFlagSet(a any, task *parseTask) error {
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return ErrNotPtrToStruct
	}
	v = v.Elem()
	t := v.Type()
	bt := smartime.NowBase()
	for i := 0; i < t.NumField(); i++ {
		var (
			err                error
			fv                 = v.Field(i)
			fp                 = fv.Addr().UnsafePointer()
			ft                 = t.Field(i)
			tag                = ft.Tag
			usage              = tag.Get("usage")
			defVal, hasDefault = tag.Lookup("default")
		)
		if name := tag.Get("flag"); name != "" {
			if !hasDefault {
				task.required[name] = struct{}{}
			}
			switch ft.Type.Kind() {
			case reflect.String:
				task.flags.StringVar((*string)(fp), name, defVal, usage)
			case reflect.Int:
				var dv int
				if hasDefault {
					if dv, err = strconv.Atoi(defVal); err != nil {
						return fmt.Errorf("%w: field: %s defaultValue: %s", ErrInvalidDefaultValue, ft.Name, defVal)
					}
				}
				task.flags.IntVar((*int)(fp), name, dv, usage)
			case reflect.Int64:
				var dv int64
				if hasDefault {
					if dv, err = strconv.ParseInt(defVal, 10, 64); err != nil {
						return fmt.Errorf("%w: field: %s defaultValue: %s", ErrInvalidDefaultValue, ft.Name, defVal)
					}
				}
				task.flags.Int64Var((*int64)(fp), name, dv, usage)
			case reflect.Uint:
				var dv uint64
				if hasDefault {
					if dv, err = strconv.ParseUint(defVal, 10, 64); err != nil {
						return fmt.Errorf("%w: field: %s defaultValue: %s", ErrInvalidDefaultValue, ft.Name, defVal)
					}
				}
				task.flags.UintVar((*uint)(fp), name, uint(dv), usage)
			case reflect.Uint64:
				var dv uint64
				if hasDefault {
					if dv, err = strconv.ParseUint(defVal, 10, 64); err != nil {
						return fmt.Errorf("%w: field: %s defaultValue: %s", ErrInvalidDefaultValue, ft.Name, defVal)
					}
				}
				task.flags.Uint64Var((*uint64)(fp), name, dv, usage)
			case reflect.Bool:
				var dv bool
				switch strings.ToLower(defVal) {
				case "true", "yes", "1":
					dv = true
				case "", "false", "no", "0":
					dv = false
				default:
					return fmt.Errorf("%w: field: %s defaultValue: %s", ErrInvalidDefaultValue, ft.Name, defVal)
				}
				task.flags.BoolVar((*bool)(fp), name, dv, usage)
			default:
				switch fv.Interface().(type) {
				case time.Time:
					var dv time.Time
					if hasDefault {
						if dv, err = bt.ParseTime(defVal); err != nil {
							return fmt.Errorf("%w: field: %s defaultValue: %s", ErrInvalidDefaultValue, ft.Name, defVal)
						}
					}
					task.flags.Var(
						newCommonValue((*time.Time)(fp), dv, func(s string, p unsafe.Pointer) error {
							if t, err := bt.ParseTime(s); err != nil {
								return err
							} else {
								*(*time.Time)(p) = t
								return nil
							}
						}),
						name, usage)
				case time.Duration:
					var dv time.Duration
					if hasDefault {
						if dv, err = time.ParseDuration(defVal); err != nil {
							return fmt.Errorf("%w: field: %s defaultValue: %s", ErrInvalidDefaultValue, ft.Name, defVal)
						}
					}
					task.flags.DurationVar((*time.Duration)(fp), name, dv, usage)
				default:
					return fmt.Errorf("%w: field %s", ErrUnsupportedFieldType, ft.Name)
				}
			}
		} else if posStr := tag.Get("pos"); posStr != "" {
			if pos, err := strconv.ParseUint(posStr, 10, 64); err != nil {
				return fmt.Errorf("%w: field: %s pos %s", ErrInvalidArgPos, ft.Name, posStr)
			} else if parser := getValueParser(fv); parser == nil {
				return fmt.Errorf("%w: field %s", ErrUnsupportedFieldType, ft.Name)
			} else if hasDefault {
				task.setArgs = append(task.setArgs, makeParseArgFuncWithDefault(pos, fp, parser, defVal))
			} else {
				task.setArgs = append(task.setArgs, makeParseArgFunc(pos, fp, parser))
			}
		}
	}
	return nil
}

func makeParseArgFunc(pos uint64, ptr unsafe.Pointer, parser func(string, unsafe.Pointer) error) func([]string) error {
	i := int(pos)
	return func(args []string) error {
		if i >= len(args) {
			return ErrNotEnoughArgs
		}
		return parser(args[i], ptr)
	}
}

func makeParseArgFuncWithDefault(pos uint64, ptr unsafe.Pointer, parser func(string, unsafe.Pointer) error, dv string) func([]string) error {
	i := int(pos)
	return func(args []string) error {
		if i >= len(args) {
			return parser(dv, ptr)
		}
		return parser(args[i], ptr)
	}
}

var valueParsers = map[reflect.Kind]func(string, unsafe.Pointer) error{
	reflect.Int: func(s string, p unsafe.Pointer) (err error) {
		*(*int)(p), err = strconv.Atoi(s)
		return
	},
	reflect.String: func(s string, p unsafe.Pointer) error {
		*(*string)(p) = s
		return nil
	},
}

func getValueParser(f reflect.Value) func(string, unsafe.Pointer) error {
	if p, ok := valueParsers[f.Kind()]; ok {
		return p
	}
	switch f.Interface().(type) {
	case time.Duration:
		return func(s string, p unsafe.Pointer) error {
			du, err := time.ParseDuration(s)
			if err != nil {
				return err
			}
			*(*time.Duration)(p) = du
			return nil
		}
	case time.Time:
		return func(s string, p unsafe.Pointer) error {
			t, err := parseTime(s)
			if err != nil {
				return err
			}
			*(*time.Time)(p) = t
			return nil
		}
	}
	return nil
}

func parseTime(s string) (t time.Time, err error) {
	var ts int64
	if strings.HasPrefix(s, "+") { // Relative time: duration after now
		if du, err := time.ParseDuration(s[1:]); err != nil {
			return t, err
		} else {
			t = time.Now().Add(du)
			return t, nil
		}
	} else if strings.HasPrefix(s, "0") { // Relative time: duration before now
		if du, err := time.ParseDuration(s[1:]); err != nil {
			return t, err
		} else {
			t = time.Now().Add(-du)
			return t, err
		}
	} else if s == "now" {
		t = time.Now()
		return
	} else { // Absolute time
		switch len(s) {
		case 6: // yymmdd
			t, err = time.Parse("060102", s)
		case 8: // YYYYmmdd yy-mm-dd
			if t, err = time.Parse("20060102", s); err == nil {
			} else if t, err = time.Parse("06-01-02", s); err == nil {
			}
		case 10: // YYYY-mm-dd timestamp(to second)
			if t, err = time.Parse("2006-01-02", s); err == nil {
			} else if ts, err = strconv.ParseInt(s, 10, 64); err == nil {
				t = time.Unix(ts, 0)
			}
		case 13: // timestamp(to millisecond)
			if ts, err = strconv.ParseInt(s, 10, 64); err == nil {
				t = time.Unix(ts/1000, (ts%1000)*1e6)
			}
		case 14: // YYYYmmddHHMMSS
			t, err = time.Parse("20060102150405", s)
		case 19: // YYYY-mm-dd HH:MM:SS
			t, err = time.Parse("2006-01-02 15:04:05", s)
		default:
			err = fmt.Errorf("unsupported time format: %s", s)
		}
		return t, err
	}
}
