package main

import (
	"fmt"
	"math"

	"github.com/lennon-guan/sargs"
)

type Sub struct {
	ToAbs bool `flag:"abs" default:"false"`
	A     int  `pos:"0"`
	B     int  `pos:"1" default:"1"`
}

func (sub *Sub) Run() {
	if sub.ToAbs {
		fmt.Println(int(math.Abs(float64(sub.A - sub.B))))
	} else {
		fmt.Println(sub.A - sub.B)
	}
}

func main() {
	sargs.RunApp(&Sub{})
}
