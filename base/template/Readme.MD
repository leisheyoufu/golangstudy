## Inventory
MacBook-Pro:template chenglong$ go run inventory.go 
17 items are made of wool
{{- }} trim space at beginning while {{ -}} trim space at end
17items are made ofwool
3 5

## Html
go run html/html.go
curl http://127.0.0.1:9090/
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Hello</title>
</head>
<body>
    <p>Hello 沙河小王子</p>
</body>

## go template with variable
```
MacBook-Pro:template chenglong$ go run variable/variable.go 



          
        <span style="color:green">1</span>




        <span style="color:red">2</span>
```

## Reference
- [Using Functions Inside Go Templates](https://www.calhoun.io/intro-to-templates-p3-functions/)
- [go template with html](https://www.cnblogs.com/nickchen121/p/11517448.html)
- [go template with variable](https://blog.csdn.net/butterfly5211314/article/details/102978117)