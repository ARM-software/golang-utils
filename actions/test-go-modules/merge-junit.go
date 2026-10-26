/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type testSuite struct {
	XMLName xml.Name   `xml:"testsuite"`
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",innerxml"`
}

type testSuites struct {
	XMLName  xml.Name    `xml:"testsuites"`
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Errors   int         `xml:"errors,attr"`
	Skipped  int         `xml:"skipped,attr"`
	Time     string      `xml:"time,attr"`
	Suites   []testSuite `xml:"testsuite"`
}

func mergeReports(directory string) ([]byte, error) {
	paths, err := filepath.Glob(filepath.Join(directory, "*.xml"))
	if err != nil {
		return nil, err
	}

	combined := testSuites{}
	var seconds float64
	var count int
	for _, path := range paths {
		if filepath.Base(path) == "combined.xml" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var report testSuites
		if err := xml.Unmarshal(data, &report); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if report.XMLName.Local != "testsuites" {
			return nil, fmt.Errorf("%s: expected a testsuites root element", path)
		}
		elapsed, err := strconv.ParseFloat(report.Time, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid report time: %w", path, err)
		}
		seconds += elapsed
		combined.Tests += report.Tests
		combined.Failures += report.Failures
		combined.Errors += report.Errors
		for _, suite := range report.Suites {
			for _, attr := range suite.Attrs {
				if attr.Name.Local == "skipped" {
					skipped, err := strconv.Atoi(attr.Value)
					if err != nil {
						return nil, fmt.Errorf("%s: invalid skipped count: %w", path, err)
					}
					combined.Skipped += skipped
				}
			}
			combined.Suites = append(combined.Suites, suite)
		}
		count++
	}
	if count == 0 {
		return nil, fmt.Errorf("no per-module JUnit XML reports in %s", directory)
	}
	combined.Time = strconv.FormatFloat(seconds, 'f', 6, 64)
	data, err := xml.MarshalIndent(combined, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal combined JUnit report: %w", err)
	}
	return append([]byte(xml.Header), append(data, '\n')...), nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: merge-junit <report-directory> <output-file>")
		os.Exit(2)
	}
	data, err := mergeReports(os.Args[1])
	if err == nil {
		err = os.WriteFile(os.Args[2], data, 0o644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
