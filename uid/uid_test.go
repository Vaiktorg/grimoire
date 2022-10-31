package uid

import (
	"fmt"
	"testing"
)

func TestLen4(t *testing.T) {
	col := map[UID]string{}
	var iterations int
	for {
		uid := NewUID(4)
		if _, ok := col[uid]; ok {
			fmt.Println(iterations)
			t.FailNow()
			return
		}

		col[uid] = ""
		iterations++
	}
}

// T1: 13,803,887
// T2: 28,843,584
// T3: 13,099,357
func TestLen8(t *testing.T) {
	col := map[UID]string{}
	var iterations int
	for {
		uid := NewUID(8)
		if _, ok := col[uid]; ok {
			fmt.Println(iterations)
			t.FailNow()
			return
		}

		col[uid] = ""
		iterations++
	}
}

// T1: Memory too high
func TestLen16(t *testing.T) {
	col := map[UID]string{}
	var iterations int
	for {
		uid := NewUID(16)
		if _, ok := col[uid]; !ok {
			col[uid] = ""
			iterations++
			continue
		}

		fmt.Println(iterations)
		t.FailNow()
	}
}
