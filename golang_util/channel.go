package golang_util

/*
#include "../libxlang_util.h"
#cgo LDFLAGS: -L. -lxlang_util -lm -ldl
*/
import "C"

import "reflect"
import "sync/atomic"
import "runtime"
import "unsafe"
// import "fmt"

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
	backoff := NewBackoff()
	tail := atomicLoadUint64(&r.inner.tail)

	for {
		// Check if the channel is disconnected.
		markBitSet := uint64(0)
		if r.inner.mark_bit != 0 {
			markBitSet = 1
		}
		if tail&markBitSet != 0 {
			token.slot = nil
			token.stamp = 0
			return true
		}

		// Deconstruct the tail.
		index := tail & (uint64(r.inner.mark_bit - 1))
		lap := tail & ^(uint64(r.inner.one_lap) - 1)

		// Inspect the corresponding slot.
		offset := uintptr(index) * slotSize
		slotPtr := uintptr(unsafe.Pointer(r.inner.buffer)) + offset
		slot := (*C.struct_Slot)(unsafe.Pointer(slotPtr))
		stamp := atomicLoadUint64(&slot.stamp)
		// fmt.Println("wrote to slot", slotPtr, index, slotSize, r.inner.buffer)

		// If the tail and the stamp match, we may attempt to push.
		if tail == stamp {
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
				// Prepare the token for the folow-up call to `write`.
				// fmt.Println("storing tail", slot, tail, newTail)
				token.slot = slot
				token.stamp = C.ulonglong(tail + 1)
				return true
			}
			tail = atomicLoadUint64(&r.inner.tail)
			backoff.Spin()
		} else if stamp+uint64(r.inner.one_lap) == tail+1 {
			head := atomicLoadUint64(&r.inner.head)

			// If the head lags one lap behind the tail as well..
			if head+uint64(r.inner.one_lap) == tail {
				// .. then the chanenl is full.
				return false
			}
			backoff.Spin()
			tail = atomicLoadUint64(&r.inner.tail)
		} else {
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
		return msg
	}

	// fmt.Println("writing to slot", token, msg, token.slot)
	token.slot.msg = *msg
	atomic.StoreUint64((*uint64)(&token.slot.stamp), uint64(token.stamp))
	// fmt.Println("stored", token.slot)
	return nil
}

type Token = C.struct_Token

func defaultToken() Token {
	return Token{
		slot:  nil,
		stamp: 0,
	}
}

// return nil on success, message on error
func (r *Channel) TrySend(msg *Message) *Message {
	token := defaultToken()
	// fmt.Println("sending", msg)
	if r.startSend(&token) {
		return r.write(&token, msg)
	}
	return msg
}

func (r *Channel) Send(msg *Message) *Message {
	token := defaultToken()
	for {
		// Try sending a message several times.
		backoff := NewBackoff()
		for {
			if r.startSend(&token) {
				return r.write(&token, msg)
			}
		}

		if backoff.IsCompleted() {
			break
		} else {
			backoff.Snooze()
		}

		backoff.Spin()
	}

	return nil
}

func (c *Channel) IsDisconnected() bool {
	markBitSet := uint64(0)
	if c.inner.mark_bit != 0 {
		markBitSet = 1
	}
	return atomicLoadUint64(&c.inner.mark_bit)&markBitSet != 0
}

func (c *Channel) read(token *Token) *Message {
	if token.slot == nil {
		return nil
	}

	slot := token.slot
	msg := &slot.msg
	atomic.StoreUint64((*uint64)(&slot.stamp), uint64(token.stamp))

	return msg
}

