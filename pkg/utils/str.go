// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import "strings"

func BeforeSpace(value string) string {
	return Before(value, " ")
}

func AfterSpace(value string) string {
	return After(value, " ")
}

func Before(value string, a string) string {
	// Get substring before a string.
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	return value[0:pos]
}

func After(value string, a string) string {
	// Get substring after a string.
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:]
}
