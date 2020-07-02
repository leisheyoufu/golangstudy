redis-cli <ip:6379>
127.0.0.1:6379> set name loch
OK
127.0.0.1:6379> get name
"loch"
127.0.0.1:6379> keys *
1) "name"