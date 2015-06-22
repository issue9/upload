// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package upload

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNotAllowExt  = errors.New("不允许的文件上传类型")
	ErrNotAllowSize = errors.New("文件上传大小超过最大设定值或是文件大小为0")
)

// 创建文件的默认权限，比如Upload.dir若不存在，会使用此权限创建目录。
const defaultMode os.FileMode = 0660

// Upload用于处理文件上传
type Upload struct {
	dir       string     // 上传文件保存的路径根目录
	maxSize   int64      // 允许的最大文件大小，以byte为单位
	exts      []string   // 允许的扩展名
	watermark *Watermark // 水印
}

// 声明一个Upload对象。
// dir 上传文件的保存目录，若目录不存在，则会尝试创建;
// maxSize 允许上传文件的最大尺寸，单位为byte；
// exts 允许的扩展名，若为空，将不允许任何文件上传。
func New(dir string, maxSize int64, exts ...string) (*Upload, error) {
	// 确保所有的后缀名都是以.作为开始符号的。
	es := make([]string, 0, len(exts))
	for _, ext := range exts {
		if ext[0] != '.' {
			ext = "." + ext
		}
		es = append(es, strings.ToLower(ext))
	}

	// 确保dir最后一个字符为目录分隔符。若传递的是一个文件，
	// 在结尾加个/后，后面的os.Stat会返回错误信息！
	last := dir[len(dir)-1]
	if last != '/' && last != filepath.Separator {
		dir = dir + string(filepath.Separator)
	}

	// 确保dir目录存在，若不存在则会尝试创建。
	if _, err := os.Stat(dir); err != nil && !os.IsExist(err) {
		if !os.IsNotExist(err) {
			return nil, err
		}

		// 尝试创建目录
		if err = os.MkdirAll(dir, 0660); err != nil {
			return nil, err
		}

		// 创建目录成功，重新获取状态
		if _, err = os.Stat(dir); err != nil {
			return nil, err
		}
	}

	return &Upload{
		dir:     dir,
		maxSize: maxSize,
		exts:    es,
	}, nil
}

// 判断扩展名是否符合要求。
// 由调用者保证ext参数为小写。
func (u *Upload) isAllowExt(ext string) bool {
	if len(ext) == 0 { // 没有扩展名，一律过滤
		return false
	}

	// 是否为允许的扩展名
	for _, e := range u.exts {
		if e == ext {
			return true
		}
	}
	return false
}

// 检测文件大小是否符合要求。
func (u *Upload) isAllowSize(file multipart.File) (bool, error) {
	var size int64

	switch f := file.(type) {
	case stater:
		stat, err := f.Stat()
		if err != nil {
			return false, err
		}
		size = stat.Size()
	case sizer:
		size = f.Size()
	default:
		return false, errors.New("上传文件时发生未知的错误")
	}

	return size > 0 && size <= u.maxSize, nil
}

func (u *Upload) getDestPath(ext string) string {
	n := time.Now()
	return n.Format("2006/01/02/") + strconv.Itoa(n.Nanosecond()) + ext
}

// 招行上传的操作。会检测上传文件是否符合要求，只要有一个文件不符合，就会中断上传。
// 返回的是相对于u.dir目录的文件名列表。
func (u *Upload) Do(field string, r *http.Request) ([]string, error) {
	r.ParseMultipartForm(32 << 20)
	heads := r.MultipartForm.File[field]
	ret := make([]string, len(heads))

	for _, head := range heads {
		file, err := head.Open()
		if err != nil {
			return nil, err
		}

		ext := strings.ToLower(filepath.Ext(head.Filename))
		if !u.isAllowExt(ext) {
			return nil, ErrNotAllowExt
		}

		ok, err := u.isAllowSize(file)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrNotAllowSize
		}

		path := u.getDestPath(ext)
		ret = append(ret, path)
		f, err := os.Create(u.dir + path)
		if err != nil {
			return nil, err
		}

		if _, err = io.Copy(f, file); err != nil {
			return nil, err
		}

		// 水印
		if u.watermark != nil && u.watermark.isAllowExt(ext) {
			if err = u.watermark.Mark(f, ext); err != nil {
				return nil, err
			}
		}

		// 循环最后关闭所有打开的文件
		f.Close()
		file.Close()
	}

	return ret, nil
}
