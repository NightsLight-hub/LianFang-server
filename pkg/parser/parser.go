/*
@Time : 2021/12/29 13:35
@Author : sunxy
@File : lsCmdParser
@description:
*/
package parser

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sxy/lianfang/pkg/models"
	"github.com/sxy/lianfang/pkg/util"
	"strings"
)

func ParseStatCmdOutput(out, parentPath string, dirOnly bool) (*models.DirResponse, error) {
	scanner := bufio.NewScanner(strings.NewReader(out))
	dirs := make([]*models.File, 0)
	files := make([]*models.File, 0)
	for scanner.Scan() {
		line := scanner.Text()
		fInfo, err := ParseStatCmdOutputSingleLineToFile(line, parentPath)
		if err != nil {
			return nil, errors.Wrapf(err, "parse stat cmd single line failed, line is %s", line)
		}
		if dirOnly && !fInfo.IsDir {
			continue
		}
		// skip . and ..
		if fInfo.Name == "." || fInfo.Name == ".." {
			continue
		}
		if fInfo.IsDir {
			dirs = append(dirs, fInfo)
		} else {
			files = append(files, fInfo)
		}
	}
	// 文件夹在前，文件在后
	return &models.DirResponse{
		Files: append(dirs, files...),
	}, nil
}

// ParseStatCmdOutputSingleLineToFile 返回stat命令的单行输出表示的file信息
func ParseStatCmdOutputSingleLineToFile(line, parentPath string) (*models.File, error) {
	/**
	  drwxr-xr-x 51 1640941246 .
	  drwxr-xr-x 51 1640941246 ..
	  -rwxr-xr-x 0 1640749283 .dockerenv
	  -rw-r--r-- 12090 1640749272 anaconda-post.log
	  dr-xr-xr-x 12288 1640749277 bin
	  drwxr-xr-x 380 1640749283 dev
	  drwxr-xr-x 66 1640749283 etc
	  drwxr-xr-x 6 1640749277 home
	  dr-xr-xr-x 265 1640749277 lib
	  dr-xr-xr-x 12288 1640749277 lib64
	  drwxr-xr-x 6 1640749277 media
	  drwxr-xr-x 6 1640749277 mnt
	  drwxr-xr-x 6 1640749277 opt
	  dr-xr-xr-x 0 1640749283 proc
	  dr-xr-x--- 27 1640941246 root
	  drwxr-xr-x 33 1640749458 run
	  dr-xr-xr-x 4096 1640749277 sbin
	  drwxr-xr-x 6 1640749277 srv
	  dr-xr-xr-x 0 1639556927 sys
	  drwxrwxrwt 117 1640749286 tmp
	  drwxr-xr-x 18 1640749282 usr
	  drwxr-xr-x 17 1640749458 var
	*/
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	accessRight := parts[0]
	isDir := false
	if strings.HasPrefix(accessRight, "d") {
		isDir = true
	}
	modTime, err := util.StrToInt64(parts[2])
	if err != nil {
		return nil, errors.Wrapf(err, "parseLsCmd modTime failed, output is %s", line)
	}
	size, err := util.StrToInt64(parts[1])
	if err != nil {
		return nil, errors.Wrapf(err, "parseLsCmd size failed, output is %s", line)
	}
	strs := strings.SplitN(line, parts[2], 2)
	if len(strs) != 2 {
		return nil, errors.Wrapf(err, "parseLsCmd FileName failed, output is %s", line)
	}
	fileName := strings.TrimSpace(strs[1])
	var filePath string
	if strings.HasPrefix(fileName, "/") {
		filePath = fileName
	} else {
		// 之所以没用filepath join， 是因为在windows下调试，filepath.join 这里会拼接反斜杠
		if strings.HasSuffix(parentPath, "/") {
			filePath = fmt.Sprintf("%s%s", parentPath, fileName)
		} else {
			filePath = fmt.Sprintf("%s/%s", parentPath, fileName)
		}
	}

	f := &models.File{
		Key:            filePath,
		Path:           filePath,
		Name:           fileName,
		IsDir:          isDir,
		Uncompressable: false,
		Size:           size,
		ModTime:        modTime,
		ArchiveInfo:    util.ArchiveInfo(isDir, fileName),
	}
	return f, nil
}

// IsLsOutDir 判断输出是文件夹或文件  文件夹  true   文件  false
func IsLsOutDir(out string) bool {
	scanner := bufio.NewScanner(strings.NewReader(out))
	if !scanner.Scan() {
		return false
	}
	line := scanner.Text()
	if strings.HasPrefix(line, "d") {
		return true
	} else {
		return false
	}
}
