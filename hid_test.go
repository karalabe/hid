// hid - Gopher Interface Devices (USB HID)
// Copyright (c) 2017 Péter Szilágyi. All rights reserved.
//
// This file is released under the 3-clause BSD license. Note however that Linux
// support depends on libusb, released under LGNU GPL 2.1 or later.

package hid

import (
	"fmt"
	"sync"
	"testing"
)

// TestThreadedEnumerate Tests that HID enumeration can be called concurrently from multiple threads.
func TestThreadedEnumerate(t *testing.T) {
	var (
		threads         = 8
		errs    []error = make([]error, threads)
		pend    sync.WaitGroup
	)
	for i := 0; i < threads; i++ {
		pend.Add(1)

		go func(index int) {
			defer pend.Done()
			for j := 0; j < 512; j++ {
				if _, err := Enumerate(uint16(index), 0); err != nil {
					errs[index] = fmt.Errorf("thread %d, iter %d: failed to enumerate-hid: %v", index, j, err)
					break
				}
			}
		}(i)
	}
	pend.Wait()
	for _, err := range errs {
		if err != nil {
			t.Error(err)
		}
	}
}
