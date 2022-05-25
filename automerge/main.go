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
		base = flag.String("base", "https://review.dev.storj.io/", "gerrit base url")
		user = flag.String("user", "", "username for the performed actions")
		pass = flag.String("pass", "", "password for the user performing actions")
	)
	flag.Parse()

	if !strings.HasSuffix(*base, "/") {
		*base = *base + "/"
	}

	cl := &Client{
		base: *base,
		user: *user,
		pass: *pass,
	}

	if err := run(context.Background(), cl); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run(ctx context.Context, cl *Client) error {
	cis, err := cl.GetChangeInfos()
	if err != nil {
		return errs.Wrap(err)
	}

	if len(cis) == 0 {
		fmt.Println("No changes ready...")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 4, 4, 4, ' ', 0)
	defer tw.Flush()

	fmt.Fprintln(tw, "change\tnumber\tmergable\tsubmittable\thas_no_parents\thas_verified\tverified\treviewed\trev")
	for _, ci := range cis {
		fmt.Fprintf(tw, "%s\t%d\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
			ci.ViewURL(cl.base),
			ci.Number,
			ci.Mergeable,
			ci.Submittable,
			ci.HasNoParents(),
			ci.HasVerified(),
			ci.Verified(),
			ci.Reviewed(),
			ci.CurrentRevision,
		)
	}

	tw.Flush()
	fmt.Println()

	if submits := filter(cis, (*ChangeInfo).CanMerge); len(submits) > 0 {
		fmt.Println("Submit", submits[0].ViewURL(cl.base))
		return cl.Submit(submits[0])
	}

	if all(cis, (*ChangeInfo).HasVerified) {
		if rebases := filter(cis, (*ChangeInfo).CanRebase); len(rebases) > 0 {
			fmt.Println("Rebase", rebases[0].ViewURL(cl.base))
			return cl.Rebase(rebases[0])
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
