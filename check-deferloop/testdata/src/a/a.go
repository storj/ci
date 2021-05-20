// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package a

func _() {
	for {
		defer func() {}() // want "defer inside a loop"
	}
}
func _() {
	for {
		defer func() { // want "defer inside a loop"
			defer func() {}()
		}()
	}
}

func _() {
	for {
		_ = func() {
			defer func() {}()
		}
	}
}

func _() {
	for {
		func() {
			defer func() {}()
		}()
	}
}

func _() {
	for {
		x := func() int {
			defer func() {}()
			return 0
		}()
		_ = x
	}
}
