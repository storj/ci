// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/zeebo/errs"
)

func main() {
	var (
		base = flag.String("base", "https://review.dev.storj.tools/", "gerrit base url")
		user = flag.String("user", os.Getenv("AUTOMERGE_USER"), "username for the performed actions")
		pass = flag.String("pass", os.Getenv("AUTOMERGE_PASS"), "password for the user performing actions")
		dry  = flag.Bool("dry", false, "dry run")
	)
	flag.Parse()

	if !strings.HasSuffix(*base, "/") {
		*base += "/"
	}

	cl := &Client{
		base: *base,
		user: *user,
		pass: *pass,
		dry:  *dry,
	}

	if err := run(context.Background(), cl); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run(ctx context.Context, cl *Client) error {
	cis, err := cl.GetChangeInfos(ctx)
	if err != nil {
		return errs.Wrap(err)
	}

	if len(cis) == 0 {
		fmt.Println("No changes ready...")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 4, 4, 4, ' ', 0)
	defer func() { _ = tw.Flush() }()

	_, _ = fmt.Fprintln(tw, "change\tnumber\tmergable\tsubmittable\thas_no_parents\thas_verified\tverified\tlatest_build\tnot_building\tcan_rebase\trev")
	for _, ci := range cis {
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
			ci.ViewURL(cl.base),
			ci.Number,
			ci.Mergeable,
			ci.Submittable,
			ci.HasNoParents(),
			ci.HasVerified(),
			ci.Verified(),
			ci.LatestBuildStarted(),
			ci.NotBuilding(),
			ci.CanRebase(),
			ci.CurrentRevision,
		)
	}

	_ = tw.Flush()
	fmt.Println()

	if submits := filter(cis, (*ChangeInfo).CanMerge); len(submits) > 0 {
		fmt.Println("Submit", submits[0].ViewURL(cl.base))
		return cl.Submit(ctx, submits[0])
	}

	byProject := make(map[string][]*ChangeInfo)
	for _, ci := range cis {
		byProject[ci.Project] = append(byProject[ci.Project], ci)
	}

	for _, cis := range byProject {
		if all(cis, (*ChangeInfo).NotBuilding) {
			if rebases := filter(cis, (*ChangeInfo).CanRebase); len(rebases) > 0 {
				for _, rebase := range rebases {
					fmt.Println("Rebase", rebase.ViewURL(cl.base))
					if err := cl.Rebase(ctx, rebase); err == nil {
						break
					}
				}
			}
		}
	}

	return nil
}

func filter[T any](xs []T, fn func(T) bool) (out []T) {
	for _, x := range xs {
		if fn(x) {
			out = append(out, x)
		}
	}
	return out
}

func all[T any](xs []T, fn func(T) bool) bool {
	for _, x := range xs {
		if !fn(x) {
			return false
		}
	}
	return true
}
