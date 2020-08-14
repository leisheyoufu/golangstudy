package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

func PrintJson(b []byte) {
	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")
	_, err := out.WriteTo(os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to print json data, Error:%s\n", err.Error())
	}
	fmt.Printf("\n")
}

func InterfaceToSlice(in interface{}) []interface{} {
	s := make([]interface{}, 0)
	if reflect.TypeOf(in).Kind() == reflect.Slice {
		value := reflect.ValueOf(in)
		for i := 0; i < value.Len(); i++ {
			s = append(s, value.Index(i).Interface())
		}
	}
	return s
}

func InterfaceToStringSlice(in interface{}) []string {
	s := make([]string, 0)
	if reflect.TypeOf(in).Kind() == reflect.Slice {
		value := reflect.ValueOf(in)
		for i := 0; i < value.Len(); i++ {
			s = append(s, fmt.Sprint(value.Index(i)))
		}
	}
	return s
}
