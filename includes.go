package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const includesFilename = "includes.json"

func GetTests(path string, includes []string) ([]*Test, error) {
	tests := make([]*Test, 0)
	var newTests []*Test
	var err error
	var newIncludes []string
	var fileInfo os.FileInfo
	for _, rfp := range includes {
		fp := filepath.Join(path, rfp)
		fileInfo, err = os.Stat(fp)
		if err != nil {
			return nil, err
		}
		if fileInfo.IsDir() {
			newTests, err = ReadDir(fp)
		} else if filepath.Base(fp) == includesFilename {
			newIncludes, err = ReadIncludesFile(fp)
			if err == nil {
				newTests, err = GetTests(filepath.Dir(fp), newIncludes)
			}
		} else {
			newTests, err = ReadTestsFile(fp)
		}
		if err != nil {
			return nil, err
		}
		tests = append(tests, newTests...)
	}
	return tests, nil
}

// Skips reading the entire directory when an includes file is present.
// Includes files preceeds above reading a directory.
func ReadDir(p string) ([]*Test, error) {
	var includes []string
	var err error
	fp := filepath.Join(p, includesFilename)
	if _, err = os.Stat(fp); err == nil {
		includes, err = ReadIncludesFile(fp)
		if err != nil {
			return nil, err
		}
	} else {
		files, err := ioutil.ReadDir(p)
		if err != nil {
			return nil, err
		}
		includes = make([]string, 0, len(files))
		for _, f := range files {
			includes = append(includes, f.Name())
		}
	}
	return GetTests(p, includes)
}

func ReadIncludesFile(fp string) ([]string, error) {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	includes := make([]string, 0)
	if err := json.Unmarshal(b, &includes); err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); !ok {
			return nil, fmt.Errorf("%s: %s", fp, err)
		}
	}
	return includes, nil
}

func ReadTestsFile(fp string) ([]*Test, error) {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}
	tests := make([]*Test, 0)
	if err := json.Unmarshal(b, &tests); err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); !ok {
			return nil, fmt.Errorf("%s: %s", fp, err)
		}
	}
	return tests, nil
}
