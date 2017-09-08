package main

import (
	"fmt"
	"time"

	"github.com/leisheyoufu/golangstudy/task/actor"
	"github.com/leisheyoufu/golangstudy/task/common"
)

func generalNoParamFunc() {
	fmt.Printf("This is general no param func\n")
}

func generalOneParamFunc(param1 interface{}) {
	fmt.Printf("This is general one param func: %s\n", param1.(string))
}

func generalTwoParamFunc(params ...interface{}) {
	s := params[0].([]interface{})[0].(string)
	id := params[0].([]interface{})[1].(int)
	fmt.Printf("This is general multiple param func: %s,%d\n", s, id)
}

func main() {
	taskManager := common.NewTaskManager(10, 16)
	stringActorTask, err := taskManager.RegisterActorWorker(&actor.StringActor{})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	intActorTask, err := taskManager.RegisterActorWorker(&actor.IntActor{})
	for i := 0; i < 3; i++ {
		taskManager.Send(stringActorTask.GetID(), fmt.Sprintf("hello world %d", i))
	}
	for i := 0; i < 3; i++ {
		taskManager.Send(intActorTask.GetID(), i)
	}
	taskManager.Register(generalNoParamFunc)
	taskManager.Register(generalOneParamFunc, "hello")
	taskManager.Register(generalTwoParamFunc, "hello", intActorTask.GetID())
	time.Sleep(time.Millisecond * 100)
	taskManager.Stop(stringActorTask.GetID())
	taskManager.Stop(intActorTask.GetID())
	for taskManager.Running() {
		fmt.Println("Still running")
		time.Sleep(time.Millisecond * 100)
	}
}
