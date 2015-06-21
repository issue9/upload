// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package upload

import (
	"io"
	"os"
	"testing"

	"github.com/issue9/assert"
)

func output(a *assert.Assertion, pos Pos, bgType, waterType string) {
	water := "./testdata/watermark" + waterType
	src := "./testdata/background" + bgType
	dest := "./testdata/output/" + waterType[1:] + bgType
	// 复制文件到output目录下，并重命名。
	destFile, err := os.Create(dest)
	a.NotError(err).NotNil(destFile)

	srcFile, err := os.Open(src)
	a.NotError(err).NotNil(srcFile)

	n, err := io.Copy(destFile, srcFile)
	a.NotError(err).True(n > 0)

	destFile.Close()
	srcFile.Close()

	// 添加水印
	w, err := NewWatermark(water, 10, pos)
	a.NotError(err).NotNil(w)
	a.NotError(w.Mark(dest))
}

func TestUploadWatermark(t *testing.T) {
	a := assert.New(t)

	output(a, TopLeft, ".jpg", ".jpg")
	output(a, TopRight, ".jpg", ".png")
	output(a, BottomLeft, ".png", ".jpg")
	output(a, BottomRight, ".png", ".png")
}
