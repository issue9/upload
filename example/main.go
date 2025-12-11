// SPDX-FileCopyrightText: 2015-2025 caixw
//
// SPDX-License-Identifier: MIT

// 简单的 upload 包的示例程序，可以一次上传一个或是多个功能！
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/issue9/upload/v4"
)

const addr = ":8082"

func h(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		post(w, r)
	}

	get(w)
}

func get(w http.ResponseWriter) {
	html := `
<doctype html>
<html>
	<head><meta charset="utf-8" /><title>upload example</title></head>
	<body>
		<form method="POST" action="" enctype="multipart/form-data">
			<input type="file" multiple="multiple" name="field">
			<button type="submit">submit</button>
		</form>
	</body>
</html>
`
	w.Header().Add("Content-Type", "text/html;charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		log.Println(err)
	}
}

func post(_ http.ResponseWriter, r *http.Request) {
	root, err := os.OpenRoot("./uploads/")
	if err != nil {
		log.Panic(err)
	}

	s, err := upload.NewLocalSaver(root, "", upload.FilenameAI)
	if err != nil {
		log.Panic(err)
	}

	u := upload.New(s, 1024*1024, ".txt", ".gif", ".png")
	if err != nil {
		log.Panic(err)
	}

	files, err := u.Do("field", r)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("本次上传[%v]份文件：\n", len(files))
	for _, file := range files {
		log.Println(file)
	}
}

func main() {
	log.Println("http://localhost" + addr)
	if err := http.ListenAndServe(addr, http.HandlerFunc(h)); err != nil {
		log.Panic(err)
	}
}
