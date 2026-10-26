/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package main

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
)

func TestMergeReports(t *testing.T) {
	directory := t.TempDir()
	reports := map[string]string{
		"first.xml": `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="2" failures="1" errors="0" time="1.25">
  <testsuite name="example/first" tests="2" failures="1" time="1.25" timestamp="2026-10-26T00:00:00Z">
    <properties><property name="go.version" value="go1.26.1"/></properties>
    <testcase classname="example/first" name="TestFails"><failure message="Failed">assertion failed</failure></testcase>
  </testsuite>
</testsuites>`,
		"second.xml": `<testsuites tests="1" failures="0" errors="0" time="0.5">
  <testsuite name="example/second" tests="1" failures="0" skipped="1" time="0.5">
    <testcase classname="example/second" name="TestSkipped"><skipped message="not supported"/></testcase>
  </testsuite>
</testsuites>`,
	}
	for name, report := range reports {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(report), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Regenerating the report must never merge the previous combined result.
	if err := os.WriteFile(filepath.Join(directory, "combined.xml"), []byte(reports["first.xml"]), 0o644); err != nil {
		t.Fatal(err)
	}

	data, err := mergeReports(directory)
	if err != nil {
		t.Fatal(err)
	}
	var combined testSuites
	if err := xml.Unmarshal(data, &combined); err != nil {
		t.Fatalf("combined report is not valid XML: %v", err)
	}
	if combined.Tests != 3 || combined.Failures != 1 || combined.Errors != 0 || combined.Skipped != 1 || combined.Time != "1.750000" {
		t.Fatalf("unexpected combined totals: %+v", combined)
	}
	if len(combined.Suites) != 2 || !bytes.Contains(data, []byte(`name="example/first"`)) || !bytes.Contains(data, []byte(`name="example/second"`)) {
		t.Fatalf("missing per-module suites: %s", data)
	}
	for _, content := range []string{`name="go.version"`, `assertion failed`, `message="not supported"`, `timestamp="2026-10-26T00:00:00Z"`} {
		if !bytes.Contains(data, []byte(content)) {
			t.Fatalf("missing report detail %q: %s", content, data)
		}
	}
}

func TestMergeReportsRejectsInvalidInput(t *testing.T) {
	directory := t.TempDir()
	if _, err := mergeReports(directory); err == nil {
		t.Fatal("expected an error when no module reports exist")
	}
	if err := os.WriteFile(filepath.Join(directory, "invalid.xml"), []byte(`<testsuites><testsuite>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mergeReports(directory); err == nil {
		t.Fatal("expected an error for malformed JUnit XML")
	}
}
