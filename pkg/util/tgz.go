package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//压缩 使用gzip压缩成tar.gz
func TgzCompress(files []*os.File, dest string, progressReceiver ...io.Writer) error {
	d, _ := os.Create(dest)
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, file := range files {
		err := tgzCompress(file, "", tw, progressReceiver...)
		if err != nil {
			return err
		}
	}
	return nil
}

//压缩文件或文件夹，输出到指定的 writer
//files 文件数组，可以是不同dir下的文件或者文件夹
//writer 压缩文件输出目的地
func TgzCompressToWriter(files []*os.File, writer io.Writer, progressReceiver ...io.Writer) error {
	gw := gzip.NewWriter(writer)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, file := range files {
		err := tgzCompress(file, "", tw, progressReceiver...)
		if err != nil {
			return err
		}
	}
	return nil
}

func tgzCompress(file *os.File, prefix string, tw *tar.Writer, progressReceiver ...io.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	if prefix != "" {
		prefix += "/"
	}
	if info.IsDir() {
		header.Name = prefix + info.Name() + "/"
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		prefix = prefix + info.Name()
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = tgzCompress(f, prefix, tw, progressReceiver...)
			if err != nil {
				return err
			}
		}
	} else {
		header.Name = prefix + header.Name
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

//解压 tar.gz
func TgzDecompress(tarFile, dest string, progressReceiver ...io.Writer) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		filename := filepath.Join(dest, hdr.Name)
		// TODO 反馈解压进度
		fmt.Println(hdr.Name + "->" + filename)
		if hdr.FileInfo().IsDir() {
			err = os.MkdirAll(filename, 0755)
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(filename), 0755)
			if err != nil {
				return err
			}
			file, err := os.Create(filename)
			if err != nil {
				return err
			}
			defer file.Close()
			io.Copy(file, tr)
		}
		os.Chmod(filename, hdr.FileInfo().Mode())
		os.Chtimes(filename, hdr.AccessTime, hdr.ModTime)
	}
	return nil
}
