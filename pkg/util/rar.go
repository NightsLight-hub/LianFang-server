package util

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// rar压缩
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

func rarCompress(file *os.File, prefix string, dest *os.File, progressReceiver ...io.Writer) error {
	return nil
}

// 解压 rar
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
	dir := Pwd()
	rarC := filepath.Join(dir, "tools", "rar-compress", "rar-c")

	output, err := exec.Command(rarC, append([]string{rarFile}, files...)...).CombinedOutput()

	if err != nil {
		return err
	}
	out := string(output)
	logrus.Debug(out)
	return nil
}

func rarX(rarFile, dest string, progressReceiver ...io.Writer) error {
	dir := Pwd()
	rarX := filepath.Join(dir, "tools", "rar-compress", "rar-x")
	if !strings.HasSuffix(dest, string(filepath.Separator)) {
		dest += string(filepath.Separator)
	}
	logrus.Debug(rarX, rarFile, dest)
	output, err := exec.Command(rarX, rarFile, dest).CombinedOutput()
	out := string(output)
	logrus.Info(out)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
