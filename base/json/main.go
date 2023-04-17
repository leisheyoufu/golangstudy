// Package main
// Copyright 2017 Kranz. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
/**
	Test read file by chunks
	➜  read ll
	total 13241936
	drwxr-xr-x  5 rlopes  staff   170B 16 Jan 01:19 ./
	drwxr-xr-x  9 rlopes  staff   306B 12 Jan 23:20 ../
	-rw-r--r--  1 rlopes  staff   6.3G 16 Jan 00:53 data.json
	-rw-r--r--  1 rlopes  staff   2.4K 16 Jan 01:19 main.go
	➜  read time go run main.go
	Total of [398945] object created.
	The [data.json] is 6.3GB long
	To parse the file took [18.832351231s]
	go run main.go  18.50s user 0.60s system 100% cpu 19.087 total
**/
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

type Elm struct {
	ID         string   `json:"_id"`
	Index      int      `json:"index"`
	GuID       string   `json:"guid"`
	IsActive   bool     `json:"isActive"`
	Balance    string   `json:"balance"`
	Picture    string   `json:"picture"`
	Age        int      `json:"age"`
	EyeColor   string   `json:"eyeColor"`
	Name       string   `json:"name"`
	Gender     string   `json:"gender"`
	Company    string   `json:"company"`
	Email      string   `json:"email"`
	Phone      string   `json:"phone"`
	Address    string   `json:"address"`
	About      string   `json:"about"`
	Registered string   `json:"registered"`
	Latitude   float64  `json:"latitude"`
	Longitude  float64  `json:"longitude"`
	Tags       []string `json:"tags"`
	Friends    []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"friends"`
	Greeting      string `json:"greeting"`
	FavoriteFruit string `json:"favoriteFruit"`
}

func (e *Elm) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}

func main() {
	start := time.Now()

	fileName := "data.json"
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error to read [file=%v]: %v", fileName, err.Error())
	}

	fi, err := f.Stat()
	if err != nil {
		log.Fatalf("Could not obtain stat, handle error: %v", err.Error())
	}

	r := bufio.NewReader(f)
	d := json.NewDecoder(r)
	e := json.NewEncoder(os.Stdout)

	i := 0

	d.Token()
	for d.More() {
		elm := &Elm{}
		d.Decode(elm)
		//fmt.Printf("%v \n", elm)
		i++
	}
	d.Token()
	elapsed := time.Since(start)

	fmt.Printf("Total of [%v] object created.\n", i)
	fmt.Printf("The [%s] is %s long\n", fileName, FileSize(fi.Size()))
	fmt.Printf("To parse the file took [%v]\n", elapsed)
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%dB", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := float64(s) / math.Pow(base, math.Floor(e))
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}

	return fmt.Sprintf(f+"%s", val, suffix)
}

// FileSize calculates the file size and generate user-friendly string.
func FileSize(s int64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(uint64(s), 1024, sizes)
}
