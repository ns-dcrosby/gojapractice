package main

import (
	"fmt"

	"github.com/dop251/goja"
)

func main() {
	vm := goja.New()
	v, err := vm.RunString("'hello '+'world'")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Got: [%v]\n", v.Export().(string))
}
