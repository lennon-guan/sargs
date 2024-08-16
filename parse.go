package sargs

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

type parseArgFunc = func([]string) error

type parseTask struct {
	flags    *flag.FlagSet
	setArgs  []parseArgFunc
	required map[string]struct{}
}

func parseFlagSet(a any, task *parseTask) error {
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return ErrNotPtrToStruct
	}
	v = v.Elem()
	t := v.Type()
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
				return fmt.Errorf("%w: field %s", ErrUnsupportedFieldType, ft.Name)
			}
		} else if posStr := tag.Get("pos"); posStr != "" {
			if pos, err := strconv.ParseUint(posStr, 10, 64); err != nil {
				return fmt.Errorf("%w: field: %s pos %s", ErrInvalidArgPos, ft.Name, posStr)
			} else if parser, found := valueParsers[ft.Type.Kind()]; !found {
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
