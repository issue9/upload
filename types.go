// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package upload

import (
	"os"
)

// io.SectionReader.Size() int64
type sizer interface {
	Size() int64
}

// os.File.Stat()(os.FileInof,error)
type stater interface {
	Stat() (os.FileInfo, error)
}
