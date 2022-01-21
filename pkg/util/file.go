package util

import (
	"github.com/sxy/lianfang/pkg/models"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const LinuxPathSeparator byte = '/'

func GenKeepBothFilePath(localfilepath string) string {
	exists, fi := IsExists(localfilepath)
	if exists { // 本地文件已存在
		name := fi.Name()
		lastDotIndex := strings.LastIndex(name, ".")
		if lastDotIndex > 0 || lastDotIndex == len(name)-1 {
			name = name[0:lastDotIndex] + "(2)" + name[lastDotIndex:]
		} else {
			name += "(2)"
		}
		return GenKeepBothFilePath(filepath.Join(filepath.Dir(localfilepath), name))
	} else { // 本地文件不存在
		return localfilepath
	}
}

// 判断指定路径是否已存在文件或目录，存在返回(true, fileinfo fs.FileInfo)，否则返回 (false, nil)
func IsExists(localfilepath string) (bool, fs.FileInfo) {
	fi, err := os.Stat(localfilepath)
	return err == nil || !os.IsNotExist(err), fi
}

// StripPathShortcuts removes any leading or trailing "../" from a given path
func StripPathShortcuts(p string) string {
	newPath := path.Clean(p)
	trimmed := strings.TrimPrefix(newPath, "../")

	for trimmed != newPath {
		newPath = trimmed
		trimmed = strings.TrimPrefix(newPath, "../")
	}

	// trim leftover {".", ".."}
	if newPath == "." || newPath == ".." {
		newPath = ""
	}

	if len(newPath) > 0 && string(newPath[0]) == "/" {
		return newPath[1:]
	}

	return newPath
}

func GetPrefix(file string) string {
	// tar strips the leading '/' if it's there, so we will too
	return strings.TrimLeft(file, "/")
}

//
//  GetLinuxPath
//  @Description: 修正路径为linux 路径分隔符
//  @param path
//  @return string
//
func GetLinuxPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func GenerateHttpHeaderFileName(fileBaseName string) string {
	fn := url.PathEscape(fileBaseName)
	if fileBaseName == fn {
		fn = "filename=" + fn
	} else {
		/**
		  The parameters "filename" and "filename*" differ only in that
		  "filename*" uses the encoding defined in [RFC5987], allowing the use
		  of characters not present in the ISO-8859-1 character set
		  ([ISO-8859-1]).
		*/
		fn = "filename=" + fileBaseName + "; filename*=utf-8''" + fn
	}
	return fn
}

func ArchiveInfo(isDir bool, fileName string) models.ArchiveInfo {
	if !isDir {
		name := fileName
		lowerName := strings.ToLower(name)
		supportedArchiveType := []string{"zip", "jar", "war", "tar.gz", "tar", "tgz", "rar"}
		for _, t := range supportedArchiveType {
			if strings.HasSuffix(lowerName, "."+t) {
				return models.ArchiveInfo{
					IsArchive:   true,
					ArchiveType: t,
					ArchiveName: name[0:strings.LastIndex(lowerName, "."+t)],
				}
			}
		}
	}
	return models.ArchiveInfo{}
}
