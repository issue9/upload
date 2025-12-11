// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package upload

import (
	"errors"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

const presetMode = fs.ModePerm

// Saver 定义了用于保存上传内容的接口
type Saver interface {
	fs.FS

	// Save 保存用户上传的文件
	//
	// filename 为用户上传时的文件名，包含扩展名部分；
	// ext 为 filename 中的扩展名部分；
	// 返回该文件对应的唯一标记。
	Save(file multipart.File, filename string, ext string) (string, error)
}

// Deleter 实现了删除功能的 [Saver]
type Deleter interface {
	Saver

	// Delete 删除文件
	//
	// filename 由 [Saver.Save] 返回的内容。
	Delete(filename string) error
}

type localSaver struct {
	root        *os.Root
	baseURL     string
	filenames   func(fs fs.FS, filename, ext string) string
	filenameMux sync.Mutex
}

// NewLocalSaver 实现了一个基于本地文件系统的 [Saver] 接口
//
// root 上传文件的保存目录；
//
// baseURL 为上传的文件生成访问地址的前缀；
//
// f 设置文件名的生成方式，要求文件在同一目录下具有唯一性，其类型如下：
//
//	func(dir fs.FS, filename, ext string) string
//
// dir 为所属的文件系统，filename 用户上传的文件名，ext 为 filename 中的扩展名部分，
// 返回值是修正后的 filename，实现者需要保证 filename 在 dir 下是唯一的，filename 的路径分隔符必须是 /，不随系统而改变。
// 如果为空，则会采用 [FilenameAI] 作为默认值；
func NewLocalSaver(root *os.Root, baseURL string, f func(dir fs.FS, filename, ext string) string) (Deleter, error) {
	if root == nil {
		panic("无效的参数 root")
	}

	if f == nil {
		f = FilenameAI
	}

	if baseURL != "" && baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	return &localSaver{
		root:      root,
		baseURL:   baseURL,
		filenames: f,
	}, nil
}

func (s *localSaver) Open(name string) (fs.File, error) { return s.root.Open(name) }

func (s *localSaver) Save(f multipart.File, filename, ext string) (string, error) {
	p := s.createFilename(filename, ext)

	dir := path.Dir(p)
	if err := s.root.MkdirAll(dir, presetMode); err != nil {
		return "", err
	}

	destFile, err := s.root.Create(p)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, f); err != nil {
		return "", err
	}

	return s.baseURL + p, nil
}

func (s *localSaver) Delete(filename string) error {
	filename = strings.TrimPrefix(filename, s.baseURL)
	return s.root.Remove(filename)
}

// 主要是为了缩小 filenameMux 的范围。
func (s *localSaver) createFilename(filename, ext string) string {
	s.filenameMux.Lock()
	defer s.filenameMux.Unlock()
	return s.filenames(s.root.FS(), filename, ext)
}

// FilenameAI 在 dir 下为 s 生成唯一文件名
//
// 如果已经存在同名的文件，会在文件后以数字形式自增。
//
// s 包含了扩展名的文件名；
func FilenameAI(dir fs.FS, s, ext string) string {
	base := strings.TrimSuffix(s, ext)

	count := 1
	p := s

RET:
	if _, err := fs.Stat(dir, p); errors.Is(err, os.ErrNotExist) {
		return p
	}

	p = base + "_" + strconv.Itoa(count) + ext
	count++
	goto RET
}
