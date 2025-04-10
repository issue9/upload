// SPDX-FileCopyrightText: 2015-2025 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -m -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/zh.yaml

package upload

import (
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/watermark"
)

// 常用错误类型
var (
	errNotAllowExt  = localeutil.Error("not allow extension")
	errNotAllowSize = localeutil.Error("the upload file is too large")
	errNoUploadFile = localeutil.Error("no upload file")
)

// Upload 用于处理文件上传
type Upload struct {
	saver     Saver
	maxSize   int64
	exts      []string
	watermark *watermark.Watermark
	moveMux   sync.Mutex
}

func ErrNotAllowExt() error { return errNotAllowExt }

func ErrNotAllowSize() error { return errNotAllowSize }

func ErrNoUploadFile() error { return errNoUploadFile }

// New 声明文件上传的对象
//
// maxSize 允许上传文件的最大尺寸，单位为 byte；
//
// exts 允许的扩展名，若为空，将不允许任何文件上传；
func New(saver Saver, maxSize int64, exts ...string) *Upload {
	// 确保所有的后缀名都是以.作为开始符号的。
	es := make([]string, 0, len(exts))
	for _, ext := range exts {
		if ext[0] != '.' {
			ext = "." + ext
		}
		es = append(es, strings.ToLower(ext))
	}

	return &Upload{
		saver:   saver,
		maxSize: maxSize,
		exts:    es,
	}
}

// Open [fs.FS] 接口
//
// 等同于 [Saver.Open]
func (u *Upload) Open(name string) (fs.File, error) { return u.saver.Open(name) }

// 判断扩展名是否符合要求
func (u *Upload) isAllowExt(ext string) bool { return slices.Contains(u.exts, ext) }

// Do 执行上传的操作
//
// field 表示用于上传的字段名称；
//
// NOTE: 若是多文件上传，其中某一个文件不符合要求，会中断后续操作，
// 但是已经处理成功的也会返回给用户，所以可能会出现两个返回参数都不为 nil 的情况。
func (u *Upload) Do(field string, r *http.Request) ([]string, error) {
	if err := r.ParseMultipartForm(u.maxSize); err != nil {
		return nil, err
	}

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, ErrNoUploadFile()
	}

	heads := r.MultipartForm.File[field]
	if len(heads) == 0 {
		return nil, ErrNoUploadFile()
	}

	ret := make([]string, 0, len(heads))
	for _, head := range heads {
		path, err := u.moveFile(head)
		if err != nil {
			return ret, err // 如果出错，则将已经移入目录的文件列表返回给用户。
		}
		ret = append(ret, path)
	}
	return ret, nil
}

// 将上传的文件移到 [Saver]
func (u *Upload) moveFile(head *multipart.FileHeader) (string, error) {
	if head.Size > u.maxSize {
		return "", ErrNotAllowSize()
	}

	ext := strings.ToLower(filepath.Ext(head.Filename))
	if !u.isAllowExt(ext) {
		return "", ErrNotAllowExt()
	}

	srcFile, err := head.Open()
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	if u.watermark != nil && watermark.IsAllowExt(ext) {
		// NOTE: srcFile 目前是可以转换为 io.ReadWriteSeeker 类型的
		if err = u.watermark.Mark(srcFile.(io.ReadWriteSeeker), ext); err != nil {
			return "", err
		}
	}

	p, err := u.saver.Save(srcFile, head.Filename, ext)
	if err != nil {
		return "", err
	}

	return p, nil
}

// SetWatermarkFile 设置水印的相关参数
//
// path 为水印文件的路径；
// padding 为水印在目标不图像上的留白大小；
// pos 水印的位置。
func (u *Upload) SetWatermarkFile(path string, padding int, pos watermark.Pos) error {
	w, err := watermark.NewFromFile(path, padding, pos)
	if err == nil {
		u.SetWatermark(w)
	}
	return err
}

// SetWatermarkFS 设置水印的相关参数
//
// path 为水印文件的路径；
// padding 为水印在目标不图像上的留白大小；
// pos 水印的位置；
func (u *Upload) SetWatermarkFS(fs fs.FS, path string, padding int, pos watermark.Pos) error {
	w, err := watermark.NewFromFS(fs, path, padding, pos)
	if err == nil {
		u.SetWatermark(w)
	}
	return err
}

// SetWatermark 设置水印的相关参数
//
// 如果 w 为 nil，则表示取消水印
func (u *Upload) SetWatermark(w *watermark.Watermark) { u.watermark = w }

// AllowWatermarkExts 允许的水印类型
func AllowWatermarkExts() []string { return watermark.AllowExts() }
