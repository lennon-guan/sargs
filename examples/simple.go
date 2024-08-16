package main

import (
	"encoding/json"
	"fmt"

	"github.com/lennon-guan/sargs"
)

func main() {
	var args struct {
		Language    string `flag:"lang" usage:"生成项目的开发语言"`
		GenClient   bool   `flag:"gencli" default:"false"`
		ProjectName string `pos:"0"`
		ProjectPath string `pos:"1" default:""`
	}
	sargs.MustParse(&args)
	bs, _ := json.MarshalIndent(&args, "", "  ")
	fmt.Println(string(bs))
}
