package metaphone3

import "testing"

func TestStringAt_Basic(t *testing.T) {
	if want, got := true, testStringAt("TESTING", 0, 1, "B", "E"); want != got {
		t.Fatalf("StringAt error, wanted %v got %v", want, got)
	}
}

func TestStringAt_Calver(t *testing.T) {
	if want, got := true, testStringAt("CALVER", 2, -2, "POLKA", "PALKO", "HALVA", "HALVO", "SALVER", "CALVER"); want != got {
		t.Fatalf("StringAt error, wanted %v got %v", want, got)
	}
}

func TestStringStart_Basic(t *testing.T) {
	if want, got := true, testStringStart("TESTING", 1, "B", "E", "TEST"); want != got {
		t.Fatalf("StringStart error, wanted %v got %v", want, got)
	}
}

func TestStringAtEnd_Basic(t *testing.T) {
	if want, got := true, testStringAtEnd("TESTING", 1, 2, "B", "E", "TING"); want != got {
		t.Fatalf("StringStart error, wanted %v got %v", want, got)
	}
	if want, got := true, testStringAtEnd("JOSE", 0, 1, "OSE"); want != got {
		t.Fatalf("StringStart error, wanted %v got %v", want, got)
	}
}

func TestRootOrInflections_Basic(t *testing.T) {
	if want, got := true, rootOrInflections([]rune("CHRISTENING"), "CHRISTEN"); want != got {
		t.Fatalf("rootOrInflections error, Wanted %v, got %v", want, got)
	}
}

func TestRootOrInflections_Ache(t *testing.T) {
	if want, got := true, rootOrInflections([]rune("ACHY"), "ACHE"); want != got {
		t.Fatalf("rootOrInflections error, Wanted %v, got %v", want, got)
	}
}

func testStringAt(in string, curIdx, offset int, vals ...string) bool {
	e := &Encoder{}
	e.in = []rune(in)
	e.idx = curIdx
	return e.stringAt(offset, vals...)
}

func testStringStart(in string, curIdx int, vals ...string) bool {
	e := &Encoder{}
	e.in = []rune(in)
	e.idx = curIdx
	return e.stringStart(vals...)
}

func testStringAtEnd(in string, curIdx, offset int, vals ...string) bool {
	e := &Encoder{}
	e.in = []rune(in)
	e.idx = curIdx
	return e.stringAtEnd(offset, vals...)
}
