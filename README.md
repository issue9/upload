upload
[![Go](https://github.com/issue9/upload/workflows/Go/badge.svg)](https://github.com/issue9/upload/actions?query=workflow%3AGo)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/upload/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/upload)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/upload/v3)](https://pkg.go.dev/github.com/issue9/upload/v3)
[![Go version](https://img.shields.io/github/go-mod/go-version/issue9/upload)](https://golang.org)
======

处理上传文件，若是图片还可以设置水印。

```go
s, err := NewLocalSaver(os.OpenRoot("~/uploads/"), "2006/01/02/", nil)
u := upload.New(s, 1024*1024*10, ".txt", ".jpg", ".png")
u.SetWatermark(...) // 设置水印图片

func(w http.ResponseWriter, r *http.Request) {
    if r.Method="POST"{
        u.Do("files", r) // 执行上传操作
    }
}
```

安装
----

```shell
go get github.com/issue9/upload/v3
```

版权
----

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
