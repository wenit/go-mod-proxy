package api

import (
	"net/http"
)

// Repository 目录
var Repository string

// StartAPIServer 启动API服务
func StartAPIServer(bindAddress string) error {

	http.HandleFunc("/", index)
	http.HandleFunc("/upload", upload)

	err := http.ListenAndServe(bindAddress, nil)

	return err
}
