upload [![Build Status](https://travis-ci.org/issue9/upload.svg?branch=master)](https://travis-ci.org/issue9/upload)
======

处理上传文件，若是图片还可以设置水印。
```go
func(w http.ResponseWriter, r *http.Request){
    u, err := upload.New("~/uploads/", "2006/01/02/", 1024*1024*10, ".txt", ".jpg", ".png")
    u.SetWatermark(...) // 设置水印图片

    if r.Method="POST"{
        u.Do("files", r) // 执行上传操作
    }
}

```


### 安装

```shell
go get github.com/issue9/upload
```


### 文档

[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/issue9/upload)
[![GoDoc](https://godoc.org/github.com/issue9/upload?status.svg)](https://godoc.org/github.com/issue9/upload)


### 版权

本项目采用[MIT](http://opensource.org/licenses/MIT)开源授权许可证，完整的授权说明可在[LICENSE](LICENSE)文件中找到。
