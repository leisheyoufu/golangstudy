ETCDCTL_API=3 etcdctl get "" --prefix --keys-only | sed '/^\s*$/d'
ETCDCTL_API=3 etcdctl --endpoints=127.0.0.1:12379 get "" --prefix --keys-only | sed '/^\s*$/d'