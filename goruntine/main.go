package main

import (
	"fmt"
	"github.com/leisheyoufu/golangstudy/goruntine/ctx"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func TestFileWrite(wireteString string) {
	err := ioutil.WriteFile("output.txt", []byte(wireteString), 0666)
	if err != nil {
		fmt.Println(err)
	}
}

func WriteFile2(content string) {
	fd, _ := os.OpenFile("output2.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fd_time := time.Now().Format("2006-01-02 15:04:05")
	fd_content := strings.Join([]string{"======", fd_time, "=====", content, " "}, "")
	buf := []byte(fd_content)
	fd.Write(buf)
	fd.Close()

}

func WriteConsoleInGoruntine() {
	runtime.GOMAXPROCS(1)
	for i := 0; i < 100; i++ {
		go func(i int) {
			//time.Sleep((time.Duration(1) * time.Second))
			//TestFileWrite(strconv.Itoa(i))
			WriteFile2(strconv.Itoa(i)) // if comment this line, the output in console will be sequential like 1, 2, 3, 4, 5
			fmt.Println(i)              // maybe the content is written to the console buffer then swith to another goruntine, so the sequence is not messy
		}(i)
	}
	time.Sleep((time.Duration(1) * time.Second))
	fmt.Println("Done")
}

func main() {
	//WriteConsoleInGoruntine()
	ctx.Cancel()
	ctx.Timeout()
	ctx.WithValue()
}
