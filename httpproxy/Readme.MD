Reference： https://www.cnblogs.com/lanyangsh/p/10293051.html

```
go run httpproxy/trueserver.go   // true server is listening on port 2003
go run httpproxy/proxyserver.go  // reverse proxy server is listening on port 2002
curl http://127.0.0.1:2003/hello // will return like 2019-12-29 16:45:44 hello, world!
```