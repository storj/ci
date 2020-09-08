// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package a

import (
	"context"

	monkit "github.com/spacemonkeygo/monkit/v3"
)

var mon = monkit.Package()

func _() {
	mon.Task() // want "monitoring not started"
}

func _() {
	ctx := context.Background()
	mon.Task()(&ctx) // want "monitoring not stopped"
}

func _() {
	var err error
	ctx := context.Background()
	mon.Task()(&ctx)(&err)
}

func _() {
	ctx := context.Background()
	_ = mon.Task()(&ctx)
}

func _() {
	_ = mon.Task()
}
