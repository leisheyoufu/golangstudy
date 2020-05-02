package main

import (
	"fmt"
)

func RangeArray() {
	var a = [...]int{1, 2, 3, 4, 5, 6}
	for _, v := range a {
		v = v + 100
	}
	fmt.Println("Not changed")
	for _, v := range a {
		fmt.Printf("%d ", v)
	}
	fmt.Printf("\n")
	for i, _ := range a {
		a[i] = a[i] + 100
	}
	fmt.Printf("Changed by index\n")
	for _, v := range a {
		fmt.Printf("%d ", v)
	}
	fmt.Printf("\nAssign value\n")
	copyA := make([]int, 6)
	for i, v := range a {
		copyA[i] = v
	}
	for _, v := range copyA {
		fmt.Printf("%d ", v)
	}
	fmt.Printf("\nAssign reference from value\n")
	aReference := make([]*int, 6)
	for i, v := range a {
		aReference[i] = &v
	}
	for _, v := range aReference {
		fmt.Printf("%d ", *v)
	}
	fmt.Printf("\nAssign reference from index\n")
	for i, _ := range a {
		aReference[i] = &a[i]
	}
	for _, v := range aReference {
		fmt.Printf("%d ", *v)
	}
}

func main() {
	RangeArray()
}
