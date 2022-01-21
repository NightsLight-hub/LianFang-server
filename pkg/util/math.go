/*
@Time : 2021/12/21 19:21
@Author : sunxy
@File : util
@description:
*/
package util

import (
	"math"
)

func Round(f float64, n int) float64 {
	pow10N := math.Pow10(n)
	return math.Trunc((f+0.5/pow10N)*pow10N) / pow10N
}
