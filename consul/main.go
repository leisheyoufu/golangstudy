package main

import (
	"fmt"
	"log"
	"os"

	consulapi "github.com/hashicorp/consul/api"
)

func registerService(client *consulapi.Client) {
	names := []string{"job3", "job4", "job5"}
	for _, name := range names {
		checkPort := 8080
		registration := new(consulapi.AgentServiceRegistration)
		registration.ID = name
		registration.Name = name
		registration.Port = 9104
		registration.Tags = []string{"inst_service:mysql-svc.loch.svc.cluster.local:3306", "loch-test"}
		registration.Address = "192.168.0.1"
		registration.Check = &consulapi.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d%s", registration.Address, checkPort, "/check"),
			Timeout:  "3s",
			Interval: "5s",
			//DeregisterCriticalServiceAfter: "30s", //check失败后30秒删除本服务
		}

		err := client.Agent().ServiceRegister(registration)

		if err != nil {
			log.Fatal("register server error : ", err)
		}
	}
}

func listUnExistedConsulServices(client *consulapi.Client) error {
	consulServices, err := client.Agent().Services()
	if err != nil {
		return err
	}
	fmt.Printf("Curret consul services:")
	for consulService, _ := range consulServices {
		fmt.Printf("%s ", consulService)
	}
	fmt.Printf("\n")
	return nil
}

func deleteConsulServices(client *consulapi.Client, service string) error {
	err := client.Agent().ServiceDeregister(service)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	//kubeconfig := flag.String("kubeconfig", "config", "(optional) absolute path to the kubeconfig file")
	//flag.Parse()
	//
	//config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//if err != nil {
	//	panic(err.Error())
	//}
	// creates the in-cluster config

	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = "127.0.0.1:8500"
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %+v\n", err)
		os.Exit(1)
	}
	registerService(consulClient)
	err = listUnExistedConsulServices(consulClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %+v\n", err)
		os.Exit(1)
	}
	err = deleteConsulServices(consulClient, "job3")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %+v\n", err)
		os.Exit(1)
	}
	err = listUnExistedConsulServices(consulClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %+v\n", err)
		os.Exit(1)
	}
}
