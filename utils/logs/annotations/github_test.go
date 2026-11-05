/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubFormatter(t *testing.T) {
	formatter := GitHubFormatter{}
	line := 12
	column := 4
	annotation := &Annotation{Severity: SeverityError, Message: "bad\nmessage", File: "pkg/file.go", Line: &line, Column: &column}
	assert.Equal(t,
		"::error file=pkg/file.go,line=12,col=4::bad%0Amessage",
		formatter.Format(annotation),
	)
}
