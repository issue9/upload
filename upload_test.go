// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package upload

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	u, err := New("./testdir", "2006/01/02/", 10*1024, "gif", ".png", ".GIF")
	a.NotError(err).NotNil(u)
	// 自动转换成小写，且加上最前面的.符号
	a.Equal(u.exts, []string{".gif", ".png", ".gif"})
	a.Equal(u.dir, "./testdir"+string(os.PathSeparator))

	// dir为一个文件
	u, err = New("./testdir/file", "2006/01/02/", 10*1024, "gif", ".png", ".GIF")
	a.Error(err).Nil(u)
}

func TestUpload_isAllowExt(t *testing.T) {
	a := assert.New(t)

	u, err := New("./testdir", "2006/01/02/", 10*1024, "gif", ".png", ".GIF")
	a.NotError(err).NotNil(u)
	a.True(u.isAllowExt(".gif"))
	a.True(u.isAllowExt(".png"))

	a.False(u.isAllowExt(".TXT"))
	a.False(u.isAllowExt(""))
	a.False(u.isAllowExt(".png"))
	a.False(u.isAllowExt(".exe"))
}

func TestUpload_getDestPath(t *testing.T) {
	a := assert.New(t)

	u, err := New("./testdir", "2006/01/02/", 10*1024, "gif", ".png", ".GIF")
	a.NotError(err).NotNil(u)

	t.Log(u.getDestPath(".png"))
	t.Log(u.getDestPath(".jpeg"))
}
