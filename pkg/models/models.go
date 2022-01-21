/*
Package models
@Time : 2022/1/21 17:33
@Author : sunxy
@File : models
@description:
*/
package models

type DirResponse struct {
	Files []*File `json:"files"`
}

type FileResponse struct {
	Content string `json:"content"`
}

type ArchiveInfo struct {
	IsArchive   bool   `json:"isArchive"`
	ArchiveType string `json:"archiveType"`
	ArchiveName string `json:"archiveName"`
}
type File struct {
	Key            string      `json:"key"`
	Path           string      `json:"path"`
	Name           string      `json:"name"`
	IsDir          bool        `json:"isDir"`
	Uncompressable bool        `json:"uncompressable"`
	Size           int64       `json:"size"`
	ModTime        int64       `json:"modTime"`
	ArchiveInfo    ArchiveInfo `json:"archiveInfo"`
}

type ContainerPrepareFileResponse struct {
	TaskId string `json:"taskId"`
}
