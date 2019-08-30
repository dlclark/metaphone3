package main

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/dlclark/regexp2"
)

func main() {

	re := regexp2.MustCompile(`,\s*([0-9]+)\s*,\s*((\"(?<val>[^\"]+)\")\s*[,\)]\s*)+`, 0)

	b, _ := ioutil.ReadFile("Metaphone3.java")

	for m, _ := re.FindStringMatch(string(b)); m != nil; m, _ = re.FindNextMatch(m) {
		strNum := m.GroupByNumber(1).String()
		num, _ := strconv.Atoi(strNum)

		g := m.GroupByName("val")

		for _, c := range g.Captures {
			if num != c.Length {
				fmt.Printf("Num=%v, Cap=%v\n", num, c.String())
			}
		}

	}

}
