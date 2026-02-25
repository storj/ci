// Copyright (C) 2026 Storj Labs, Inc.
// See LICENSE for copying information.

package a

// --- Bad: outer variable modified without reset ---

func badAppend() {
	rows := []int{}
	WithRetry(func() {
		rows = append(rows, 123) // want `variable "rows" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = rows
}

func badIncrement() {
	count := 0
	WithRetry(func() {
		count++ // want `variable "count" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = count
}

func badCompoundAssign() {
	var result int
	WithRetry(func() {
		result += doWork() // want `variable "result" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = result
}

func badMethodRetry() {
	r := Retrier{}
	var total int
	r.ReadWriteTransaction(func(tx int) error {
		total += tx // want `variable "total" is modified inside retry callback but not reset at the top of the callback`
		return nil
	})
	_ = total
}

func badReadTransaction() {
	r := Retrier{}
	var total int
	r.ReadTransaction(func(tx int) error {
		total += tx // want `variable "total" is modified inside retry callback but not reset at the top of the callback`
		return nil
	})
	_ = total
}

func badPartialReset() {
	var a, b int
	WithRetry(func() {
		a = 0
		// b is not reset
		a += doWork()
		b += doWork() // want `variable "b" is modified inside retry callback but not reset at the top of the callback`
	})
	_, _ = a, b
}

func badDecrement() {
	var x int
	WithRetry(func() {
		x-- // want `variable "x" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = x
}

func badSelfReferentialReset() {
	var items []string
	WithRetry(func() {
		items = append(items, "a") // want `variable "items" is modified inside retry callback but not reset at the top of the callback`
		items = append(items, "b") // want `variable "items" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = items
}

func badMultipleUnreset() {
	var deleted, segments int
	WithRetry(func() {
		deleted++  // want `variable "deleted" is modified inside retry callback but not reset at the top of the callback`
		segments++ // want `variable "segments" is modified inside retry callback but not reset at the top of the callback`
	})
	_, _ = deleted, segments
}

// --- Good: outer variable properly reset ---

func goodReset() {
	rows := []int{}
	WithRetry(func() {
		rows = []int{}
		rows = append(rows, 123)
	})
	_ = rows
}

func goodResetCounter() {
	count := 0
	WithRetry(func() {
		count = 0
		count++
	})
	_ = count
}

func goodResetMultiple() {
	var a, b int
	WithRetry(func() {
		a = 0
		b = 0
		a += doWork()
		b += doWork()
	})
	_, _ = a, b
}

func goodResetAndAssign() {
	var result int
	WithRetry(func() {
		result = doWork()
		result += 1
	})
	_ = result
}

func goodLocalOnly() {
	WithRetry(func() {
		x := 0
		x++
		_ = x
	})
}

func goodNoModification() {
	x := 5
	WithRetry(func() {
		_ = x // reading is fine
	})
}

func goodCallbackParam() {
	r := Retrier{}
	r.ReadWriteTransaction(func(tx int) error {
		tx++ // modifying a parameter is fine
		return nil
	})
}

func goodNestedFunc() {
	var x int
	WithRetry(func() {
		x = 0
		fn := func() {
			x++
		}
		fn()
	})
	_ = x
}

func goodNonRetryFunc() {
	var x int
	notRetry(func() {
		x = 1
	})
	_ = x
}

func goodResetStruct() {
	type metrics struct{ count int }
	var m metrics
	WithRetry(func() {
		m = metrics{}
		m.count++
	})
	_ = m
}

func goodPlainAssignOnly() {
	var a, b int
	WithRetry(func() {
		a = 1
		b = 2
	})
	_, _ = a, b
}

func goodMethodResetAndUse() {
	r := Retrier{}
	var result int
	r.ReadWriteTransaction(func(tx int) error {
		result = tx
		return nil
	})
	_ = result
}

// --- ReadWriteTransactionWithOptions (callback is not the last arg) ---

func badRWTWithOptions() {
	r := Retrier{}
	var total int
	r.ReadWriteTransactionWithOptions(func(tx int) error {
		total += tx // want `variable "total" is modified inside retry callback but not reset at the top of the callback`
		return nil
	}, Options{})
	_ = total
}

func goodRWTWithOptions() {
	r := Retrier{}
	var total int
	r.ReadWriteTransactionWithOptions(func(tx int) error {
		total = 0
		total += tx
		return nil
	}, Options{})
	_ = total
}

// --- Var declaration then assign (should not be a false positive) ---

func goodVarDeclThenAssign() {
	r := Retrier{}
	var info int
	r.WithTx(func(tx int) error {
		var err error
		info, err = work2(tx)
		_ = err
		return nil
	})
	_ = info
}

func goodShortVarDeclThenAssign() {
	r := Retrier{}
	var info int
	r.WithTx(func(tx int) error {
		extra := doWork()
		info = extra + tx
		return nil
	})
	_ = info
}

func goodMultipleVarDeclsThenAssign() {
	r := Retrier{}
	var a, b int
	r.WithTx(func(tx int) error {
		var x int
		var y int
		a = x
		b = y
		_, _ = x, y
		return nil
	})
	_, _ = a, b
}

// --- Named return values assigned inside callback (reported false positive) ---

func goodNamedReturnInCallback() (resultPieces int, err error) {
	r := Retrier{}
	_, err = r.ReadWriteTransactionWithOptions(func(tx int) error {
		resultPieces, err = CollectRow(tx, nil)
		if err != nil {
			return err
		}
		return nil
	}, Options{})
	return resultPieces, err
}

func goodNamedReturnWithIfBefore() (resultPieces int, err error) {
	r := Retrier{}
	updateRepairAt := true
	_, err = r.ReadWriteTransactionWithOptions(func(tx int) error {
		resultPieces, err = CollectRow(tx, nil)
		if err != nil {
			return err
		}
		if updateRepairAt {
			_ = tx
		}
		return nil
	}, Options{})
	return resultPieces, err
}

func goodNamedReturnWithIfBefore2() (resultPieces int, err error) {
	r := Retrier{}
	_, err = r.ReadWriteTransactionWithOptions(func(tx int) error {
		if resultPieces, err = CollectRow(tx, nil); err != nil {
			return err
		}
		return nil
	}, Options{})
	return resultPieces, err
}

// --- Closure in RHS should not count as self-reference ---

func goodClosureInRHS() (resultPieces int, err error) {
	r := Retrier{}
	_, err = r.ReadWriteTransactionWithOptions(func(tx int) error {
		resultPieces, err = CollectRowWithCallback(tx, func(item *int) error {
			err = decodeRow(item)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	}, Options{})
	return resultPieces, err
}

// --- Bad: map/slice index assignment without reset ---

func badMapIndexAssign() {
	r := Retrier{}
	inviteTokens := make(map[string]string)
	r.WithTx(func(tx int) error {
		inviteTokens["email"] = "token" // want `variable "inviteTokens" is modified inside retry callback but not reset at the top of the callback`
		return nil
	})
	_ = inviteTokens
}

func badSliceIndexAssign() {
	items := make([]int, 5)
	WithRetry(func() {
		items[0] = 42 // want `variable "items" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = items
}

// --- Good: map index assignment with reset ---

func goodMapIndexAssignWithReset() {
	r := Retrier{}
	inviteTokens := make(map[string]string)
	r.WithTx(func(tx int) error {
		inviteTokens = make(map[string]string)
		inviteTokens["email"] = "token"
		return nil
	})
	_ = inviteTokens
}

func goodMapIndexAssignWithClear() {
	r := Retrier{}
	inviteTokens := make(map[string]string)
	r.WithTx(func(tx int) error {
		clear(inviteTokens)
		inviteTokens["email"] = "token"
		return nil
	})
	_ = inviteTokens
}

// --- Bad: struct field assignment without reset ---

func badStructFieldAssign() {
	type stats struct{ count int }
	var s stats
	WithRetry(func() {
		s.count = 42 // want `variable "s" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = s
}

func badStructFieldAssignInTx() {
	r := Retrier{}
	type result struct{ total int }
	var res result
	r.WithTx(func(tx int) error {
		res.total = tx // want `variable "res" is modified inside retry callback but not reset at the top of the callback`
		return nil
	})
	_ = res
}

func badStructFieldIncrement() {
	type counter struct{ n int }
	var c counter
	WithRetry(func() {
		c.n++ // want `variable "c" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = c
}

// --- Good: nolint suppresses diagnostic ---

func goodNolintSuppressed() {
	rows := []int{}
	WithRetry(func() {
		rows = append(rows, 123) //check-retry:ignore
	})
	_ = rows
}

func goodNolintMapIndex() {
	r := Retrier{}
	inviteTokens := make(map[string]string)
	r.WithTx(func(tx int) error {
		inviteTokens["email"] = "token" //check-retry:ignore
		return nil
	})
	_ = inviteTokens
}

// --- Bad: range assigning to outer vars ---

func badRangeOuterKey() {
	var lastKey int
	items := []string{"a", "b"}
	WithRetry(func() {
		for lastKey = range items { // want `variable "lastKey" is modified inside retry callback but not reset at the top of the callback`
			_ = lastKey
		}
	})
	_ = lastKey
}

func badRangeOuterKeyValue() {
	var key int
	var val string
	items := []string{"a", "b"}
	WithRetry(func() {
		for key, val = range items { // want `variable "key" is modified inside retry callback but not reset at the top of the callback` `variable "val" is modified inside retry callback but not reset at the top of the callback`
			_, _ = key, val
		}
	})
	_, _ = key, val
}

// --- Good: range with := declares local vars ---

func goodRangeLocal() {
	items := []string{"a", "b"}
	WithRetry(func() {
		for k, v := range items {
			_, _ = k, v
		}
	})
}

// --- Bad: pointer dereference assignment ---

func badPointerDeref() {
	var x int
	p := &x
	WithRetry(func() {
		*p = 42 // want `variable "p" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = p
}

// --- Bad: delete on outer map ---

func badDeleteOuterMap() {
	m := map[string]int{"a": 1}
	WithRetry(func() {
		delete(m, "a") // want `variable "m" is modified inside retry callback but not reset at the top of the callback`
	})
	_ = m
}

func goodDeleteWithReset() {
	m := map[string]int{"a": 1}
	WithRetry(func() {
		m = map[string]int{"a": 1}
		delete(m, "a")
	})
	_ = m
}

// --- Bad: modification inside nested closure without reset ---

func badNestedClosureAppend() {
	r := Retrier{}
	var items []int
	r.ReadWriteTransactionWithOptions(func(tx int) error {
		forEach(func() {
			items = append(items, tx) // want `variable "items" is modified inside retry callback but not reset at the top of the callback`
		})
		return nil
	}, Options{})
	_ = items
}

func badNestedClosureIncrement() {
	var count int
	WithRetry(func() {
		forEach(func() {
			count++ // want `variable "count" is modified inside retry callback but not reset at the top of the callback`
		})
	})
	_ = count
}

func badNestedClosureAssign() {
	var result int
	WithRetry(func() {
		forEach(func() {
			result = doWork() // want `variable "result" is modified inside retry callback but not reset at the top of the callback`
		})
	})
	_ = result
}

// Good: nested closure modification with proper reset at top.

func goodNestedClosureWithReset() {
	var items []int
	WithRetry(func() {
		items = []int{}
		forEach(func() {
			items = append(items, 1)
		})
	})
	_ = items
}

// Good: modification of nested closure's own parameter.

func goodNestedClosureParam() {
	WithRetry(func() {
		forEachInt(func(x int) {
			x++
			_ = x
		})
	})
}

// --- Helpers ---

func forEach(fn func())    { fn() }
func forEachInt(fn func(int)) { fn(0) }

func doWork() int                                                          { return 1 }
func notRetry(fn func())                                                   { fn() }
func work2(tx int) (int, error)                                            { return tx, nil }
func CollectRowWithCallback(tx int, fn func(item *int) error) (int, error) { return tx, fn(nil) }
func decodeRow(item *int) error                                            { return nil }
