// Package consistenthash provides an implementation of a weighted ring hash.
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash :
type Hash func(data []byte) uint32

// Map :
type Map struct {
	hash     Hash
	replicas int
	keys     []int // Sorted
	hashMap  map[int]string
	keyCount int
}

// New :
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// IsEmpty returns true if there are no items available.
func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}

// Add some keys and weight to the hash.
func (m *Map) Add(keyMap map[string]int) {
	m.keyCount = len(keyMap)
	for key, weight := range keyMap {
		m.addSingleKey(key, weight)
	}
	sort.Ints(m.keys)
}

func (m *Map) addSingleKey(key string, weight int) {
	for i := 0; i < m.replicas*weight; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		m.keys = append(m.keys, hash)
		m.hashMap[hash] = key
	}
}

// Get the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	// Means we have cycled back to the first replica.
	if idx == len(m.keys) {
		idx = 0
	}

	return m.hashMap[m.keys[idx]]
}

// GetItems :
func (m *Map) GetItems(key string) []string {
	if m.IsEmpty() {
		return nil
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	// Means we have cycled back to the first replica.
	if idx == len(m.keys) {
		idx = 0
	}

	hit := make(map[string]struct{})
	var items []string

	for i := 0; i < len(m.keys); i++ {
		currentIdx := (idx + i) % len(m.keys)
		item := m.hashMap[m.keys[currentIdx]]
		if _, exists := hit[item]; !exists {
			items = append(items, item)
			hit[item] = struct{}{}
			if len(items) == m.keyCount {
				return items
			}
		}
	}

	return items
}
