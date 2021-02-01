package packago

import (
	"encoding/base64"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/legenove/utils"
)

var ignore = map[string]bool{
	".git":   true,
	"vendor": true,
}

const BackQuota = "$BACK_QUOTA$"

func SetIgnore(names ...string) {
	for _, n := range names {
		ignore[n] = true
	}
}

type FileInfo struct {
	FullPath string
	SubPath  string
	SubDir   string
	Name     string
}

func GetDirAllFile(dirpath string) ([]FileInfo, error) {
	var dir_list []FileInfo
	dir_err := filepath.Walk(dirpath,
		func(path string, f os.FileInfo, err error) error {
			basePath := strings.Replace(path, dirpath, "", -1)
			if strings.HasPrefix(basePath, "/") {
				basePath = basePath[1:]
			}
			if f == nil {
				return err
			}
			for ig := range ignore {
				if strings.HasPrefix(basePath, ig) {
					return nil
				}
			}
			if f.IsDir() {
				return nil
			}
			subDir := basePath[:len(basePath)-len(f.Name())]
			dir_list = append(dir_list, FileInfo{
				FullPath: path,
				SubPath:  basePath,
				Name:     f.Name(),
				SubDir:   subDir,
			})

			return nil
		})
	return dir_list, dir_err
}

func PackagerAllFile(dirpath string, out string, goPkg string) error {
	var outPath = out
	if strings.HasPrefix(outPath, "./") {
		outPath = path.Join(dirpath, outPath)
	}
	files, _ := GetDirAllFile(dirpath)
	var err error
	pkgList := map[string]bool{}
	for _, f := range files {
		goPkgName := path.Join(goPkg, out, "tamplates", f.SubDir)
		if strings.HasSuffix(goPkgName, "/") {
			goPkgName = goPkgName[:len(goPkgName)-1]
		}
		err = utils.CreateDir(path.Join(outPath, "tamplates", f.SubDir))
		if err != nil {
			return err
		}
		var pkgName = "tpl"
		if len(f.SubDir) > 0 {
			subDir := strings.Split(f.SubDir, "/")
			for _, d := range subDir {
				d = strings.Replace(d, ".", "", -1)
				if len(d) == 0 {
					continue
				}
				pkgName += strings.Title(d)
			}
		}
		outFileName := strings.Join([]string{f.Name, "tpl", "go"}, ".")
		outFileFullPath := path.Join(outPath, "tamplates", f.SubDir, outFileName)

		// read from file
		data, err := utils.LoadDataFrom(f.FullPath)
		if err != nil {
			return err
		}
		var dataStr string
		var isBase64 bool
		if IsText(data) {
			dataStr = strings.Replace(string(data), "`", BackQuota, -1)
			isBase64 = false
		} else {
			dataStr = base64.StdEncoding.EncodeToString(data)
			isBase64 = true
		}
		err = utils.WriteDataTo(outFileFullPath, utils.FormatTplByMap(tlpFileTlp, map[string]interface{}{
			"package":     pkgName,
			"fileVarName": GetVarName(f.Name),
			"fileSubPath": f.SubPath,
			"strData":     dataStr,
			"base64":      isBase64,
		}))
		if err != nil {
			return err
		}
		pkgList[goPkgName] = true
	}
	err = utils.WriteDataTo(path.Join(outPath, "unpackage.go"), utils.FormatTplByMap(tlpGenTlp, map[string]interface{}{
		"pkgList": pkgList,
	}))

	return nil
}

var tlpFileTlp = `package {{.package}}

import (
	"github.com/legenove/packago"
)

func init() {
	packago.FileList[{{.fileVarName}}FileName] = &packago.FileGenInfo{
		Content : {{.fileVarName}}Tlp,
		Overide : false,
		Example : false,
		Base64  : {{.base64}},
	}
}

var {{.fileVarName}}FileName = "{{.fileSubPath}}"
var {{.fileVarName}}Tlp = ` + "`" + `{{.strData}}` + "`"

var tlpGenTlp = `package packagoGen

import (
	{{ range $key, $value := .pkgList -}}
	_ "{{$key}}"
	{{ end }}
	"github.com/legenove/packago"
)

func Unpackage(outPath string, kv map[string]interface{}, example, first bool) error {
	return packago.Unpackage(outPath, kv, example, first)
}
`

func IsText(buff []byte) bool {
	filetype := http.DetectContentType(buff)
	if strings.HasPrefix(filetype, "text/") {
		return true
	}
	return false
}
