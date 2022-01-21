package util

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.thunisoft.com/research/cloud/caas/volume-explorer/server/logger"
)

// TODO ierr         发送所有压缩文档到 stderr
//rar压缩
func RarCompress(files []*os.File, dest string, progressReceiver ...io.Writer) error {
	d, _ := os.Create(dest)
	defer d.Close()
	for _, file := range files {
		err := rarCompress(file, "", d, progressReceiver...)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO rar压缩
func rarCompress(file *os.File, prefix string, dest *os.File, progressReceiver ...io.Writer) error {
	// info, err := file.Stat()
	// if err != nil {
	// 	return err
	// }
	// if info.IsDir() {
	// 	prefix = prefix + "/" + info.Name()
	// 	fileInfos, err := file.Readdir(-1)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	for _, fi := range fileInfos {
	// 		// TODO
	// 	}
	// } else {
	// 	header, err := tar.FileInfoHeader(info, "")
	// 	header.Name = prefix + "/" + header.Name
	// 	if err != nil {
	// 		return err
	// 	}
	// 	err = tw.WriteHeader(header)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	_, err = io.Copy(tw, file)
	// 	file.Close()
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

//解压 rar
func RarDecompress(rarFile, dest string, progressReceiver ...io.Writer) error {

	fileInfo, err := os.Stat(rarFile)
	if err != nil {
		return errors.WithStack(err)
	}
	if fileInfo.IsDir() {
		return errors.WithStack(err)
	}
	return rarX(rarFile, dest, progressReceiver...)
}

func rarC(rarFile string, files []string, progressReceiver ...io.Writer) error {
	dir := GetAppDir()
	rarC := filepath.Join(dir, "tools", "rar-compress", "rar-c")

	output, err := exec.Command(rarC, append([]string{rarFile}, files...)...).CombinedOutput()

	if err != nil {
		return err
	}
	out := string(output)
	logger.Info(out)

	return nil
}

func rarX(rarFile, dest string, progressReceiver ...io.Writer) error {
	dir := GetAppDir()
	rarX := filepath.Join(dir, "tools", "rar-compress", "rar-x")
	if !strings.HasSuffix(dest, string(filepath.Separator)) {
		dest += string(filepath.Separator)
	}
	logger.Debug(rarX, rarFile, dest)
	output, err := exec.Command(rarX, rarFile, dest).CombinedOutput()
	out := string(output)
	// TODO 反馈解压进度
	logger.Info(out)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
