// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package upload

import (
	"testing"

	"github.com/issue9/assert"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	u, err := New("./testdir", 10*1024, "2006/05/04/", "gif", ".png", ".GIF")
	a.NotError(err).NotNil(u)
	// 自动转换成小写，且加上最前面的.符号
	a.Equal(u.exts, []string{".gif", ".png", ".gif"})
	a.Equal(u.dir, "./testdir/")

	// dir为一个文件
	u, err = New("./testdir/file", 10*1024, "2006/05/04/", "gif", ".png", ".GIF")
	a.Error(err).Nil(u)
}

func TestUpload_isAllowExt(t *testing.T) {
	a := assert.New(t)

	u, err := New("./testdir", 10*1024, "2006/05/04/", "gif", ".png", ".GIF")
	a.NotError(err).NotNil(u)
	a.True(u.isAllowExt(".gif"))
	a.True(u.isAllowExt(".GIF"))
	a.True(u.isAllowExt(".png"))

	a.False(u.isAllowExt(".TXT"))
	a.False(u.isAllowExt(".exe"))
}
