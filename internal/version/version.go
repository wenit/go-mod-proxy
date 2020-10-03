package version

import (
	"encoding/json"
	"fmt"
	"runtime"
)

var (
	version   string = "1.0.0"
	buildDate string
	author    string
)

// Version 版本信息
type Version struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
	Author    string `json:"author"`
}

func (o Version) String() string {
	s, _ := json.MarshalIndent(o, "", "    ")
	//s, _ := json.Marshal(o)
	return string(s)
}

// GetVersion 获取版本信息
func GetVersion() Version {
	if len(author) <= 0 {
		author = "zwb"
	}
	return Version{
		Version:   version,
		BuildDate: buildDate,
		Author:    author,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
