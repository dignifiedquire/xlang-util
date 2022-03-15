package golang_util_test

import "testing"
import "github.com/dignifiedquire/xlang-util/golang_util"


func makeMessage(i byte) golang_util.Message {
	data := make([]byte, 10)
	for k := range data {
		data[k] = byte(i)
	}
	return golang_util.NewMessage(data)
}

func TestDropFilledChannel(t *testing.T) {
	c := golang_util.NewChannel(100)
	for i := 0; i < 10 ; i++ {
		msg := makeMessage(byte(i))
		defer msg.Drop()

		if c.TrySend(&msg) != nil {
			t.Fatalf("failed to send message %v", i)
		}
	}
	if c.Len() != 10 {
		t.Fatalf("invalid channel len: %v", c.Len())
	}
	// finalizer must not fail
}

func TestTrySendRecv(t *testing.T) {
	c := golang_util.NewChannel(10)
	for i := 0; i < 10 ; i++ {
		msg := makeMessage(byte(i))
		defer msg.Drop()

		if c.TrySend(&msg) != nil {
			t.Fatalf("failed to send message %v", i)
		}
	}

	// channel is full
	msg := makeMessage(10)
	defer msg.Drop()
	if c.TrySend(&msg) == nil {
		t.Fatalf("channel sent, even though it was full")
	}

	if !c.IsFull() {
		t.Fatalf("channel is full")
	}

	for i := 0; i < 10 ; i++ {
		msg := c.TryRecv()
		if msg == nil {
			t.Fatalf("failed to recv message %v", i)
		}
		if msg.Len() != 10 {
			t.Fatalf("wrong message length received")
		}
		bytes := msg.Bytes()
		for k := range bytes {
			if bytes[k] != byte(i) {
				t.Fatalf("wrong message bytes received")
			}
		}
	}

	if c.TryRecv() != nil {
		t.Fatalf("received message, even though channel was empty")
	}
}

func TestSendRecv(t *testing.T) {
	c := golang_util.NewChannel(10)
	for i := 0; i < 10 ; i++ {
		msg := makeMessage(byte(i))
		defer msg.Drop()

		if c.Send(&msg) != nil {
			t.Fatalf("failed to send message %v", i)
		}
	}

	// channel is full
	if !c.IsFull() {
		t.Fatalf("channel is full")
	}

	for i := 0; i < 10 ; i++ {
		msg := c.Recv()
		if msg == nil {
			t.Fatalf("failed to recv message %v", i)
		}
		if msg.Len() != 10 {
			t.Fatalf("wrong message length received")
		}
		bytes := msg.Bytes()
		for k := range bytes {
			if bytes[k] != byte(i) {
				t.Fatalf("wrong message bytes received")
			}
		}
	}
}

func TestSendGoRecvRust(t *testing.T) {
	c := golang_util.NewChannel(10)
	for i := 0; i < 10 ; i++ {
		msg := makeMessage(byte(i))
		defer msg.Drop()

		if c.Send(&msg) != nil {
			t.Fatalf("failed to send message %v", i)
		}
	}

	// channel is full
	if !c.IsFull() {
		t.Fatalf("channel is full")
	}

	for i := 0; i < 10 ; i++ {
		msg := c.RecvRust()
		if msg == nil {
			t.Fatalf("failed to recv message %v", i)
		}
		if msg.Len() != 10 {
			t.Fatalf("wrong message length received")
		}
		bytes := msg.Bytes()
		for k := range bytes {
			if bytes[k] != byte(i) {
				t.Fatalf("wrong message bytes received")
			}
		}
	}
}

func TestSendRecvRust(t *testing.T) {
	c := golang_util.NewChannel(10)
	for i := 0; i < 10 ; i++ {
		msg := makeMessage(byte(i))
		// defer msg.Drop()

		if c.SendRust(&msg) != nil {
			t.Fatalf("failed to send message %v", i)
		}
	}

	// channel is full
	if !c.IsFull() {
		t.Fatalf("channel is full")
	}

	for i := 0; i < 10 ; i++ {
		msg := c.RecvRust()
		if msg == nil {
			t.Fatalf("failed to recv message %v", i)
		}
		if msg.Len() != 10 {
			t.Fatalf("wrong message length received")
		}
		bytes := msg.Bytes()
		for k := range bytes {
			if bytes[k] != byte(i) {
				t.Fatalf("wrong message bytes received")
			}
		}
	}
}

func TestSendRecvPar(t *testing.T) {
	c := golang_util.NewChannel(10)
	done := make(chan bool)

	go func() {
		for i := 0; i < 100 ; i++ {
			msg := makeMessage(byte(i))

			if c.Send(&msg) != nil {
				t.Fatalf("failed to send message %v", i)
			}
		}
	}()

	go func() {
		for i := 0; i < 100 ; i++ {
			msg := c.Recv()
			if msg == nil {
				t.Fatalf("failed to recv message %v", i)
			}
			if msg.Len() != 10 {
				t.Fatalf("wrong message length received")
			}
			bytes := msg.Bytes()
			for k := range bytes {
				if bytes[k] != byte(i) {
					t.Fatalf("wrong message bytes received: %v (%v)", bytes, byte(i))
				}
			}
			msg.Drop()
		}
		done <- true
	}()

	<- done
}


