package common

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

var defaultFilter = []string{".svn", ".git", ".vscode"}

// GetGoPath 获取GOPATH路径
func GetGoPath() string {
	goPath := os.Getenv("GOPATH")
	return goPath
}

// GetGoModulePath 获取本地Go mod 存储路径
func GetGoModulePath() string {
	goPath := GetGoPath()
	modulePath := filepath.Join(goPath, "pkg/mod")
	return modulePath
}

// CopyFile 拷贝文件
func CopyFile(srcFile, destFile string) error {
	input, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}

	absDir := filepath.Dir(destFile)

	if !PathExists(absDir) {
		err := MkDirs(absDir)
		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(destFile, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetGoModuleCacheDownloadPath 获取本地Go mod 存储路径
func GetGoModuleCacheDownloadPath() string {
	goPath := GetGoPath()
	modulePath := filepath.Join(goPath, "pkg/mod/cache/download")
	return modulePath
}

// PathExists 判断文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// GetCurrentDirectory 获取程序运行路径
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {

	}
	return strings.Replace(dir, "\\", "/", -1)
}

// MkDirs 创建文件夹
func MkDirs(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	return err
}

// ZipFilter 压缩
func ZipFilter(fileDir string, zipFile string, prefix string, filter []string) error {
	absZipFile, _ := filepath.Abs(zipFile)
	absZipDir := filepath.Dir(zipFile)

	if !PathExists(absZipDir) {
		err := MkDirs(absZipDir)
		if err != nil {
			return err
		}
	}
	outFile, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer outFile.Close()
	w := zip.NewWriter(outFile)
	defer w.Close()

	err = filepath.Walk(fileDir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(fileDir, path)
		if filter != nil && len(filter) > 0 {

			if Filter(rel, filter) {
				return nil
			}
		}

		absFilePath, _ := filepath.Abs(path)

		if absFilePath == absZipFile {
			return nil
		}
		if prefix != "" {
			rel = fmt.Sprintf("%s/%s", prefix, rel)
		}

		compressErr := compress(rel, path, w)
		if compressErr != nil {
			return compressErr
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Filter 过滤
func Filter(rel string, filter []string) bool {
	if filter != nil && len(filter) > 0 {
		for _, v := range filter {
			if strings.HasPrefix(rel, v) {
				return true
			}
		}
	}
	return false
}

// Zip 压缩
func Zip(fileDir string, zipFile string) error {
	return ZipFilter(fileDir, zipFile, "", defaultFilter)
}

func compress(rel string, path string, zw *zip.Writer) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	writer, err := zw.Create(rel)
	if err != nil {
		return err
	}
	io.Copy(writer, file)
	defer file.Close()
	return nil
}

// Unzip 解压
func Unzip(zipFile, dest string) error {
	if !PathExists(dest) {
		err := MkDirs(dest)
		if err != nil {
			return err
		}
	}

	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		absPath := filepath.Join(dest, file.Name)
		os.MkdirAll(filepath.Dir(absPath), 0755)
		w, err := os.Create(absPath)
		if err != nil {
			return err
		}
		io.Copy(w, rc)
		w.Close()
		rc.Close()
	}
	return nil
}

// ZipHash 计算zip文件hash值
func ZipHash(zipFile string) string {
	hash, err := dirhash.HashZip(zipFile, dirhash.Hash1)
	if err != nil {
		log.Println("计算ZipHash失败：", err)
		return ""
	}
	return hash
}
