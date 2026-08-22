package migrations

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var embedded embed.FS

func FS() fs.FS {
	return embedded
}
