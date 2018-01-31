// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package upload 提供文件上传的功能。
//
// 处理上传文件，若是图片还可以设置水印。
//  func(w http.ResponseWriter, r *http.Request) {
//     u, err := upload.New("~/uploads/", 1024*1024*10, ".txt", ".jpg", ".png")
//     u.SetWatermark(...) // 设置水印图片
//
//     if r.Method="POST"{
//         u.Do("files", r) // 执行上传操作
//     }
//  }
package upload
