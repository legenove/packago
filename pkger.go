package packago

import (
	"github.com/legenove/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
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
		dataStr := string(data)
		dataStr = strings.Replace(dataStr, "`", BackQuota, -1)
		err = utils.WriteDataTo(outFileFullPath, utils.FormatTplByMap(tlpFileTlp, map[string]interface{}{
			"package":     pkgName,
			"fileVarName": GetVarName(f.Name),
			"fileSubPath": f.SubPath,
			"strData":     dataStr,
		}))
		if err != nil {
			return err
		}
	}
	return nil
}

var tlpFileTlp = `package {{.package}}

import (
	"github.com/legenove/packager"
)

func init() {
	packager.FileList[{{.fileVarName}}FileName] = {{.fileVarName}}Tlp
}

var {{.fileVarName}}FileName = "{{.fileSubPath}}"
var {{.fileVarName}}Tlp = ` + "`" + `{{.strData}}` + "`"
