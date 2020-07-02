package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	. "gopkg.in/check.v1"
)

var a int = 1

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

type HttpData struct {
	Flag int    `json:"flag"`
	Msg  string `json:"msg"`
}

var _ = Suite(&MySuite{})

var testurl string = "http://127.0.0.1:12345"

func (s *MySuite) SetUpSuite(c *C) {
	str3 := "第1次套件开始执行"
	fmt.Println(str3)
	//c.Skip("Skip TestSutie")
}

func (s *MySuite) TearDownSuite(c *C) {
	str4 := "第1次套件执行完成"
	fmt.Println(str4)
}

func (s *MySuite) SetUpTest(c *C) {
	str1 := "第" + strconv.Itoa(a) + "条用例开始执行"
	fmt.Println(str1)

}

func (s *MySuite) TearDownTest(c *C) {
	str2 := "第" + strconv.Itoa(a) + "条用例执行完成"
	fmt.Println(str2)
	a = a + 1
}

func (s *MySuite) TestHttpGet(c *C) {
	geturl := fmt.Sprintf("%v/checkon", testurl)
	respget, err := http.Get(geturl)
	if err != nil {
		panic(err)
	}
	defer respget.Body.Close() //关闭连接

	body, err := ioutil.ReadAll(respget.Body) //读取body的内容
	var gdat map[string]interface{}           //定义map用于解析resp.body的内容
	if err := json.Unmarshal([]byte(string(body)), &gdat); err == nil {
		fmt.Println(gdat)
	} else {
		fmt.Println(err)
	}
	var gmsg = gdat["msg"]
	c.Assert(gmsg, Equals, "terrychow") //模拟失败的断言

}

func (s *MySuite) TestHttpPost(c *C) {

	url := fmt.Sprintf("%v/postdata", testurl)
	contentType := "application/json;charset=utf-8"

	var httpdata HttpData
	httpdata.Flag = 1
	httpdata.Msg = "terrychow"

	b, err := json.Marshal(httpdata)
	if err != nil {
		fmt.Println("json format error:", err)
		return
	}

	body := bytes.NewBuffer(b)

	resp, err := http.Post(url, contentType, body)
	if err != nil {
		fmt.Println("Post failed:", err)
		return
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read failed:", err)
		return
	}
	var dat map[string]interface{} //定义map用于解析resp.body的内容
	if err := json.Unmarshal([]byte(string(content)), &dat); err == nil {
		fmt.Println(dat)
	} else {
		fmt.Println(err)
	}
	var msg = dat["msg"]
	c.Assert(msg, Equals, "terrychow") //模拟成功的断言
}