func TestSendRecvRustPar(t *testing.T) {
	c := golang_util.NewChannel(10)
	done := make(chan bool)

	go func() {
		for i := 0; i < 100 ; i++ {
			msg := makeMessage(byte(i))

			if c.SendRust(&msg) != nil {
				t.Fatalf("failed to send message %v", i)
			}
		}
	}()

	go func() {
		for i := 0; i < 100 ; i++ {
			msg := c.RecvRust()
			if msg == nil {
				t.Fatalf("failed to recv message %v", i)
			}
			if msg.Len() != 10 {
				t.Fatalf("wrong message length received")
			}
			bytes := msg.Bytes()
			for k := range bytes {
				if bytes[k] != byte(i) {
					t.Fatalf("wrong message bytes received: %v (%v)", bytes, byte(i))
				}
			}
			msg.Drop()
		}
		done <- true
	}()

	<- done
}


func TestMin(t *testing.T) {
	if golang_util.Min(5, 3) != 3 {
		t.Fatalf("invalid min")
	}
}

func TestNewMessage(t *testing.T) {
	data := "helloworld"
	msg := golang_util.NewMessage(([]byte)(data))
	if data != string(msg.Bytes()) {
		t.Fatalf(`message contains wrong values %s != %s`, data, string(msg.Bytes()))
	}
}


func BenchmarkNewMessageCgo(b *testing.B) {
        data := make([]byte, 100)
	for i := range data {
		data[i] = 1
	}

	msgs := make([]golang_util.Message, b.N)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		msgs[i] = golang_util.NewMessage(data)
	}

	for i := range msgs {
		msgs[i].Drop()
	}
}

func BenchmarkNewMessageGo(b *testing.B) {
	msgs := make([][]byte, b.N)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data := make([]byte, 100)
		for i := range data {
			data[i] = 1
		}
		msgs[i] = data
	}
}


func BenchmarkTrySend(b *testing.B) {
        data := make([]byte, 100)
	for i := range data {
		data[i] = 1
	}

	msgs := make([]golang_util.Message, b.N)
	c := golang_util.NewChannel(uint32(b.N))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		msgs[i] = golang_util.NewMessage(data)
		c.TrySend(&msgs[i])
	}

	for i := range msgs {
		msgs[i].Drop()
	}
}

func BenchmarkByteSliceGoChanCap1(b *testing.B) {
	ch := make(chan []byte, 1)
	go func() {
		for {
			<-ch
		}
	}()
	data := make([]byte, 10)

	b.ResetTimer()
	b.ReportAllocs()
	for k := range data {
		data[k] = byte(1)
	}

	for i := 0; i < b.N; i++ {
		ch <- data
	}
}

func BenchmarkByteSliceGoChanCap10(b *testing.B) {
	ch := make(chan []byte, 10)
	go func() {
		for {
			<-ch
		}
	}()
	data := make([]byte, 10)

	b.ResetTimer()
	b.ReportAllocs()
	for k := range data {
		data[k] = byte(1)
	}

	for i := 0; i < b.N; i++ {
		ch <- data
	}
}


func BenchmarkByteSliceChannelCap1(b *testing.B) {
	ch := golang_util.NewChannel(1)
	go func() {
		for {
			_ = ch.Recv()
		}
	}()

	msg := makeMessage(10)
	defer msg.Drop()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ch.Send(&msg)
	}
}


func BenchmarkByteSliceChannelCap10(b *testing.B) {
	ch := golang_util.NewChannel(10)
	go func() {
		for {
			_ = ch.Recv()
		}
	}()

	msg := makeMessage(10)
	defer msg.Drop()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ch.Send(&msg)
	}
}

func BenchmarkCgoCgoSendRecvCap10(b *testing.B) {
	// Send from Rust -> Rust

	ch := golang_util.NewChannel(10)

	go func() {
		for {
			_ = ch.RecvRust()
		}
	}()

	msg := makeMessage(10)
	defer msg.Drop()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ch.SendRust(&msg)
	}
}

func BenchmarkGoCgoSendRecvCap10(b *testing.B) {
	// Send from Go -> Rust
	ch := golang_util.NewChannel(10)

	go func() {
		for {
			_ = ch.RecvRust()
		}
	}()

	msg := makeMessage(10)
	// defer msg.Drop()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ch.Send(&msg)
	}
}

func BenchmarkCgoGoSendRecvCap10(b *testing.B) {
	// Send from Rust -> Go
	
	ch := golang_util.NewChannel(10)

	go func() {
		for {
			_ = ch.Recv()
		}
	}()

	msg := makeMessage(10)
	defer msg.Drop()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ch.SendRust(&msg)
	}
}
