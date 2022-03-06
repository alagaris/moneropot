package util

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"testing"
)

func TestXmrToUSD(t *testing.T) {
	price, err := XmrToUSD(true)
	if err != nil {
		t.Errorf("Wanted no error, got error %v", err)
	}
	usd := (float64(price) / 1e8)
	if usd < 150 {
		t.Errorf("Wanted > 150 got less = %f", usd)
	}
}

func TestHashMatch(t *testing.T) {
	s256 := sha256.New()
	sample := "447e2f6f2b88bc7cd7b76e027953131d3080cd021017a64142bed008805c622a"
	sign := "cca559627a767e3e99564700a6ed31aa0db3be879225d25155f6a363b26c8f04"
	m1 := make(map[int]int)
	m2 := make(map[int]int)
	for i := 1; i <= 200; i++ {
		s256.Write([]byte(sign + strconv.Itoa(i)))
		hash := fmt.Sprintf("%x", s256.Sum(nil))
		a := HashMatchAlign(sample, hash)
		b := HashMatchChar(sample, hash)
		// fmt.Println("Align", hash)
		// fmt.Println("-----", sample)
		// fmt.Println("-->", a)
		// fmt.Println("Chars", hash)
		// fmt.Println("-----", sample)
		// fmt.Println("-->", b)
		s256.Reset()
		aa, ok := m1[a]
		if !ok {
			m1[a] = 0
		}
		m1[a] = aa + 1
		bb, ok := m2[b]
		if !ok {
			m2[b] = 0
		}
		m2[b] = bb + 1
	}

	tables := []struct {
		x int
		y int
	}{
		{0, 5},
		{1, 12},
		{2, 32},
		{3, 36},
		{4, 29},
		{5, 34},
		{6, 22},
		{7, 22},
		{8, 4},
		{9, 1},
		{10, 2},
		{11, 1},
	}

	for _, table := range tables {
		if m1[table.x] != table.y {
			t.Errorf("Wanted %d in %d key got %d", table.y, table.x, m1[table.x])
		}
	}

	tables2 := []struct {
		x string
		y string
		z int
	}{
		{"abc", "def", 0},
		{"abc", "dbf", 1},
		{"abc", "dbc", 2},
		{"abc", "abc", 3},
		{"aaa", "aaa", 3},
		{"A12345", "abcDe5", 2},
	}

	for _, table := range tables2 {
		m := HashMatchAlign(table.x, table.y)
		if m != table.z {
			t.Errorf("Wanted %d got %d", table.z, m)
		}
	}
	// t.Errorf("m1: %v\n m2: %v", m1, m2)
}
