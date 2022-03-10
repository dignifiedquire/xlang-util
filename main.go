package main

/*
#include <stdint.h>
#include "./libxlang_util.h"
#cgo LDFLAGS: -L. -lxlang_util -lm -ldl
*/
import "C"

import "sync/atomic"
import "runtime"
import "unsafe"
import "github.com/dignifiedquire/xlang-util/golang_util"

type Channel struct {
    inner *C.struct_Channel
}

func NewChannel(cap uint32) Channel {
	channel := Channel {
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
	
	for {
		// Check if the channel is disconnected.
		if tail & (uint64)(r.inner.mark_bit) != 0 {
			token.slot = nil
			token.stamp = 0;
			return true
		}

		// Deconstruct the tail.
		index := tail & (uint64(r.inner.mark_bit) - 1)
		lap := tail & ^(uint64(r.inner.one_lap) - 1)

		// Inspect the corresponding slot.
		slotPtr := unsafe.Add(
			unsafe.Pointer(r.inner.buffer),
			uintptr(index) * unsafe.Sizeof(C.struct_Slot),
		)
		slot := (*C.struct_Slot)(slotPtr)
		stamp := atomicLoadUint64(&slot.stamp)

		// If the tail and the stamp match, we may attempt to push.
		if tail == stamp {
			var newTail uint64
			if index + 1 < uint64(r.inner.cap) {
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
				token.slot = slot
				token.stamp = tail + 1
				return true
			} else {
				tail = atomicLoadUint64(r.inner.tail)
				backoff.Spin()
			}
		} else if stamp + r.inner.one_lap == tail + 1 {
			head := atomicLoadUint64(r.inner.head)
			
			// If the head lags one lap behind the tail as well..
			if head + r.inner.one_lap == tail {
				// .. then the chanenl is full.
				return false
			}
			backoff.Spin()
			tail = atomicLoadUint64(r.inner.tail)
		} else {
			// Snooze because we need to wait for the stamp to get updated
			backoff.Snooze()
			tail = atomicLoadUint64(r.inner.tail)
		}
	}
}

// Returns `nil` on success, otherwise the original message
func (r *Channel) write(token *C.struct_Token, msg *C.struct_Message) *C.struct_message {
	// If there is no slot, the channel is disconnected.
	if token.slot == nil {
		return msg
	}

	slot := &token.slot
	slot.msg = msg
	atomic.StoreUint64(slot.stamp, token.stamp)

	return nil
}

func defaultToken() C.struct_Token {
	C.struct_Token  {
		slot: nil,
		stamp: 0,
	}
}

// return nil on success, message on error
func (r *Channel) TrySend(msg *C.struct_Message) *C.struct_Message {
	token := defaultToken()
	if r.start_send(&token) {
		return r.write(token, msg)
	} else {
		return msg
	}
}



func main() {
	s := NewChannel()
}
