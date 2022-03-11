package main

/*
#include "./libxlang_util.h"
#cgo LDFLAGS: -L. -lxlang_util -lm -ldl
uintptr_t slotSize() {
  return sizeof(Slot);
}
*/
import "C"

import "fmt"
import "sync/atomic"
import "runtime"
import "unsafe"
import "github.com/dignifiedquire/xlang-util/golang_util"

var slotSize uintptr

func init() {
	slotSize = uintptr(C.slotSize())
}

type Channel struct {
	inner *C.struct_Channel
}

func NewChannel(cap uint32) Channel {
	channel := Channel{
		inner: C.new_channel((C.uint)(cap)),
	}
	runtime.SetFinalizer(&channel, func(r *Channel) {
		r.Drop()
	})
	return channel
}

func (a *Channel) Drop() {
	if a.inner != nil {
		C.drop_channel(a.inner)
		a.inner = nil
	}
}

func atomicLoadUint64(val *C.ulonglong) uint64 {
	return atomic.LoadUint64((*uint64)(val))
}

// startSend attempts to reserve a slot for sending a message.
func (r *Channel) startSend(token *C.struct_Token) bool {
	backoff := golang_util.NewBackoff()
	tail := atomicLoadUint64(&r.inner.tail)
	fmt.Println("tail", tail, r.inner.tail)

	fmt.Println("startSend")
	for {
		// Check if the channel is disconnected.
		markBitSet := uint64(0)
		if r.inner.mark_bit != 0 {
			markBitSet = 1
		}
		if tail&markBitSet != 0 {
			fmt.Println("disconnected")
			token.slot = nil
			token.stamp = 0
			return true
		}

		// Deconstruct the tail.
		index := tail & (uint64(r.inner.mark_bit - 1))
		lap := tail & ^(uint64(r.inner.one_lap) - 1)

		// Inspect the corresponding slot.
		offset := uintptr(index) * slotSize
		fmt.Println("offset", offset, slotSize, index, tail, r.inner.mark_bit)
		slotPtr := uintptr(unsafe.Pointer(r.inner.buffer)) + offset
		slot := (*C.struct_Slot)(unsafe.Pointer(slotPtr))
		stamp := atomicLoadUint64(&slot.stamp)

		// If the tail and the stamp match, we may attempt to push.
		if tail == stamp {
			fmt.Println("tail == stamp")
			var newTail uint64
			if index+1 < uint64(r.inner.cap) {
				// Same lap, incremented index.
				// Set to `{ lap: lap, mark: 0, index: index + 1 }`.
				newTail = tail + 1
			} else {
				// One lap forward, index wraps around to zero.
				// Set to `{ lap: lap.wrapping_add(1), mark: 0, index: 0 }`.

				newTail = lap + uint64(r.inner.one_lap)
			}

			// Try moving the tail.
			if atomic.CompareAndSwapUint64((*uint64)(&r.inner.tail), tail, newTail) {
				fmt.Println("swapped", slot)
				// Prepare the token for the folow-up call to `write`.
				token.slot = slot
				token.stamp = C.ulonglong(tail + 1)
				return true
			}
			tail = atomicLoadUint64(&r.inner.tail)
			backoff.Spin()
		} else if stamp+uint64(r.inner.one_lap) == tail+1 {
			fmt.Println("lag")
			head := atomicLoadUint64(&r.inner.head)

			// If the head lags one lap behind the tail as well..
			if head+uint64(r.inner.one_lap) == tail {
				// .. then the chanenl is full.
				return false
			}
			backoff.Spin()
			tail = atomicLoadUint64(&r.inner.tail)
		} else {
			// fmt.Println("snooze")
			// Snooze because we need to wait for the stamp to get updated
			backoff.Snooze()
			tail = atomicLoadUint64(&r.inner.tail)
		}
	}
}

// Returns `nil` on success, otherwise the original message
func (r *Channel) write(token *C.struct_Token, msg *Message) *Message {
	// If there is no slot, the channel is disconnected.
	if token.slot == nil {
		fmt.Println("missing slot")
		return msg
	}

	slot := token.slot
	slot.msg = msg
	atomic.StoreUint64((*uint64)(&slot.stamp), uint64(token.stamp))

	return nil
}

func defaultToken() C.struct_Token {
	return C.struct_Token{
		slot:  nil,
		stamp: 0,
	}
}

// return nil on success, message on error
func (r *Channel) TrySend(msg *Message) *Message {
	token := defaultToken()
	if r.startSend(&token) {
		fmt.Println("write")
		return r.write(&token, msg)
	}
	return msg
}

type Message = C.struct_Message

func NewMessage(bytes []byte) Message {
	l := C.ulonglong(len(bytes))
	ptr := C.new_message_bytes((*C.uchar)(unsafe.Pointer(&bytes[0])), l)
	msg := C.struct_Message{
		ptr: ptr,
		len: l,
	}

	runtime.SetFinalizer(&msg, func(msg *Message) {
		C.drop_message(msg)
		msg.ptr = nil
	})
	return msg
}

func (msg *Message) Len() uint64 {
	return uint64(msg.len)
}

func main() {
	c := NewChannel(10)
	msg := NewMessage([]byte("helloworld"))

	fmt.Println("trying to send message")

	if c.TrySend(&msg) == nil {
		fmt.Println("Sent message")
	} else {
		fmt.Println("Failed to send message")
	}
}
