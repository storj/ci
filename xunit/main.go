// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mfridman/tparse/parse"
)

var xunit = flag.String("out", "", "xunit output file")

func main() {
	flag.Parse()

	if *xunit == "" {
		_, _ = fmt.Fprintf(os.Stderr, "xunit file not specified\n")
		os.Exit(1)
	}

	var buffer bytes.Buffer
	stdin := io.TeeReader(os.Stdin, &buffer)

	pkgs, err := parse.Process(stdin, parse.WithFollowOutput(true))
	errcode := pkgs.ExitCode()
	if err != nil {
		if errors.Is(err, parse.ErrNotParsable) {
			_, _ = fmt.Fprintf(os.Stderr, "tparse error: no parseable events: call go test with -json flag\n\n")
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "tparse error: %v\n\n", err)
		}
		errcode = 1
	}
	defer os.Exit(errcode)

	output, err := os.Create(*xunit)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "create error: %v\n\n", err)
		return
	}
	defer func() {
		if err := output.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "close error: %v\n\n", err)
		}
	}()

	_, _ = output.Write([]byte(xml.Header))

	encoder := &printingEncoder{xml.NewEncoder(output)}
	encoder.Indent("", "\t")
	defer encoder.Flush()

	encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "testsuites"}, Attr: nil})
	defer encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "testsuites"}})

	for _, pkg := range pkgs.Packages {
		failed := TestsByAction(pkg, parse.ActionFail)
		skipped := TestsByAction(pkg, parse.ActionSkip)
		passed := TestsByAction(pkg, parse.ActionPass)

		skipped = withoutEmptyName(skipped)

		all := []*parse.Test{}
		all = append(all, failed...)
		all = append(all, skipped...)
		all = append(all, passed...)

		if !pkg.HasPanic && (pkg.NoTests || len(all) == 0) {
			continue
		}

		func() {
			encoder.EncodeToken(xml.StartElement{
				Name: xml.Name{Local: "testsuite"},
				Attr: []xml.Attr{
					{Name: xml.Name{Local: "name"}, Value: pkg.Summary.Package},
					{Name: xml.Name{Local: "time"}, Value: fmt.Sprintf("%.2f", pkg.Summary.Elapsed)},

					{Name: xml.Name{Local: "tests"}, Value: strconv.Itoa(len(all))},
					{Name: xml.Name{Local: "failures"}, Value: strconv.Itoa(len(failed))},
					{Name: xml.Name{Local: "skips"}, Value: strconv.Itoa(len(skipped))},
				},
			})
			defer encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "testsuite"}})

			if pkg.HasPanic {
				encoder.EncodeToken(xml.StartElement{
					Name: xml.Name{Local: "testcase"},
					Attr: []xml.Attr{
						{Name: xml.Name{Local: "classname"}, Value: pkg.Summary.Package},
						{Name: xml.Name{Local: "name"}, Value: "Panic"},
					},
				})
				encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "failure"}, Attr: nil})
				encoder.EncodeToken(xml.CharData(eventOutput(pkg.PanicEvents)))
				encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "failure"}})

				encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "testcase"}})
			}

			for _, t := range all {
				t.SortEvents()
				func() {
					encoder.EncodeToken(xml.StartElement{
						Name: xml.Name{Local: "testcase"},
						Attr: []xml.Attr{
							{Name: xml.Name{Local: "classname"}, Value: t.Package},
							{Name: xml.Name{Local: "name"}, Value: t.Name},
							{Name: xml.Name{Local: "time"}, Value: fmt.Sprintf("%.2f", t.Elapsed())},
						},
					})
					defer encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "testcase"}})

					encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "system-out"}})
					encoder.EncodeToken(xml.CharData(eventOutput(t.Events)))
					encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "system-out"}})

					switch TestStatus(t) {
					case parse.ActionSkip:
						encoder.EncodeToken(xml.StartElement{
							Name: xml.Name{Local: "skipped"},
							Attr: []xml.Attr{
								{Name: xml.Name{Local: "message"}, Value: testStack(t)},
							},
						})
						encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "skipped"}})
					case parse.ActionFail:
						encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "failure"}, Attr: nil})
						encoder.EncodeToken(xml.CharData(testStack(t)))
						encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "failure"}})
					}
				}()
			}
		}()
	}
}

func testStack(t *parse.Test) string {
	t.SortEvents()
	for i, ev := range t.Events {
		beginningOfStack := strings.Contains(ev.Output, "--- PASS:") ||
			strings.Contains(ev.Output, "--- FAIL:") ||
			strings.Contains(ev.Output, "--- SKIP:")
		if !beginningOfStack {
			continue
		}

		var stack strings.Builder
		for _, ev := range t.Events[i:] {
			stack.WriteString(ev.Output)
		}
		return stack.String()
	}
	return ""
}

type printingEncoder struct {
	*xml.Encoder
}

func (encoder *printingEncoder) EncodeToken(token xml.Token) {
	err := encoder.Encoder.EncodeToken(token)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "encoder: failed encoding %v: %v\n", token, err)
	}
}

func (encoder *printingEncoder) Flush() {
	err := encoder.Encoder.Flush()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "encoder: failed to flush: %v\n", err)
	}
}

func eventOutput(events []*parse.Event) string {
	var out strings.Builder
	for _, event := range events {
		out.WriteString(event.Output)
	}
	return out.String()
}

func withoutEmptyName(tests []*parse.Test) []*parse.Test {
	out := tests[:0]
	for _, test := range tests {
		if test.Name != "" {
			out = append(out, test)
		}
	}
	return out
}

// TestStatus reports the outcome of the test represented as a single Action: pass, fail or skip.
//
// Custom status to check packages properly.
func TestStatus(t *parse.Test) parse.Action {

	// sort by time and scan for an action in reverse order.
	// The first action we come across (in reverse order) is
	// the outcome of the test, which will be one of pass|fail|skip.
	t.SortEvents()

	for i := len(t.Events) - 1; i >= 0; i-- {
		switch t.Events[i].Action {
		case parse.ActionPass:
			return parse.ActionPass
		case parse.ActionSkip:
			return parse.ActionSkip
		case parse.ActionFail:
			return parse.ActionFail
		}
	}

	if t.Name == "" {
		return parse.ActionPass
	}
	return parse.ActionFail
}

// TestsByAction returns all tests that identify as one of the following
// actions: pass, skip or fail.
//
// An empty slice if returned if there are no tests.
func TestsByAction(p *parse.Package, action parse.Action) []*parse.Test {
	tests := []*parse.Test{}

	for _, t := range p.Tests {
		if TestStatus(t) == action {
			tests = append(tests, t)
		}
	}

	return tests
}
