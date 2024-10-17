// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package upload

import (
	"bytes"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
)

var _ fs.FS = &Upload{}

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	u, err := New("./testdir", "2006/01/02/", 10*1024, Filename, "gif", ".png", ".GIF")
	a.NotError(err).NotNil(u)
	// 自动转换成小写，且加上最前面的.符号
	a.Equal(u.exts, []string{".gif", ".png", ".gif"})
	a.Equal(u.dir, "./testdir"+string(os.PathSeparator))

	// dir为一个文件
	u, err = New("./testdir/file", "2006/01/02/", 10*1024, Filename, "gif", ".png", ".GIF")
	a.Error(err).Nil(u)
}

func TestUpload_isAllowExt(t *testing.T) {
	a := assert.New(t, false)
	u, err := New("./testdir", "2006/01/02/", 10*1024, Filename, "gif", ".png", ".GIF")
	a.NotError(err).NotNil(u)

	a.True(u.isAllowExt(".gif"))
	a.True(u.isAllowExt(".png"))

	a.False(u.isAllowExt(".TXT"))
	a.False(u.isAllowExt(""))
	a.False(u.isAllowExt("png"))
	a.False(u.isAllowExt(".exe"))
}

func TestUpload_Do(t *testing.T) {
	a := assert.New(t, false)
	u, err := New("./testdir", Day, 10*1024, Filename, "xml")
	a.NotError(err).NotNil(u)
	filename := "./testdir/file.xml"

	f, err := os.Open(filename)
	a.NotError(err).NotNil(f)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", filename)
	a.NotError(err).NotNil(fw)

	_, err = io.Copy(fw, f)
	a.NotError(err)

	err = writer.WriteField("filename", filename)
	a.NotError(err)

	err = writer.Close() // close writer before POST request
	a.NotError(err)

	r, err := http.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Add("content-type", writer.FormDataContentType())
	a.NotError(err).NotNil(r)

	paths, err := u.Do("file", r)
	a.NotError(err).
		Length(paths, 1).
		Equal(paths[0], path.Join(time.Now().Format(Day), "file.xml"))
}

func TestFilename(t *testing.T) {
	a := assert.New(t, false)

	f := Filename("./testdir/", "abc", "")
	a.Equal(f, "./testdir/abc")

	f = Filename("./testdir/", "file.xml", ".xml")
	a.Equal(f, "./testdir/file_1.xml")
}
