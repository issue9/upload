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
	root, err := os.OpenRoot("./testdir")
	a.NotError(err).NotNil(root)

	s, err := NewLocalSaver(root, "", Day, Filename)
	a.NotError(err).NotNil(s)

	u := New(s, 10*1024, "gif", ".png", ".GIF")
	a.NotNil(u)
	// 自动转换成小写，且加上最前面的.符号
	a.Equal(u.exts, []string{".gif", ".png", ".gif"}).
		Equal(s.(*localSaver).root, root)
}

func TestUpload_isAllowExt(t *testing.T) {
	a := assert.New(t, false)
	root, err := os.OpenRoot("./testdir")
	a.NotError(err).NotNil(root)

	s, err := NewLocalSaver(root, "", Day, Filename)
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
	root, err := os.OpenRoot("./testdir")
	a.NotError(err).NotNil(root)

	s, err := NewLocalSaver(root, "https://example.com", Day, Filename)
	a.NotError(err).NotNil(s)

	u := New(s, 10*1024, "xml")
	a.NotError(err)

	filename := "./testdir/file.xml"

	f, err := os.Open(filename)
	a.NotError(err).NotNil(f)
	defer f.Close()

	body, ct := formData(a, filename)

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
	root, err := os.OpenRoot("./testdir")
	a.NotError(err).NotNil(root)

	s, err := NewLocalSaver(root, "https://example.com", None, Filename)
	a.NotError(err).NotNil(s)

	u := New(s, 10*1024, "xml")
	a.NotError(err)

	filename := "./testdir/file.xml"

	f, err := os.Open(filename)
	a.NotError(err).NotNil(f)
	defer f.Close()

	body, ct := formData(a, filename)

	r, err := http.NewRequest(http.MethodPost, "/upload", body)
	a.NotError(err).NotNil(r)
	r.Header.Add("content-type", ct)

	paths, err := u.Do("file", r)
	a.NotError(err).
		Length(paths, 1).
		Equal(paths[0], "https://example.com/"+"file_1.xml") // 已经有 file.xml
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
