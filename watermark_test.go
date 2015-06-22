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

// 复制文件到output目录下，并重命名。
func copyBackgroundFile(a *assert.Assertion, dest, src string) {
	destFile, err := os.Create(dest)
	a.NotError(err).NotNil(destFile)

	srcFile, err := os.Open(src)
	a.NotError(err).NotNil(srcFile)

	n, err := io.Copy(destFile, srcFile)
	a.NotError(err).True(n > 0)

	destFile.Close()
	srcFile.Close()
}

// 输出各种组合的水印图片。
// bgExt 表示背景图片的扩展名。
// water 表示水印图片的扩展名。
func output(a *assert.Assertion, pos Pos, bgExt, waterExt string) {
	water := "./testdata/watermark" + waterExt
	src := "./testdata/background" + bgExt
	dest := "./testdata/output/" + waterExt[1:] + bgExt

	copyBackgroundFile(a, dest, src)

	// 添加水印
	w, err := NewWatermark(water, 10, pos)
	a.NotError(err).NotNil(w)
	a.NotError(w.MarkFile(dest))
}

func TestUploadWatermark(t *testing.T) {
	a := assert.New(t)

	output(a, TopLeft, ".jpg", ".jpg")
	output(a, TopRight, ".jpg", ".png")
	output(a, Center, ".jpg", ".gif")

	output(a, BottomLeft, ".png", ".jpg")
	output(a, BottomRight, ".png", ".png")
	output(a, Center, ".png", ".gif")

	output(a, BottomLeft, ".gif", ".jpg")
	output(a, BottomRight, ".gif", ".png")
	output(a, Center, ".gif", ".gif")
}

// BenchmarkWater_MakeImage_500xJPEG	   50000	     30030 ns/op
func BenchmarkWater_MakeImage_500xJPEG(b *testing.B) {
	a := assert.New(b)

	copyBackgroundFile(a, "./testdata/output/bench.jpg", "./testdata/background.jpg")

	w, err := NewWatermark("./testdata/watermark.jpg", 10, TopLeft)
	a.NotError(err).NotNil(w)

	file, err := os.OpenFile("./testdata/output/bench.jpg", os.O_RDWR, os.ModePerm)
	a.NotError(err).NotNil(file)
	defer file.Close()

	for i := 0; i < b.N; i++ {
		w.Mark(file, ".jpg")
	}
}

// BenchmarkWater_MakeImage_500xPNG	  500000	      2482 ns/op
func BenchmarkWater_MakeImage_500xPNG(b *testing.B) {
	a := assert.New(b)

	copyBackgroundFile(a, "./testdata/output/bench.png", "./testdata/background.png")

	w, err := NewWatermark("./testdata/watermark.png", 10, TopLeft)
	a.NotError(err).NotNil(w)

	file, err := os.OpenFile("./testdata/output/bench.png", os.O_RDWR, os.ModePerm)
	a.NotError(err).NotNil(file)
	defer file.Close()

	for i := 0; i < b.N; i++ {
		w.Mark(file, ".png")
	}
}
