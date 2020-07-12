package main

import (
	"fmt"
	"html/template"
	"os"
)

type Inventory struct {
	Material string
	Count    uint
}

func add(left int, right int) int {
	return left + right
}

func multi(left int, right int) int {
	return left * right
}

func ParseCustomeFunc() {
	tmpl, err := template.New("test").Funcs(template.FuncMap{
		"add":   add,
		"multi": multi,
	}).Parse("{{add 1 2}} {{multi 1 5}}\n")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, nil)
	if err != nil {
		panic(err)
	}
}

func Inventory1() {
	sweaters := Inventory{"wool", 17}
	tmpl, err := template.New("test").Parse("{{.Count}} items are made of {{.Material}}\n")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, sweaters)
	if err != nil {
		panic(err)
	}
}

func Inventory2() {
	sweaters := Inventory{"wool", 17}
	fmt.Printf("{{- }} trim space at beginning while {{ -}} trim space at end\n")
	tmpl, err := template.New("test").Parse("{{.Count -}} items are made of {{- .Material}}\n")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, sweaters)
	if err != nil {
		panic(err)
	}
}

func main() {
	Inventory1()
	Inventory2()
	ParseCustomeFunc()
}
