/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package collection

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemove(t *testing.T) {
	// Given a list of strings and a string
	// Returns the list without the string
	tests := []struct {
		SrcList      []string
		ToDelete     []string
		ExpectedList []string
	}{
		{
			SrcList:      []string{"a", "b", "c", "d"},
			ToDelete:     []string{"c"},
			ExpectedList: []string{"a", "b", "d"},
		},
		{
			SrcList:      []string{"a", "b", "c", "d"},
			ToDelete:     []string{"h"},
			ExpectedList: []string{"a", "b", "c", "d"},
		},
		{
			SrcList:      []string{"a", "b", "c", "d"},
			ToDelete:     []string{"d"},
			ExpectedList: []string{"a", "b", "c"},
		},
		{
			SrcList:      []string{"d", "d", "d", "d"},
			ToDelete:     []string{"d"},
			ExpectedList: []string{},
		},
		{
			SrcList:      []string{},
			ToDelete:     []string{"d"},
			ExpectedList: []string{},
		},
		{
			SrcList:      []string{},
			ToDelete:     []string{"d", "e"},
			ExpectedList: []string{},
		},
		{
			SrcList:      []string{"a", "b", "c", "d"},
			ToDelete:     []string{"a", "b", "d"},
			ExpectedList: []string{"c"},
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("test_%v", i), func(t *testing.T) {
			t.Parallel()
			newList := Remove(test.SrcList, test.ToDelete...)
			assert.Equal(t, test.ExpectedList, newList)
		})
	}
}
