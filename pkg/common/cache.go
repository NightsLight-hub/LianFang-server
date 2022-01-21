/*
Package common
@Time : 2022/1/21 17:50
@Author : sunxy
@File : cache
@description:
*/
package common

import (
	"fmt"
	"github.com/muesli/cache2go"
)

var Cache *cache2go.CacheTable

func SetupCache() {
	Cache = cache2go.Cache("default")
	// cache2go supports a few handy callbacks and loading mechanisms.
	// cache.SetAboutToDeleteItemCallback(func(e *cache2go.CacheItem) {
	// 	fmt.Println("Deleting:", e.Key(), e.Data().(*myStruct).text, e.CreatedOn())
	// })
}

func LsCommandCacheKey(namespace, pod, container, path string) string {
	return fmt.Sprintf("ls-%s-%s-%s-%s", namespace, pod, container, path)
}

func StatCommandCacheKey(containerId, path string) string {
	return fmt.Sprintf("stat-%s-%s", containerId, path)
}
