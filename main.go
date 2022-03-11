package main

import "fmt"
import "github.com/dignifiedquire/xlang-util/golang_util"

func main() {
	c := golang_util.NewChannel(10)
	msg := golang_util.NewMessage([]byte("helloworld"))

	fmt.Println("trying to send message", string(msg.Bytes()))
	if c.TrySend(&msg) != nil {
		fmt.Println("Failed to send message")
		return
	}

	if c.Len() != 1 {
		fmt.Println("Invalid length", c.Len())
		return
	}

	newMsg := c.TryRecv()
	if newMsg == nil {
		fmt.Println("Failed to recv messaeg")
		return
	}

	fmt.Println("received", string(newMsg.Bytes()))
}
