//go:build !linux

/*
 * Copyright (C) 2020-2024 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package find

import (
	"context"
	"regexp"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/proc"
)

func findProcessByRegex(ctx context.Context, re *regexp.Regexp) (processes []proc.IProcess, err error) {
	ps, err := proc.Ps(ctx)
	if err != nil || len(ps) == 0 {
		return
	}

	processes, err = parallelisation.Filter[proc.IProcess](ctx, 10, ps, func(iProcess proc.IProcess) bool {
		if iProcess == nil {
			return false
		}
		return collection.AnyTrue(re.MatchString(iProcess.Name()), re.MatchString(iProcess.Executable()), re.MatchString(iProcess.Cmdline()))
	})
	return
}
