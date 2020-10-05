package cacher

import (
	"context"
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/goproxy/goproxy"
)

// Disk implements the `goproxy.Cacher` by using the disk.
type Disk struct {
	// Root is the root of the caches.
	Root string `mapstructure:"root"`
}

// NewHash implements the `goproxy.Cacher`.
func (d *Disk) NewHash() hash.Hash {
	return md5.New()
}

// Cache implements the `goproxy.Cacher`.
func (d *Disk) Cache(ctx context.Context, name string) (goproxy.Cache, error) {
	log.Printf("get module info [%s]", name)
	filename := filepath.Join(d.Root, filepath.FromSlash(name))
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, goproxy.ErrCacheNotFound
		}

		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// fileMIMEType, err := mime.TypeByExtension(ext)
	fileMIMEType, err := ioutil.ReadFile(fmt.Sprint(filename, ".mime-type"))
	if err != nil {
		log.Println("get fileMIMEType err", err)
		fileMIMEType = []byte(getContentType(filename))
	}

	fileChecksum, err := ioutil.ReadFile(fmt.Sprint(filename, ".checksum"))
	if err != nil {
		log.Println("get fileChecksum err", err)
		fileHash := d.NewHash()
		if _, err := io.Copy(fileHash, file); err != nil {
			return nil, err
		}
		fileChecksum = fileHash.Sum(nil)
	}

	return &diskCache{
		file:     file,
		name:     name,
		mimeType: string(fileMIMEType),
		size:     fileInfo.Size(),
		modTime:  fileInfo.ModTime(),
		checksum: fileChecksum,
	}, nil
}

// SetCache implements the `goproxy.Cacher`.
func (d *Disk) SetCache(ctx context.Context, c goproxy.Cache) error {
	filename := filepath.Join(d.Root, filepath.FromSlash(c.Name()))
	if err := os.MkdirAll(
		filepath.Dir(filename),
		0644,
	); err != nil {
		return err
	}

	if err := ioutil.WriteFile(
		fmt.Sprint(filename, ".mime-type"),
		[]byte(c.MIMEType()),
		0644,
	); err != nil {
		return err
	}

	if err := ioutil.WriteFile(
		fmt.Sprint(filename, ".checksum"),
		c.Checksum(),
		0644,
	); err != nil {
		return err
	}

	b, err := ioutil.ReadAll(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, b, 0644)
}

// diskCache implements the `goproxy.Cache`. It is the cache unit of the `Disk`.
type diskCache struct {
	file     *os.File
	name     string
	mimeType string
	size     int64
	modTime  time.Time
	checksum []byte
}

// Read implements the `goproxy.Cache`.
func (dc *diskCache) Read(b []byte) (int, error) {
	return dc.file.Read(b)
}

// Seek implements the `goproxy.Cache`.
func (dc *diskCache) Seek(offset int64, whence int) (int64, error) {
	return dc.file.Seek(offset, whence)
}

// Close implements the `goproxy.Cache`.
func (dc *diskCache) Close() error {
	return dc.file.Close()
}

// Name implements the `goproxy.Cache`.
func (dc *diskCache) Name() string {
	return dc.name
}

// MIMEType implements the `goproxy.Cache`.
func (dc *diskCache) MIMEType() string {
	return dc.mimeType
}

// Size implements the `goproxy.Cache`.
func (dc *diskCache) Size() int64 {
	return dc.size
}

// ModTime implements the `goproxy.Cache`.
func (dc *diskCache) ModTime() time.Time {
	return dc.modTime
}

// Checksum implements the `goproxy.Cache`.
func (dc *diskCache) Checksum() []byte {
	return dc.checksum
}

// GetContentType 获取ContentType
func getContentType(name string) string {
	var mimeType string
	switch ext := strings.ToLower(path.Ext(name)); ext {
	case ".info":
		mimeType = "application/json; charset=utf-8"
	case ".mod":
		mimeType = "text/plain; charset=utf-8"
	case ".zip":
		mimeType = "application/zip"
	default:
		mimeType = mime.TypeByExtension(ext)
	}
	return mimeType
}
