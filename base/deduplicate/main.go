package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
)

type User struct {
	name     string
	password string
	id       int
	books    []string
}

type Game struct {
	name    string
	version int
}

// InArray will search element inside array with any type.
// Will return boolean and index for matched element.
// True and index more than 0 if element is exist.
// needle is element to search, haystack is slice of value to be search.
func InArray(needle interface{}, haystack interface{}) (exists bool, index int) {
	exists = false
	index = -1
	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Array, reflect.Slice:
		s := reflect.ValueOf(haystack)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(needle, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}
	return
}

// can not work if struct contains slice member
func Dedumplicate(data interface{}) interface{} {
	inArr := reflect.ValueOf(data)
	if inArr.Kind() != reflect.Slice && inArr.Kind() != reflect.Array {
		return data
	}

	existMap := make(map[interface{}]bool)
	outArr := reflect.MakeSlice(inArr.Type(), 0, inArr.Len())

	for i := 0; i < inArr.Len(); i++ {
		iVal := inArr.Index(i)

		if _, ok := existMap[iVal.Interface()]; !ok {
			outArr = reflect.Append(outArr, inArr.Index(i))
			existMap[iVal.Interface()] = true
		}
	}

	return outArr.Interface()
}

func Dedumplicate2(data interface{}) interface{} {
	inArr := reflect.ValueOf(data)
	if inArr.Kind() != reflect.Slice && inArr.Kind() != reflect.Array {
		return data
	}
	n := inArr.Len()
	ret := reflect.MakeSlice(inArr.Type(), 0, 5)
	for i := 0; i < n; i++ {
		found := false
		for j := i + 1; j < n; j++ {
			if reflect.DeepEqual(inArr.Index(i).Interface(), inArr.Index(j).Interface()) {
				found = true
				break
			}
		}
		if !found {
			ret = reflect.Append(ret, inArr.Index(i))
		}
	}
	return ret.Interface()
}

func Remove(targets interface{}, items interface{}) (interface{}, error) {
	removeIndexs := make([]int, 0)
	targetArray := reflect.ValueOf(targets)
	if targetArray.Kind() != reflect.Array && targetArray.Kind() != reflect.Slice {
		return targets, nil
	}
	itemArray := reflect.ValueOf(items)
	if itemArray.Kind() != reflect.Array && itemArray.Kind() != reflect.Slice {
		return targets, nil
	}

	for i := 0; i < itemArray.Len(); i++ {
		found := false

		for j := 0; j < targetArray.Len(); j++ {
			if reflect.DeepEqual(itemArray.Index(i).Interface(), targetArray.Index(j).Interface()) == true {
				removeIndexs = append(removeIndexs, j)
				found = true
			}
		}
		if !found {
			b, err := json.Marshal(itemArray.Index(i).Interface())
			if err != nil {
				return nil, errors.New("Can not marshal remove item into json")
			}
			return nil, errors.New(fmt.Sprintf("Can not find remove item %s", string(b)))
		}
	}
	ret := reflect.MakeSlice(targetArray.Type(), 0, targetArray.Len()-len(removeIndexs))
	for i := 0; i < targetArray.Len(); i++ {
		if exist, _ := InArray(i, removeIndexs); exist {
			continue
		}
		ret = reflect.Append(ret, targetArray.Index(i))
	}
	return ret.Interface(), nil
}

func main() {

	users := []User{
		{
			name:     "golang1",
			password: "password1",
			id:       1,
			books: []string{
				"hello world",
				"hello golang",
			},
		},
		{
			name:     "golang2",
			password: "password2",
			id:       2,
			books: []string{
				"youwen",
				"hello golang",
			},
		},
		{
			name:     "golang1",
			password: "password1",
			id:       1,
			books: []string{
				"hello world",
				"hello golang",
			},
		},
	}
	removeUsers := []User{
		{
			name:     "golang2",
			password: "password2",
			id:       2,
			books: []string{
				"youwen",
				"hello golang",
			},
		},
	}
	//users = Dedumplicate(users).([]User)
	temp, err := Remove(users, removeUsers)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(temp.([]User))
	temp2 := temp.([]User)
	temp2 = Dedumplicate2(temp2).([]User)
	fmt.Println(temp2)
}
