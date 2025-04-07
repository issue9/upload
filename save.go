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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 创建文件的默认权限，比如 [NewLocalSaver] 的参数 dir 若不存在，会使用此权限创建目录。
const presetMode = fs.ModePerm

// Saver 定义了用于保存上传内容的接口
type Saver interface {
	fs.FS

	// Save 保存用户上传的文件
	//
	// filename 为用户上传时的文件名，不包含扩展名部分；
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

// 为 [New] 的参数 format 所允许的几种取值
const (
	None  = ""
	Year  = "2006/"
	Month = "2006/01/"
	Day   = "2006/01/02/"
)

type localSaver struct {
	fs        fs.FS
	dir       string // TODO: 采用 [os.Root]，需要 os.Root 实现 MkdirAll。
	baseURL   string
	format    string
	filenames func(dir, filename, ext string) string
	moveMux   sync.Mutex
}

// NewLocalSaver 实现了一个基于本地文件系统的 [Saver] 接口
//
// dir 上传文件的保存目录，若目录不存在，则会尝试创建；
//
// baseURL 为上传的文件生成访问地址的前缀；
//
// format 子目录的格式，只能是时间格式，取值只能是 [None]、[Year]、[Month] 和 [Day]；
//
// f 设置文件名的生成方式，要求文件在同一目录下具有唯一性，其类型如下：
//
//	func(dir, filename, ext string) string
//
// dir 为文件夹名称，以 / 结尾，filename 为用户上传的文件名，ext 为 filename 中的扩展名部分，
// 返回值是 dir + filename 的路径，实现者可能要调整 filename 的值，以保证在 dir 下唯一。
// 如果为空，则会采用 [Filename] 作为默认值；
func NewLocalSaver(dir, baseURL, format string, f func(dir, filename, ext string) string) (Deleter, error) {
	// 确保 dir 最后一个字符为目录分隔符。
	last := dir[len(dir)-1]
	if last != '/' && last != filepath.Separator {
		dir += string(filepath.Separator)
	}

	// 确保 dir 最后一个字符为目录分隔符。
	last = dir[len(dir)-1]
	if last != '/' && last != filepath.Separator {
		dir += string(filepath.Separator)
	}

	if format != Year && format != Month && format != Day && format != None {
		panic("无效的参数 format")
	}

	// 若不存在目录，则尝试创建
	if err := os.MkdirAll(dir, presetMode); err != nil {
		return nil, err
	}

	// 确保 dir 目录存在。
	// NOTE:此处的 dir 最后个字符为/，所以不用判断是否为目录。
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}

	if f == nil {
		f = Filename
	}

	if baseURL != "" && baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	return &localSaver{
		fs:        os.DirFS(dir),
		dir:       dir,
		baseURL:   baseURL,
		format:    format,
		filenames: f,
	}, nil
}

func (s *localSaver) Open(name string) (fs.File, error) { return s.fs.Open(name) }

func (s *localSaver) Save(f multipart.File, filename string, ext string) (string, error) {
	var relDir string
	if s.format != None {
		relDir = time.Now().Format(s.format)
	}
	dir := s.dir + relDir

	if err := os.MkdirAll(dir, presetMode); err != nil { // 若路径不存在，则创建
		return "", err
	}

	p, destFile, err := s.createFile(dir, filename, ext)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, f); err != nil {
		return "", err
	}

	return s.baseURL + path.Join(relDir, filepath.Base(p)), nil
}

func (s *localSaver) Delete(filename string) error {
	filename = strings.TrimPrefix(filename, s.baseURL)
	return os.Remove(filepath.Join(s.dir, filename))
}

// 主要是为了缩小 moveMux 的范围，只要保证在创建文件时是有效的就行。
func (s *localSaver) createFile(dir, filename, ext string) (string, *os.File, error) {
	s.moveMux.Lock()
	defer s.moveMux.Unlock()

	p := s.filenames(dir, filename, ext)
	destFile, err := os.Create(p)
	if err != nil {
		return "", nil, err
	}
	return p, destFile, nil
}

// Filename 在 dir 下为 s 生成唯一文件名
func Filename(dir, s, ext string) string {
	base := strings.TrimSuffix(s, ext)

	count := 1
	p := dir + s

RET:
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		return p
	}

	p = dir + base + "_" + strconv.Itoa(count) + ext
	count++
	goto RET
}
