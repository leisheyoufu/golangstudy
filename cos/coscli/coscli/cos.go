package coscli

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mozillazg/go-cos"
	"github.com/spf13/cobra"
)

var (
	SecretID  string
	SecretKey string
	Prefix    string
	Bucket    string
)

func ListBucketCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "list-bucket",
		Long: `List cos bucket`,
		Run:  listBucket,
	}
	return cmd
}

func listBucket(cmd *cobra.Command, args []string) {
	c := cos.NewClient(nil, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  SecretID,
			SecretKey: SecretKey,
		},
	})
	s, _, err := c.Service.Get(context.Background())
	if err != nil {
		panic(err)
	}

	for _, b := range s.Buckets {
		fmt.Printf("%#v\n", b)
	}
}

func ListObjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "list-object",
		Long: `List cos object`,
		Run:  listObject,
	}
	cmd.Flags().StringVarP(&Prefix, "prefix", "p", "", "prefix of object")
	return cmd
}

func listObject(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Please specify kafka bucket")
		os.Exit(1)
	}
	u, _ := url.Parse(args[0])
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  SecretID,
			SecretKey: SecretKey,
		},
	})

	opt := &cos.BucketGetOptions{
		MaxKeys: 1000,
	}
	if Prefix != "" {
		opt.Prefix = Prefix
	}
	v, _, err := c.Bucket.Get(context.Background(), opt)
	if err != nil {
		panic(err)
	}

	for _, c := range v.Contents {
		fmt.Printf("%s, %d\n", c.Key, c.Size)
	}
}

func GetObjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "get-object",
		Long: `Get cos object`,
		Run:  getObject,
	}
	cmd.Flags().StringVarP(&Bucket, "bucket", "b", "", "bucket")
	return cmd
}

func getObject(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Please specify cos object")
		os.Exit(1)
	}
	if Bucket == "" {
		fmt.Fprintln(os.Stderr, "Please specify cos bucket")
		os.Exit(1)
	}
	u, _ := url.Parse(Bucket)
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  SecretID,
			SecretKey: SecretKey,
		},
	})

	// 1.通过响应体获取对象
	name := args[0]
	resp, err := c.Object.Get(context.Background(), name, nil)
	if err != nil {
		panic(err)
	}
	bs, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	output := strings.Replace(name, "/", "_", -1)
	output = "./" + output
	ioutil.WriteFile(output, bs, 0666) //写入文件(字节数组)
	fmt.Printf("%s\n", string(bs))
	// 2.获取对象到本地文件

}
