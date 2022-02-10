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

func Uint16ToBytes(p uint16) []byte {
	var first8 uint16 = 0b_11111111_00000000
	var last8 uint16 = 0b_00000000_11111111
	bf := make([]byte, 2)
	bf[0] = uint8(p & first8 >> 8)
	bf[1] = uint8(p & last8)
	return bf
}
