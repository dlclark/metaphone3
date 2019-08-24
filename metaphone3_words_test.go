package metaphone3

import "testing"

func TestBasicWords(t *testing.T) {
	vals := []struct{ in, prim, sec string }{
		{"A", "A", ""},
		//{"jose", "JS", "HS"},
		{"ack", "AK", ""},
		{"eek", "AK", ""},
		{"ache", "AK", "AX"},
	}
	e := &Encoder{}

	for _, v := range vals {
		prim, sec := e.Encode(v.in)
		if prim != v.prim {
			t.Errorf("Invalid primary output on '%v', wanted %v, got %v", v.in, v.prim, prim)
		}
		if sec != v.sec {
			t.Errorf("Invalid secondary output on '%v', wanted %v, got %v", v.in, v.sec, sec)
		}
	}
}
