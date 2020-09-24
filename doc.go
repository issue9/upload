// SPDX-License-Identifier: MIT

// Package upload 提供文件上传的功能
//
// 处理上传文件，若是图片还可以设置水印。
//  func(w http.ResponseWriter, r *http.Request) {
//     u, err := upload.New("~/uploads/", 1024*1024*10, ".txt", ".jpg", ".png")
//     u.SetWatermarkFile(...) // 可根据需要设置水印图片
//
//     if r.Method="POST"{
//         u.Do("files", r) // 执行上传操作
//     }
//  }
package upload
