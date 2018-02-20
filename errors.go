// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package upload

import "errors"

// 常用错误类型
var (
	ErrNotAllowExt     = errors.New("不允许的文件上传类型")
	ErrNotAllowSize    = errors.New("文件上传大小超过最大设定值或是文件大小为0")
	ErrUnknownFileSize = errors.New("未知的文件大小")
)
