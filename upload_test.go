// SPDX-FileCopyrightText: 2015-2025 caixw
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
	s, err := NewLocalSaver("./testdir", "", Day, Filename)
	a.NotError(err).NotNil(s)

	u := New(s, 10*1024, "gif", ".png", ".GIF")
	a.NotNil(u)
	// 自动转换成小写，且加上最前面的.符号
	a.Equal(u.exts, []string{".gif", ".png", ".gif"})
	a.Equal(s.(*localSaver).dir, "./testdir"+string(os.PathSeparator))

	// dir 为一个文件
	s, err = NewLocalSaver("./testdir/file", "", Day, Filename)
	a.Error(err).Nil(s)
}

func TestUpload_isAllowExt(t *testing.T) {
	a := assert.New(t, false)
	s, err := NewLocalSaver("./testdir", "", Day, Filename)
	a.NotError(err).NotNil(s)

	u := New(s, 10*1024, "gif", ".png", ".GIF")
	a.NotError(err)

	a.True(u.isAllowExt(".gif"))
	a.True(u.isAllowExt(".png"))

	a.False(u.isAllowExt(".TXT"))
	a.False(u.isAllowExt(""))
	a.False(u.isAllowExt("png"))
	a.False(u.isAllowExt(".exe"))
}

func TestUpload_Do(t *testing.T) {
	a := assert.New(t, false)
	s, err := NewLocalSaver("./testdir", "https://example.com", Day, Filename)
	a.NotError(err).NotNil(s)

	u := New(s, 10*1024, "xml")
	a.NotError(err)

	body, ct := formData(a, "./testdir/file.xml")

	r, err := http.NewRequest(http.MethodPost, "/upload", body)
	a.NotError(err).NotNil(r)
	r.Header.Add("content-type", ct)

	paths, err := u.Do("file", r)
	a.NotError(err).
		Length(paths, 1).
		Equal(paths[0], "https://example.com/"+path.Join(time.Now().Format(Day), "file.xml"))

	a.NotError(s.Delete(paths[0]))
}

func TestUpload_Do_None(t *testing.T) {
	a := assert.New(t, false)
	s, err := NewLocalSaver("./testdir", "https://example.com", None, Filename)
	a.NotError(err).NotNil(s)

	u := New(s, 10*1024, "xml")
	a.NotError(err)

	body, ct := formData(a, "./testdir/file.xml")

	r, err := http.NewRequest(http.MethodPost, "/upload", body)
	a.NotError(err).NotNil(r)
	r.Header.Add("content-type", ct)

	paths, err := u.Do("file", r)
	a.NotError(err).
		Length(paths, 1).
		Equal(paths[0], "https://example.com/"+"file_1.xml") // 已经有 file.xml

	a.NotError(s.Delete(paths[0]))
}

func formData(a *assert.Assertion, filename string) (*bytes.Buffer, string) {
	f, err := os.Open(filename)
	a.NotError(err).NotNil(f)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", filename)
	a.NotError(err).NotNil(fw)

	_, err = io.Copy(fw, f)
	a.NotError(err)

	ct := writer.FormDataContentType()
	err = writer.Close() // close writer before POST request
	a.NotError(err)

	return body, ct
}
