/*
@Time : 2022/1/6 10:21
@Author : sunxy
@File : exec
@description:
*/
package util

import (
	"bytes"
	"os/exec"
)

func BashCommand(cmd string) []string {
	return []string{"/bin/bash", "-c", cmd}
}

// ExecShell use bash to execute command
func ExecShell(command []string) (string, string, error) {
	var cmd *exec.Cmd
	if len(command) > 1 {
		cmd = exec.Command(command[0], command[1:]...)
	} else {
		cmd = exec.Command(command[0])
	}
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var outBuffer = new(bytes.Buffer)
	var errBuffer = new(bytes.Buffer)
	cmd.Stdout = outBuffer
	cmd.Stderr = errBuffer

	//Run执行cmd包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	err := cmd.Run()
	if err != nil {
		return outBuffer.String(), errBuffer.String(), err
	} else {
		return outBuffer.String(), errBuffer.String(), nil
	}
}
