// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package upload

import (
	"mime/multipart"
	"os"
)

// io.SectionReader.Size() int64
type sizer interface {
	Size() int64
}

// os.File.Stat()(os.FileInfo,error)
type stater interface {
	Stat() (os.FileInfo, error)
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
		return false, ErrUnknownFileSize
	}

	return size > 0 && size <= u.maxSize, nil
}