func (c *Channel) startRecv(token *Token) bool {
	backoff := NewBackoff()
	head := atomicLoadUint64(&c.inner.head)

	for {
		// Deconstruct the head.
		index := head & uint64(c.inner.mark_bit-1)
		lap := head & ^uint64(c.inner.one_lap-1)

		// Inspect the corresponding slot.
		offset := uintptr(index) * slotSize
		slotPtr := uintptr(unsafe.Pointer(c.inner.buffer)) + offset
		slot := (*C.struct_Slot)(unsafe.Pointer(slotPtr))
		stamp := atomicLoadUint64(&slot.stamp)

		// If the stamp is ahead of the head by 1, we may attempt to pop.
		if head+1 == stamp {
			var new uint64
			if index+1 < uint64(c.inner.cap) {
				// Same lap, incremented index.
				// Set to `{ lap: lap, mark: 0, index: index + 1 }`.
				new = head + 1
			} else {
				// One lap forward, index wraps around to zero.
				// Set to `{ lap: lap.wrapping_add(1), mark: 0, index: 0 }`.
				new = lap + uint64(c.inner.one_lap)
			}

			// Try moving the head.
			if atomic.CompareAndSwapUint64((*uint64)(&c.inner.head), head, new) {
				// Prepare the token fo the follow-up call to `read`
				token.slot = slot
				token.stamp = C.ulonglong(head + uint64(c.inner.one_lap))
				return true
			}
			head = atomicLoadUint64(&c.inner.head)
			backoff.Spin()
		} else if stamp == head {
			tail := atomicLoadUint64(&c.inner.tail)

			// If the tail equals the head, that means the channel is empty.
			if tail & ^uint64(c.inner.mark_bit) == head {
				// If the channel is disconnected..
				if tail&uint64(c.inner.mark_bit) != 0 {
					// ..then receive an error.
					token.slot = nil
					token.stamp = 0
					return true
				}
				// Otherwise the receive operation is not ready.
				return false
			}

			backoff.Spin()
			head = atomicLoadUint64(&c.inner.head)
		} else {
			// Snooze because we need to wait for the stamp to get updated.
			backoff.Snooze()
			head = atomicLoadUint64(&c.inner.head)
		}
	}
}

func (c *Channel) TryRecv() *Message {
	token := defaultToken()

	if c.startRecv(&token) {
		return c.read(&token)
	}

	return nil
}

func (c *Channel) TryRecvRust() *Message {
	msg := C.channel_try_recv(c.inner)
	return msg
}

func DropMessage(msg *Message) {
	C.drop_message(msg)
}

func (c *Channel) Recv() *Message {
	token := defaultToken()

	for {
		// Try receiving a message several times.
		backoff := NewBackoff()
		for {
			if c.startRecv(&token) {
				return c.read(&token)
			}

			if backoff.IsCompleted() {
				break
			} else {
				backoff.Snooze()
			}
		}

		backoff.Spin()
	}

	return nil
}

// Len returns the current number of messages inside the channel.
func (c *Channel) Len() uint64 {
	for {
		// Load the tail, then load the head
		tail := atomicLoadUint64(&c.inner.tail)
		head := atomicLoadUint64(&c.inner.head)

		// If the tail didn't change, we've got consistent values to work with.
		if atomicLoadUint64(&c.inner.tail) == tail {
			hix := head & uint64(c.inner.mark_bit-1)
			tix := tail & uint64(c.inner.mark_bit-1)

			if hix < tix {
				return tix - hix
			}
			if hix > tix {
				return uint64(c.inner.cap) - hix + tix
			}
			if tail & ^uint64(c.inner.mark_bit) == head {
				return 0
			}
			return uint64(c.inner.cap)
		}
	}
}

type Message = C.struct_Message

func NewMessage(bytes []byte) Message {
	l := C.ulonglong(len(bytes))
	ptr := C.new_message_bytes((*C.uchar)(unsafe.Pointer(&bytes[0])), l)

	// runtime.SetFinalizer(&msg, func(msg *Message) {
	// 	C.drop_message_bytes(msg.ptr, msg.len)
	// })
	
	return Message {
		ptr: ptr,
		len: l,
	}
}

func (msg *Message) Len() uint64 {
	return uint64(msg.len)
}

func (msg *Message) Bytes() []byte {
	if msg.ptr == nil {
		return nil
	}

	slice := (*[1 << 30]byte)(unsafe.Pointer(msg.ptr))[:msg.len]
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = int(msg.len)
	return slice
}

