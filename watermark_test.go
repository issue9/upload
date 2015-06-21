// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package upload

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/issue9/assert"
)

func saveImage(a *assert.Assertion, u *Upload, dest, src string) error {
	destFile, err := os.Create(dest)
	a.NotError(err).NotNil(destFile)

	srcFile, err := os.Open(src)
	a.NotError(err).NotNil(srcFile)

	ext := strings.ToLower(filepath.Ext(src))

	return u.watermark.saveAsImage(destFile, srcFile, ext)
}

func TestUploadWatermark(t *testing.T) {
	a := assert.New(t)
	u, err := New("./testdir", 10*1024, "gif", ".png", ".GIF")
	a.NotError(err).NotNil(u)

	u.SetWaterMark("./testdata/watermark.jpg", 10, TopLeft)
	saveImage(a, u, "./testdata/output/jpg-jpg.jpg", "./testdata/background.jpg")

	u.SetWaterMark("./testdata/watermark.jpg", 10, TopRight)
	saveImage(a, u, "./testdata/output/png-jpg.jpg", "./testdata/background.png")

	u.SetWaterMark("./testdata/watermark.png", 10, BottomLeft)
	saveImage(a, u, "./testdata/output/jpg-png.jpg", "./testdata/background.jpg")

	u.SetWaterMark("./testdata/watermark.png", 10, BottomRight)
	saveImage(a, u, "./testdata/output/png-png.jpg", "./testdata/background.png")
}
