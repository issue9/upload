// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package upload

import (
	"errors"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/watermark"
)

// 创建文件的默认权限，比如 Upload.dir 若不存在，会使用此权限创建目录。
const presetMode = fs.ModePerm

// 常用错误类型
var (
	errNotAllowExt  = errors.New("不允许的文件上传类型")
	errNotAllowSize = errors.New("文件上传大小超过最大设定值")
	errNoUploadFile = errors.New("客户端没有上传文件")
)

// Upload 用于处理文件上传
type Upload struct {
	fs        fs.FS
	dir       string
	format    string
	maxSize   int64
	exts      []string
	watermark *watermark.Watermark
	filenames func(string, string) string
}

func ErrNotAllowExt() error { return errNotAllowExt }

func ErrNotAllowSize() error { return errNotAllowSize }

func ErrNoUploadFile() error { return errNoUploadFile }

// New 声明文件上传的对象
//
// dir 上传文件的保存目录，若目录不存在，则会尝试创建；
//
// format 子目录的格式，只能是时间格式；
//
// maxSize 允许上传文件的最大尺寸，单位为 byte；
//
// f 设置文件名的生成方式，要求文件在同一目录下具有唯一性，其类型如下：
//
//	func(dir, filename string) string
//
// dir 为文件夹名称，以 / 结尾，filename 为用户上传的文件名，
// 返回值是 dir + filename 的路径，实现者可能要调整 filename 的值，以保证在 dir 下唯一；
//
// exts 允许的扩展名，若为空，将不允许任何文件上传。
func New(dir, format string, maxSize int64, f func(string, string) string, exts ...string) (*Upload, error) {
	// 确保所有的后缀名都是以.作为开始符号的。
	es := make([]string, 0, len(exts))
	for _, ext := range exts {
		if ext[0] != '.' {
			ext = "." + ext
		}
		es = append(es, strings.ToLower(ext))
	}

	// 确保 dir 最后一个字符为目录分隔符。
	last := dir[len(dir)-1]
	if last != '/' && last != filepath.Separator {
		dir += string(filepath.Separator)
	}

	last = format[len(format)-1]
	if last != '/' && last != filepath.Separator {
		format += string(filepath.Separator)
	}

	// 若不存在目录，则尝试创建
	if err := os.MkdirAll(dir, presetMode); err != nil {
		return nil, err
	}

	// 确保 dir 目录存在。
	// NOTE:此处的 dir 最后个字符为/，所以不用判断是否为目录。
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}

	return &Upload{
		fs:        os.DirFS(dir),
		dir:       dir,
		format:    format,
		maxSize:   maxSize,
		exts:      es,
		filenames: f,
	}, nil
}

func (u *Upload) Open(name string) (fs.File, error) { return u.fs.Open(name) }

// 判断扩展名是否符合要求
func (u *Upload) isAllowExt(ext string) bool {
	if len(ext) == 0 { // 没有扩展名，一律过滤
		return false
	}

	for _, e := range u.exts {
		if e == ext {
			return true
		}
	}
	return false
}

// Dir 获取上传文件的保存目录
func (u *Upload) Dir() string { return u.dir }

// Do 执行上传的操作
//
// 若是多文件上传，其中某一个文件不符合要求，会中断后续操作，
// 但是已经处理成功的也会返回给用户，所以可能会出现两个返回参数都不为 nil 的情况。
//
// field 表示用于上传的字段名称；
// 返回的是相对于 [Upload.Dir] 目录的文件名列表。
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
			return ret, err // 已经上传的内容 ret，需要正常返回给用户
		}

		ret = append(ret, path)
	}

	return ret, nil
}

// 将上传的文件移到 u.Dir 目录下
//
// 返回相对于 u.Dir 的地址
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

	relDir := time.Now().Format(u.format)
	dir := u.dir + relDir
	if err = os.MkdirAll(dir, presetMode); err != nil { // 若路径不存在，则创建
		return "", err
	}

	p := u.filenames(dir, head.Filename)
	destFile, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, srcFile); err != nil {
		return "", err
	}

	if u.watermark != nil && watermark.IsAllowExt(ext) {
		if err = u.watermark.Mark(destFile, ext); err != nil {
			return "", err
		}
	}

	return path.Join(relDir, filepath.Base(p)), nil
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
// pos 水印的位置。
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

// Filename 在 dir 下为 s 生成唯一文件名
func Filename(dir, s string) string {
	ext := filepath.Ext(s)
	base := strings.TrimSuffix(s, ext)

	count := 1
	path := dir + s

RET:
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path
	}

	path = dir + base + "_" + strconv.Itoa(count) + ext
	count++
	goto RET
}
