package sargs

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func ParseArgs(name string, args []string, to any) error {
	task := parseTask{
		flags:    flag.NewFlagSet(name, flag.ContinueOnError),
		required: make(map[string]struct{}),
	}
	if err := parseFlagSet(to, &task); err != nil {
		return err
	} else if err := task.flags.Parse(args); err != nil {
		return err
	} else {
		// check required
		if len(task.required) > 0 {
			fs := make(map[string]struct{})
			task.flags.Visit(func(f *flag.Flag) {
				fs[f.Name] = struct{}{}
			})
			for r := range task.required {
				if _, found := fs[r]; !found {
					return fmt.Errorf("%w: %s", ErrMissingFlag, r)
				}
			}
		}
		// fill arguments
		fsArgs := task.flags.Args()
		for _, p := range task.setArgs {
			if err := p(fsArgs); err != nil {
				return err
			}
		}
	}
	return nil
}

func Parse(to any) error {
	return ParseArgs(filepath.Base(os.Args[0]), os.Args[1:], to)
}

func must(err error) {
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		panic(err)
	}
}

func MustParse(to any) {
	must(Parse(to))
}
