package smokego

import "testing"

func TestHello(t *testing.T) {
	if Hello() != "hello" {
		t.Fatal("hello")
	}
}
