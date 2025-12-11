// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package upload

import (
	"os"
	"testing"

	"github.com/issue9/assert/v4"
)

var (
	_ Saver   = &localSaver{}
	_ Deleter = &localSaver{}
)

func TestFilenameAI(t *testing.T) {
	a := assert.New(t, false)

	f := FilenameAI(os.DirFS("./testdir/"), "abc", "")
	a.Equal(f, "abc")

	f = FilenameAI(os.DirFS("./"), "testdir/file.xml", ".xml")
	a.Equal(f, "testdir/file_1.xml")
}
