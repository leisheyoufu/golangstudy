package main

import (
	"fmt"
)

type Human struct {
	name string
}

func (self *Human) SetPointerName(name string) {
	self.name = name
}

func (self Human) SetInstanceName(name string) {
	self.name = name
}

func main() {
	human := &Human{name: "name"}
	human.SetPointerName("pointer name")
	fmt.Printf("Pointer name = %s\n", human.name)

	human = &Human{name: "name"}
	human.SetInstanceName("instance name")
	fmt.Printf("Instance name is not changed: = %s\n", human.name)

	fmt.Println("-------Init Object -------------")
	human2 := Human{name: "name"}
	human2.SetPointerName("pointer name")
	fmt.Printf("Pointer name = %s\n", human2.name)

	human2 = Human{name: "name"}
	human2.SetInstanceName("instance name")
	fmt.Printf("Instance name is not changed: = %s\n", human2.name)
}
