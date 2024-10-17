// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package upload

import (
	"testing"

	"github.com/issue9/assert/v4"
)

var _ Saver = &localSaver{}

func TestFilename(t *testing.T) {
	a := assert.New(t, false)

	f := Filename("./testdir/", "abc", "")
	a.Equal(f, "./testdir/abc")

	f = Filename("./testdir/", "file.xml", ".xml")
	a.Equal(f, "./testdir/file_1.xml")
}
