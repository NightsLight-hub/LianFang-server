/*
@Time : 2021/12/30 10:30
@Author : sunxy
@File : errors
@description:
*/
package util

func IsResourceNotExistError(err error) bool {
	switch err.(type) {
	case *ResourceNotExistError:
		return true
	default:
		return false
	}
}

func IsTarCmdFailedError(err error) bool {
	switch err.(type) {
	case *TarCmdFailedError:
		return true
	default:
		return false
	}
}

type ResourceNotExistError struct {
	Msg string
}

func (r ResourceNotExistError) Error() string {
	return r.Msg
}

type TarCmdFailedError struct {
	Msg string
}

func (t TarCmdFailedError) Error() string {
	return t.Msg
}
