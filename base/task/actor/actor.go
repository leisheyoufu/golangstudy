package actor

import (
	"fmt"
	"reflect"
)

type StringActor struct{}

func (actor *StringActor) OnReceive(in interface{}) {
	if reflect.TypeOf(in).Kind() == reflect.String {
		fmt.Printf("%s:%s\n", actor.GetName(), in.(string))
	}
}

func (actor *StringActor) GetName() string {
	return "StringActor"
}

type IntActor struct{}

func (actor *IntActor) OnReceive(in interface{}) {
	if reflect.TypeOf(in).Kind() == reflect.Int {
		fmt.Printf("%s:%d\n", actor.GetName(), in.(int))
	}
}

func (actor *IntActor) GetName() string {
	return "IntegerActor"
}
