package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//压缩文件或文件夹，输出到压缩文件
//files 文件数组，可以是不同dir下的文件或者文件夹
//dest 压缩文件存放地址
func ZipCompressToFile(files []*os.File, dest string, progressReceiver ...io.Writer) error {
	d, _ := os.Create(dest)
	defer d.Close()
	err := ZipCompressToWriter(files, d)
	if err != nil {
		return err
	}

	return nil
}

//压缩文件或文件夹，输出到指定的 writer
//files 文件数组，可以是不同dir下的文件或者文件夹
//writer 压缩文件输出目的地
func ZipCompressToWriter(files []*os.File, writer io.Writer, progressReceiver ...io.Writer) error {
	archive := zip.NewWriter(writer)
	defer archive.Close()
	for _, file := range files {
		err := zipCompress(file, "", archive, progressReceiver...)
		if err != nil {
			return err
		}
	}
	return nil
}

func zipCompress(file *os.File, prefix string, archive *zip.Writer, progressReceiver ...io.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	header.Method = zip.Deflate
	if err != nil {
		return err
	}
	if prefix != "" {
		prefix += "/"
	}
	if info.IsDir() {
		header.Name = prefix + info.Name() + "/"
		archive.CreateHeader(header)

		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		prefix += info.Name()
		for _, fi := range fileInfos {
			f, err := os.Open(filepath.Join(file.Name(), fi.Name()))
			if f != nil {
				defer f.Close()
			}
			if err != nil {
				return err
			}
			err = zipCompress(f, prefix, archive, progressReceiver...)
			if err != nil {
				return err
			}
		}
	} else {
		header.Name = prefix + header.Name

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		defer file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

//解压
func ZipDecompress(zipFile, dest string, progressReceiver ...io.Writer) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		filename := filepath.Join(dest, file.Name)
		// TODO 反馈解压进度
		fmt.Println(file.Name + "->" + filename)
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(filename, 0755)
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(filename), 0755)
			if err != nil {
				return err
			}
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			w, err := os.Create(filename)
			if err != nil {
				return err
			}
			defer w.Close()
			_, err = io.Copy(w, rc)
			if err != nil {
				return err
			}
		}
		os.Chmod(filename, file.Mode())
		os.Chtimes(filename, file.Modified, file.Modified)
	}
	return nil
}
