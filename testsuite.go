package main

import (
	"fmt"
	"os"
	"time"
)

type TestSuite struct {
	Default                *Test    `json:"default"`
	First                  []*Test  `json:"first,omitempty"`
	Includes               []string `json:"includes"`
	Last                   []*Test  `json:"last,omitempty"`
	total, count, ok, fail int
	startTime              time.Time
	fp string
}

func (ts *TestSuite) Run() {
	/*cp, err := os.Getwd()
	if err != nil {
		fmt.Printf("\033[1;31m%s\033[0m\n", err)
		os.Exit(1)
	}*/
	tests, err := GetTests(ts.fp, ts.Includes)
	if err != nil {
		fmt.Printf("\033[1;31m%s\033[0m\n", err)
		os.Exit(1)
	}
	ts.total = len(ts.First) + len(tests) + len(ts.Last)
	fmt.Printf("\033[1;37mExecuted %d of %d\033[0m", ts.count, ts.total)
	ts.Default.Prepare(nil)
	ts.startTime = time.Now()
	for _, t := range ts.First {
		if !ts.runTest(t) {
			fmt.Println("\n\033[1;31mone of the first tests failed I will not continue to execute the other tests\033[0m")
			os.Exit(1)
		}
	}
	for _, t := range tests {
		ts.runTest(t)
	}
	for _, t := range ts.Last {
		ts.runTest(t)
	}
}

func (ts *TestSuite) runTest(t *Test) bool {
	t.Prepare(ts.Default)
	ok := t.Run()
	ts.count++
	if ok {
		ts.ok++
	} else {
		ts.fail++
	}
	ts.printProgress()
	return ok
}

func (ts *TestSuite) printProgress() {
	fmt.Printf("\r\033[1;37mExecuted %d of %d\033[0m", ts.count, ts.total)
	if ts.fail > 0 {
		fmt.Printf(" \033[1;31m(%d FAILED)\033[0m", ts.fail)
	}
	if ts.count == ts.total && ts.fail == 0 {
		fmt.Printf("\r\033[1;32mExecuted %d of %d\033[0m", ts.ok, ts.total)
	}
	fmt.Printf(" (%v)", time.Since(ts.startTime))
	if ts.count == ts.total {
		fmt.Println("")
	}
}
