package main

import "fmt"
import "github.com/dignifiedquire/xlang-util/golang_util"

func main() {
	c := golang_util.NewChannel(5)
	msg := golang_util.NewMessage([]byte("helloworld"))

	fmt.Println("trying to send message", string(msg.Bytes()))
	if c.TrySend(msg) != nil {
		fmt.Println("Failed to send message")
		return
	}

	if c.Len() != 1 {
		fmt.Println("Invalid length", c.Len())
		return
	}

	// FIXME: reusing the channel doesn't work yet
	// newMsg := c.TryRecv()
	// if newMsg == nil {
	// 	fmt.Println("Failed to recv messaeg")
	// 	return
	// }

	// fmt.Println("received", string(newMsg.Bytes()))
	// golang_util.DropMessage(newMsg)

	// msg2 := golang_util.NewMessage([]byte("hellorust"))

	// if c.TrySend(msg2) != nil {
	// 	fmt.Println("Failed to send message")
	// 	return
	// }

	// if c.Len() != 1 {
	// 	fmt.Println("Invalid length", c.Len())
	// 	return
	// }

	newMsg2 := c.TryRecvRust()
	if newMsg2 == nil {
		fmt.Println("Failed to recv message from rust")
		return
	}

	fmt.Println("received", newMsg2.Len(), string(newMsg2.Bytes()))

	golang_util.DropMessage(newMsg2)

}
