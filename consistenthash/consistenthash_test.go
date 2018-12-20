package consistenthash

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {

	// Override the hash function to return easier to reason about values. Assumes
	// the keys can be converted to an integer.
	hash := New(3, func(key []byte) uint32 {
		i, err := strconv.Atoi(string(key))
		if err != nil {
			panic(err)
		}
		return uint32(i)
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add(map[string]int{
		"6": 1,
		"4": 1,
		"2": 1,
	})

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// Adds 8, 18, 28
	hash.Add(map[string]int{"8": 1})

	// 27 should now map to 8.
	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
}

func TestConsistency(t *testing.T) {
	hash1 := New(1, nil)
	hash2 := New(1, nil)

	hash1.Add(map[string]int{
		"Bill":  1,
		"Bob":   1,
		"Bonny": 1,
	})
	hash2.Add(map[string]int{
		"Bob":   1,
		"Bonny": 1,
		"Bill":  1,
	})

	if hash1.Get("Ben") != hash2.Get("Ben") {
		t.Error("Fetching 'Ben' from both hashes should be the same")
	}

	hash2.Add(map[string]int{
		"Becky": 1,
		"Ben":   1,
		"Bobby": 1,
	})

	if hash1.Get("Ben") != hash2.Get("Ben") ||
		hash1.Get("Bob") != hash2.Get("Bob") ||
		hash1.Get("Bonny") != hash2.Get("Bonny") {
		t.Error("Direct matches should always return the same entry")
	}

}

func TestWeight(t *testing.T) {
	hash := New(3, nil)

	hash.Add(map[string]int{
		"one": 1,
		"two": 2,
	})

	expectOne := 3
	expectTwo := 6
	actualOne := 0
	actualTwo := 0
	for _, v := range hash.hashMap {
		if v == "one" {
			actualOne++
		} else if v == "two" {
			actualTwo++
		}
	}

	if expectOne != actualOne || expectTwo != actualTwo {
		t.Errorf("expectOne : %d, actualOne : %d, expectTwo : %d, actualTwo : %d", expectOne, actualOne, expectTwo, actualTwo)
	}

	hitMap := make(map[string]int)
	for i := 0; i < 100000; i++ {
		key := strconv.Itoa(i)
		node := hash.Get(key)
		hitMap[node]++
	}

	fmt.Println()
	for k, v := range hitMap {
		fmt.Printf("key : %s, hit : %d\n", k, v)
	}
}

func TestGetItems(t *testing.T) {
	// Override the hash function to return easier to reason about values. Assumes
	// the keys can be converted to an integer.
	hash := New(3, func(key []byte) uint32 {
		i, err := strconv.Atoi(string(key))
		if err != nil {
			panic(err)
		}
		return uint32(i)
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add(map[string]int{
		"6": 1,
		"4": 1,
		"2": 1,
	})

	testCases := map[string][]string{
		"2":  []string{"2", "4", "6"},
		"11": []string{"2", "4", "6"},
		"23": []string{"4", "6", "2"},
		"25": []string{"6", "2", "4"},
		"27": []string{"2", "4", "6"},
	}

	for k, v := range testCases {
		if !reflect.DeepEqual(hash.GetItems(k), v) {
			t.Errorf("Asking for %v, should have yielded %v", k, v)
		}
	}
}
