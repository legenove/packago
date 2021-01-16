package packago

import (
	"encoding/base64"
	"fmt"
	"path"
	"strings"

	"github.com/legenove/utils"
)

var FileList = map[string]*FileGenInfo{}

type FileGenInfo struct {
	Content string
	Overide bool
	Example bool
	Base64  bool
	HasKV   bool
}

func Unpackage(out string, kv map[string]interface{}, first bool) error {
	if kv == nil {
		kv = map[string]interface{}{}
	}
	for f, gi := range FileList {
		outPath := path.Join(out, f)
		if !gi.Overide && utils.FileExists(outPath) {
			continue
		}
		lastPath, _ := path.Split(outPath)
		utils.CreateDir(lastPath)
		if !gi.Example && !first {
			continue
		}
		content := gi.Content
		if gi.Base64 {
			b, err := base64.StdEncoding.DecodeString(content)
			if err != nil {
				return err
			}
			content = string(b)
		} else {
			content = strings.Replace(content, BackQuota, "`", -1)
		}
		if gi.HasKV {
			content = utils.FormatTplByMap(content, kv)
		}
		fmt.Println(outPath)
		utils.WriteDataTo(outPath, content)
	}
	return nil
}
