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
)

type Upload struct {
	dir     string   // 上传文件保存的路径根目录
	maxSize int64    // 允许的最大文件大小，以byte为单位
	role    string   // 文件命名方式
	exts    []string // 允许的扩展名
}

func New(dir string, maxSize int64, role string, exts ...string) *Upload {
	es := make([]string, 0, len(exts))
	for _, ext := range exts {
		if ext[0] != '.' {
			es = append(es, "."+ext)
			continue
		}
		es = append(es, ext)
	}

	return &Upload{
		dir:     dir,
		maxSize: maxSize,
		role:    role,
		exts:    es,
	}
}

// 判断扩展名是否符合要求。
func (u *Upload) checkExt(ext string) bool {
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
func (u *Upload) checkSize(file multipart.File) (bool, error) {
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

	return size <= u.maxSize, nil
}

func (u *Upload) Do(field string, w *http.ResponseWriter, r *http.Request) ([]string, error) {
	r.ParseMultipartForm(32 << 20)
	heads := r.MultipartForm.File[field]
	ret := make([]string, len(heads))

	for _, head := range heads {
		file, err := head.Open()
		if err != nil {
			return nil, err
		}

		ext := filepath.Ext(head.Filename)
		if !u.checkExt(ext) {
			return nil, errors.New("包含无效的文件类型")
		}

		// TODO 两种判断类型，分开
		if ok, err := u.checkSize(file); !ok || err != nil {
			return nil, errors.New("超过最大的文件大小")
		}

		filename := u.dir + "1" + ext // TODO
		ret = append(ret, filename)
		f, err := os.Create(filename)
		if err != nil {
			return nil, err
		}

		io.Copy(f, file)

		f.Close()
		file.Close() // for的最后关闭file
	}

	return ret, nil
}
