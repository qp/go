package shared

import "sync"

// StringDES is a thread-safe double ended stack of strings
// StringDES can directly manipulate an existing slice of strings
// as though it were a DES.
type StringDES []string

var lock sync.RWMutex

// Push pushes a new value on the top of the StringDES.
func (s *StringDES) Push(v string) {
	lock.Lock()
	*s = append(*s, "")      // add another bucket
	copy((*s)[1:], (*s)[0:]) // move all elements down by one
	(*s)[0] = v              // assign new value to front
	lock.Unlock()
}

// Pop removes a value from the top of the DES and returns it.
func (s *StringDES) Pop() string {
	if len(*s) == 0 {
		return ""
	}
	var v string
	lock.Lock()
	// retrieves the first item for returning
	// reslices the slice to place the second item at the beginning
	// also shrinks the slice's capacity
	v, *s = (*s)[0], (*s)[1:len(*s):len(*s)]
	lock.Unlock()
	return v
}

// Peek returns the top value from the DES without removing it
func (s *StringDES) Peek() string {
	if len(*s) == 0 {
		return ""
	}
	var v string
	lock.RLock()
	v = (*s)[0]
	lock.RUnlock()
	return v
}

// BPush pushes a new value on the top of the StringDES.
func (s *StringDES) BPush(v string) {
	lock.Lock()
	*s = append(*s, v) // add another bucket
	lock.Unlock()
}

// BPop removes a value from the bottom of the DES and returns it.
func (s *StringDES) BPop() string {
	if len(*s) == 0 {
		return ""
	}
	var v string
	lock.Lock()
	// retrieves the first item for returning
	// reslices the slice to place the second item at the beginning
	// also shrinks the slice's capacity
	v, *s = (*s)[len(*s)-1], (*s)[:len(*s)-1:len(*s)-1]
	lock.Unlock()
	return v
}

// BPeek returns the bottom value from the DES without removing it
func (s *StringDES) BPeek() string {
	if len(*s) == 0 {
		return ""
	}
	var v string
	lock.RLock()
	v = (*s)[len(*s)-1]
	lock.RUnlock()
	return v
}
