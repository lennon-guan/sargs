package sargs

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
)

type (
	Command interface {
		Run()
	}
	CommandWithName interface {
		Command
		Name() string
	}
)

func RunApp(commands ...Command) {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "command name is missing")
		return
	}
	cmd := os.Args[1]
	for _, command := range commands {
		if genName(command) == cmd {
			must(ParseArgs(cmd, os.Args[2:], command))
			command.Run()
			return
		}
	}
	fmt.Fprintf(os.Stderr, "commands %s not found\n", cmd)
}

func genName(cmd Command) string {
	if cn, is := cmd.(CommandWithName); is {
		fmt.Println("->", cn.Name())
		return cn.Name()
	}
	typeName := reflect.TypeOf(cmd).Elem().Name()
	words := regexp.MustCompile(`[A-Z][a-z0-9]*`).FindAllString(typeName, -1)
	return strings.ToLower(strings.Join(words, "-"))
}
