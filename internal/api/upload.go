package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/wenit/go-mod-proxy/pkg/common"
	"github.com/wenit/go-mod-proxy/pkg/mod"
	"golang.org/x/mod/modfile"
)

func upload(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(32 << 20)

	if err != nil {
		wirteError(w, err)
		return
	}

	infoFile, _, err := r.FormFile("infoFile")
	if err != nil {
		wirteError(w, err)
		return
	}

	infoData, err := ioutil.ReadAll(infoFile)
	if err != nil {
		wirteError(w, err)
		return
	}

	infoFileObj := &mod.InfoFile{}

	json.Unmarshal(infoData, infoFileObj)

	modFile, _, err := r.FormFile("modFile")
	if err != nil {
		wirteError(w, err)
		return
	}

	tempModFile, modData := saveModFile(modFile)

	defer os.Remove(tempModFile)

	modFileObj, err := getModuleFile(tempModFile, infoFileObj.Version)
	if err != nil {
		wirteError(w, err)
		return
	}

	log.Printf("开始上传MOD：[%s/%s]", modFileObj.Module.Mod.Path, modFileObj.Module.Mod.Version)

	modRootDir := filepath.Join(Repository, modFileObj.Module.Mod.Path, "@v")

	if !common.PathExists(modRootDir) {
		err := common.MkDirs(modRootDir)
		if err != nil {
			wirteError(w, err)
			return
		}
	}

	// 1. write info file
	dstInfoFilePath := filepath.Join(modRootDir, infoFileObj.Version+".info")
	err = wirteFile(infoData, dstInfoFilePath)
	if err != nil {
		wirteError(w, err)
		return
	}
	// 2. write mod file
	dstModFilePath := filepath.Join(modRootDir, infoFileObj.Version+".mod")
	err = wirteFile(modData, dstModFilePath)
	if err != nil {
		wirteError(w, err)
		return
	}
	// 3. write zip file
	zipFile, _, err := r.FormFile("zipFile")
	if err != nil {
		wirteError(w, err)
		return
	}
	dstZipFilePath := filepath.Join(modRootDir, infoFileObj.Version+".zip")
	err = copyFile(zipFile, dstZipFilePath)
	if err != nil {
		wirteError(w, err)
		return
	}
	log.Printf("上传MOD成功：[%s/%s]", modFileObj.Module.Mod.Path, modFileObj.Module.Mod.Version)
	fmt.Fprintln(w, "upload ok!")
}

func wirteError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	fmt.Fprintln(w, err)
}

func saveModFile(file io.Reader) (string, []byte) {
	appDir := common.GetCurrentDirectory()
	tempDir := filepath.Join(appDir, "temp")

	if !common.PathExists(tempDir) {
		common.MkDirs(tempDir)
	}

	data, _ := ioutil.ReadAll(file)

	uuidStr := uuid.New().String()

	absPath := filepath.Join(tempDir, uuidStr)

	ioutil.WriteFile(absPath, data, 0644)

	return absPath, data
}

func getModuleFile(path string, version string) (*modfile.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open module file: %w", err)
	}
	defer file.Close()

	moduleBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read module file: %w", err)
	}

	moduleFile, err := modfile.Parse(path, moduleBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("parse module file: %w", err)
	}

	if moduleFile.Module == nil {
		return nil, fmt.Errorf("parsing module returned nil module")
	}

	moduleFile.Module.Mod.Version = version
	return moduleFile, nil
}

func wirteFile(data []byte, dst string) error {
	return ioutil.WriteFile(dst, data, 0644)
}

func copyFile(file io.Reader, dst string) error {

	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, file)
	return nil
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(tpl))
}

const tpl = `<html>
<head>
<title>上传文件</title>
</head>
<style>
span {
	width: 100px;
	display: inline-block;
}
</style>
<body>
<form enctype="multipart/form-data" action="/upload" method="POST">
 <span>info文件:</span><input type="file" name="infoFile" ></input><br/>
 <span>mod文件:</span><input type="file" name="modFile" /><br/>
 <span>zip文件:</span><input type="file" name="zipFile" /><br/>
 <!--<span>ziphash文件:</span><input type="file" name="ziphashFile" /><br/><br/>-->
 <input type="submit" value="上传" />
</form>
</body>
</html>`
