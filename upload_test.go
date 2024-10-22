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
		Equal(paths[0], "https://example.com/"+path.Join(time.Now().Format(Day), "file.xml"))
}
