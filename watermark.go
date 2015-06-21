// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package upload

import (
	"errors"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Pos int

// 水印的位置
const (
	TopLeft Pos = iota
	TopRight
	BottomLeft
	BottomRight
	Center
)

var ErrUnsupportWatermarkType = errors.New("不支持的水印类型")

// 允许做水印的图片
var watermarkExts = []string{
	".gif", ".jpg", ".jpeg", ".png",
}

// 设置水印的相关参数。
// path为水印文件的路径；
// padding为水印在目标不图像上的留白大小；
// pos水印的位置。
func (u *Upload) SetWaterMark(path string, padding int, pos Pos) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var img image.Image
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
	case ".png":
		img, err = png.Decode(f)
	case ".gif":
		img, err = gif.Decode(f)
	default:
		return ErrUnsupportWatermarkType
	}
	if err != nil {
		return err
	}

	u.SetWatermarkImage(img, padding, pos)
	return nil
}

// 设置水印的image及位置。
func (u *Upload) SetWatermarkImage(img image.Image, padding int, pos Pos) {
	u.wmImage = img
	u.wmPadding = padding
	u.wmPos = pos
}

// 是否允许使用水印。
func (u *Upload) isAllowWatermark(ext string) bool {
	if u.wmImage == nil {
		return false
	}

	for _, e := range watermarkExts {
		if e == ext {
			return true
		}
	}
	return false
}

// 将水印作用于src，并将最终图像保存到dst中。
// srcExt为src文件的扩展名，由调用者保证其值为全部小写。
func (u *Upload) saveAsImage(dst io.Writer, src io.Reader, srcExt string) error {
	var srcImg image.Image
	var err error
	switch srcExt {
	case ".jpg", ".jpeg":
		srcImg, err = jpeg.Decode(src)
	case ".png":
		srcImg, err = png.Decode(src)
	case ".gif":
		srcImg, err = gif.Decode(src)
	default:
		return ErrUnsupportWatermarkType
	}
	if err != nil {
		return err
	}

	var point image.Point
	srcw := srcImg.Bounds().Dx()
	srch := srcImg.Bounds().Dy()
	switch u.wmPos {
	case TopLeft:
		point = image.Point{X: -u.wmPadding, Y: -u.wmPadding}
	case TopRight:
		point = image.Point{
			X: -(srcw - u.wmPadding - u.wmImage.Bounds().Dx()),
			Y: -u.wmPadding,
		}
	case BottomLeft:
		point = image.Point{
			X: -u.wmPadding,
			Y: -(srch - u.wmPadding - u.wmImage.Bounds().Dy()),
		}
	case BottomRight:
		point = image.Point{
			X: -(srcw - u.wmPadding - u.wmImage.Bounds().Dx()),
			Y: -(srch - u.wmPadding - u.wmImage.Bounds().Dy()),
		}
	case Center:
		point = image.Point{
			X: -(srcw - u.wmPadding - u.wmImage.Bounds().Dx()) / 2,
			Y: -(srch - u.wmPadding - u.wmImage.Bounds().Dy()) / 2,
		}
	}

	dstImg := image.NewNRGBA64(srcImg.Bounds())
	draw.Draw(dstImg, dstImg.Bounds(), srcImg, image.ZP, draw.Src)
	draw.Draw(dstImg, dstImg.Bounds(), u.wmImage, point, draw.Over)

	switch srcExt {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(dst, dstImg, nil)
	case ".png":
		err = png.Encode(dst, dstImg)
	case ".gif":
		err = gif.Encode(dst, dstImg, nil)
		// default: // 由前一个Switch确保此处没有default的出现。
	}

	return nil
}
