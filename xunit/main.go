// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

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

	pkgs, err := ProcessWithEcho(stdin)
	errcode := pkgs.ExitCode()
	if err != nil {
		if errors.Is(err, parse.ErrNotParseable) {
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

	for _, pkg := range pkgs {
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
								{Name: xml.Name{Local: "message"}, Value: t.Stack()},
							},
						})
						encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "skipped"}})
					case parse.ActionFail:
						encoder.EncodeToken(xml.StartElement{Name: xml.Name{Local: "failure"}, Attr: nil})
						encoder.EncodeToken(xml.CharData(t.Stack()))
						encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: "failure"}})
					}
				}()
			}
		}()
	}
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

func eventOutput(events parse.Events) string {
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

// Code based on: https://github.com/mfridman/tparse/blob/master/parse/process.go#L27

// ProcessWithEcho processes go test -json output and echos the usual output to stdout.
func ProcessWithEcho(r io.Reader) (parse.Packages, error) {
	pkgs := parse.Packages{}

	var hasRace bool

	var scan bool
	var badLines int

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		// Scan up-to 50 lines for a parseable event, if we get one, expect
		// no errors to follow until EOF.
		event, err := parse.NewEvent(scanner.Bytes())
		if err != nil {
			badLines++
			if scan || badLines > 50 {
				var jserr *json.SyntaxError
				if errors.As(err, &jserr) {
					return nil, parse.ErrNotParseable
				}
				return nil, err
			}
			continue
		}
		scan = true

		if line := strings.TrimRightFunc(event.Output, unicode.IsSpace); line != "" {
			_, _ = fmt.Fprintln(os.Stdout, line)
		}

		pkg, ok := pkgs[event.Package]
		if !ok {
			pkg = parse.NewPackage()
			pkgs[event.Package] = pkg
		}

		if event.IsPanic() {
			pkg.HasPanic = true
			pkg.Summary.Action = parse.ActionFail
			pkg.Summary.Package = event.Package
			pkg.Summary.Test = event.Test
		}
		// Short circuit output when panic is detected.
		if pkg.HasPanic {
			pkg.PanicEvents = append(pkg.PanicEvents, event)
			continue
		}

		if event.IsRace() {
			hasRace = true
		}

		if event.IsCached() {
			pkg.Cached = true
		}

		if event.NoTestFiles() {
			pkg.NoTestFiles = true
			// Manually mark [no test files] as "pass", because the go test tool reports the
			// package Summary action as "skip".
			pkg.Summary.Package = event.Package
			pkg.Summary.Action = parse.ActionPass
		}
		if event.NoTestsWarn() {
			// One or more tests within the package contains no tests.
			pkg.NoTestSlice = append(pkg.NoTestSlice, event)
		}

		if event.NoTestsToRun() {
			// Only pkgs marked as "pass" will contain a summary line appended with [no tests to run].
			// This indicates one or more tests is marked as having no tests to run.
			pkg.NoTests = true
			pkg.Summary.Package = event.Package
			pkg.Summary.Action = parse.ActionPass
		}

		if event.LastLine() {
			pkg.Summary = event
			continue
		}

		cover, ok := event.Cover()
		if ok {
			pkg.Cover = true
			pkg.Coverage = cover
		}

		// special case for tooling checking
		if event.Action == parse.ActionOutput && strings.HasPrefix(event.Output, "FAIL\t") {
			event.Action = parse.ActionFail
		}

		if !Discard(event) {
			pkg.AddEvent(event)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("bufio scanner error: %w", err)
	}
	if !scan {
		return nil, parse.ErrNotParseable
	}
	if hasRace {
		return pkgs, parse.ErrRaceDetected
	}

	return pkgs, nil
}

// Discard checks whether the event should be ignored.
func Discard(e *parse.Event) bool {
	for i := range updates {
		if strings.HasPrefix(e.Output, updates[i]) {
			return true
		}
	}
	return false
}

var (
	updates = []string{
		"=== RUN   ",
		"=== PAUSE ",
		"=== CONT  ",
	}
)

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
