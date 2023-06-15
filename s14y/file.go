package s14y

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/dianpeng/hi-doctor/run"

	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

type fileSd struct {
	dir string
}

func (f *fileSd) readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *fileSd) md5(data string) string {
	hash := md5.Sum([]byte(data))
	val := hex.EncodeToString(hash[:])
	return val
}

func (f *fileSd) Refresh(currentJ []Job) []NewJob {
	dirList, err := os.ReadDir(f.dir)
	if err != nil {
		return nil // cannot open
	}

	targetList := make(map[string]NewJob)
	for _, entry := range dirList {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := path.Ext(name)
		if ext == ".yaml" {
			path := filepath.Join(f.dir, name)
			data, err := f.readFile(path)
			if err != nil {
				continue
			}
			name, err := run.GetInspectionName(data)
			if err != nil {
				continue
			}
			targetList[name] = NewJob{
				Name:   name,
				Md5:    f.md5(data),
				Data:   data,
				Origin: fmt.Sprintf("file://%s", path),
				Delete: false,
			}
		}
	}

	currentLookup := make(map[string]Job)
	for _, v := range currentJ {
		currentLookup[v.Name] = v
	}

	out := []NewJob{}
	for _, entry := range targetList {
		if old, ok := currentLookup[entry.Name]; ok {
			if old.Md5 == entry.Md5 {
				continue
			}
		}
		out = append(out, entry)
	}

	for _, existed := range currentJ {
		if _, has := targetList[existed.Name]; !has {
			out = append(out, NewJob{
				Name:   existed.Name,
				Delete: true,
			})
		}
	}

	return out
}

func newDir(dir string) S14y {
	return &fileSd{
		dir: dir,
	}
}
