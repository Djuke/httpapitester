package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("v2.1.2")
	if len(os.Args) != 2 || os.Args[1] == "" {
		fmt.Println("HTTP API tester is a tool to test HTTP APIs\n\nusage: httpapitester [test suite file]\n")
		os.Exit(0)
	}
	testSuiteFP := os.Args[1]

	b, err := ioutil.ReadFile(testSuiteFP)
	if err != nil {
		log.Fatal(err)
	}
	testSuite := &TestSuite{fp: filepath.Dir(testSuiteFP)}
	if err := json.Unmarshal(b, testSuite); err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); !ok {
			log.Fatal(err)
		}
	}
	testSuite.Run()
}
