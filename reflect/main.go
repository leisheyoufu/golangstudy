package main

import (
	"fmt"
	"reflect"
)

type Human struct {
	name   string
	age    int
	weight int
}

func (h *Human) SayHello() {
	fmt.Println("Hello reflect")
}

func main() {
	t := reflect.ValueOf(Human{}).Type()
	//    h := reflect.New(t).Elem()
	// new return address pointer
	h := reflect.New(t).Interface()
	fmt.Println(h)
	hh := h.(*Human)
	if hh.name == "" {
		fmt.Println("Name is nil")
	}
	fmt.Println(hh)
	hh.SayHello()
	hh.age = 123
	hh.name = "abc"
	hh.weight = 345
	hh.SayHello()

	i := Human{"Emp", 25, 120}
	fmt.Println(reflect.TypeOf(i).Field(0).Type)
	fmt.Println(reflect.ValueOf(i).Field(1))
	//	reflect.ValueOf(i).Field(1).Elem().SetInt(88)
	//	fmt.Println(reflect.ValueOf(i).Field(1))
}
