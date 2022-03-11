package golang_util_test

import "testing"
import "github.com/dignifiedquire/xlang-util/golang_util"


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
	for i := 0; i < b.N; i++ {
		msgs[i] = golang_util.NewMessage(data)
	}

	for i := range msgs {
		msgs[i].Drop()
	}
}

func BenchmarkNewMessageGo(b *testing.B) {
	msgs := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data := make([]byte, 100)
		for i := range data {
			data[i] = 1
		}
		msgs[i] = data
	}
}
