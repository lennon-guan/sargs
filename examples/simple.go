package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lennon-guan/sargs"
)

func main() {
	var args struct {
		Language    string    `flag:"lang" usage:"生成项目的开发语言"`
		GenClient   bool      `flag:"gencli" default:"false"`
		Time        time.Time `flag:"time" default:"now" usage:"简单的时间字段"`
		ProjectName string    `pos:"0"`
		ProjectPath string    `pos:"1" default:""`
	}
	sargs.MustParse(&args)
	bs, _ := json.MarshalIndent(&args, "", "  ")
	fmt.Println(string(bs))
}
