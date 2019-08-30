package metaphone3

import (
	"encoding/csv"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestBasicWords(t *testing.T) {
	vals := []struct{ in, prim, sec string }{
		{"A", "A", ""},
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

func TestHarness(t *testing.T) {
	debug = false
	e := &Encoder{
		EncodeVowels: true,
	}
	out, _ := e.Encode("supernode")
	if want, got := "SAPARNAT", out; want != got {
		t.Fatalf("want: %v, got %v", want, got)
	}
}

func TestNameFiles(t *testing.T) {
	debug = false
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".test" {
			continue
		}

		csvFile, err := os.Open(filepath.Join("testdata", file.Name()))
		if err != nil {
			t.Fatal(err)
		}

		reader := csv.NewReader(csvFile)

		enc := &Encoder{}
		encV := &Encoder{EncodeVowels: true}
		encE := &Encoder{EncodeExact: true}
		encEV := &Encoder{EncodeVowels: true, EncodeExact: true}

		var cnt, encErr, encVErr, encEErr, encEVErr int

		for {
			// line format of the test files:
			// EncodeVowels - v == true, !v == false
			// EncodeExact - e == true, !v == false
			// originalWord,main !v!e,alt !v!e,main ve,alt ve,main !ve,alt !ve,main v!e,alt v!e
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Fatal(err)
			}
			in := line[0]

			cnt++
			encodeSafe(t, "Enc", enc, in, line[1], line[2], &encErr)
			encodeSafe(t, "EncEV", encEV, in, line[3], line[4], &encEVErr)
			encodeSafe(t, "EncE", encE, in, line[5], line[6], &encEErr)
			encodeSafe(t, "EncV", encV, in, line[7], line[8], &encVErr)
			//if t.Failed() {
			//	t.FailNow()
			//}
		}

		// now we're done with reading the file, output stats
		csvFile.Close()

		// output stats
		outputStat(t, "Enc", encErr, cnt)
		outputStat(t, "EncEV", encEVErr, cnt)
		outputStat(t, "EncE", encEErr, cnt)
		outputStat(t, "EncV", encVErr, cnt)

		if encErr+encEVErr+encEErr+encVErr > 0 {
			t.Errorf("Errors when processing %v", file.Name())
		}

	}
}

func outputStat(t *testing.T, name string, err, cnt int) {
	percent := float32(err) * 100.0 / float32(cnt)
	t.Logf("Encoder %v, error percent: %v%%", name, percent)
}

func encodeSafe(t *testing.T, name string, e *Encoder, in, main, alt string, errCt *int) {
	out1, out2 := e.Encode(in)

	if main != out1 {
		t.Errorf("Error Encoding '%v' with %v.  Out1 want '%v' got '%v'", in, name, main, out1)
	}
	if alt != out2 {
		t.Errorf("Error Encoding '%v' with %v.  Out2 want '%v' got '%v'", in, name, alt, out2)
	}

	if main != out1 || alt != out2 {
		*errCt++
	}

}
