package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
)

func IfInsert(cli *clientv3.Client) error {
	var resp *clientv3.TxnResponse
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	printFunc()
	resp, err = cli.Txn(ctx).
		If(clientv3.Compare(clientv3.Value("/nodes/txn1"), "=", `{"driver":"ssh", "params": {"port":22, "user":"root"}`)).
		Then(clientv3.OpPut("/nodes/txn1", `{"driver":"ssh", "params": {"port":22, "user":"root"}`), clientv3.OpPut("/nodes/txn2", `{"hello":"hellow world"}`)).
		Else(clientv3.OpPut("/nodes/txn3", `{"driver":"ssh", "params": {"port":22, "user":"root"}`), clientv3.OpPut("/nodes/txn4", `{"hello":"hellow world"}`)).
		Commit()
	if err != nil {
		return err
	}
	// Note(cl): succeeded is only the result of If statement
	if !resp.Succeeded {
		fmt.Printf("IfInsert: if fail\n")
	}
	return nil
}

func notFoundInsert(cli *clientv3.Client) error {
	var resp *clientv3.TxnResponse
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	printFunc()
	resp, err = cli.Txn(ctx).If(notFound("nodes/txn5")).Then(clientv3.OpPut("/nodes/txn5", `{"driver":"ssh", "params": {"port":22, "user":"root"}`)).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		fmt.Printf("notFoundInsert: if fail\n")
	}
	return nil
}

func foundDelete(cli *clientv3.Client, paths ...string) error {
	var resp *clientv3.TxnResponse
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	printFunc()
	var cmpOps []clientv3.Cmp
	var delOps []clientv3.Op
	for _, path := range paths {
		cmpOps = append(cmpOps, found(path))
		delOps = append(delOps, clientv3.OpDelete(path))
	}
	resp, err = cli.Txn(ctx).If(cmpOps...).Then(delOps...).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		fmt.Printf("foundDelete: if fail\n")
	}
	return nil
}

func txnDelete(cli *clientv3.Client) error {
	var resp *clientv3.TxnResponse
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	printFunc()
	resp, err = cli.Txn(ctx).If(found("/nodes/txn1"), found("/nodes/txn2")).Then(clientv3.OpDelete("/nodes/txn1"), clientv3.OpDelete("/nodes/txn2")).
		Else(clientv3.OpDelete("/nodes/txn3"), clientv3.OpDelete("/nodes/txn4")).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		fmt.Printf("txnDelete: if fail\n")
	}
	return nil
}

func TestTxn() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints[:],
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cli.Close()
	err = put(cli, "/nodes/txn1", `{"driver":"ssh", "params": {"port":22, "user":"root"}`)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = IfInsert(cli)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = notFoundInsert(cli)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = foundDelete(cli, "/nodes/txn5")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = txnDelete(cli)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("TestTxn: ok\n")
}
