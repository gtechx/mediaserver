package utils

import (
	"bytes"
	"strconv"
	//"unsafe"
)

//int convert
func StrToInt(s string) (int, error) {
	i64, err := strconv.ParseInt(s, 10, 0)
	return int(i64), err
}

func StrToUint(s string) (uint, error) {
	ui64, err := strconv.ParseUint(s, 10, 0)
	return uint(ui64), err
}

func IntToStr(n int) string {
	return strconv.FormatInt(int64(n), 10)
}

func UintToStr(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
}

//int32 convert
func StrToInt32(s string) (int32, error) {
	i64, err := strconv.ParseInt(s, 10, 0)
	return int32(i64), err
}

func StrToUint32(s string) (uint32, error) {
	i64, err := strconv.ParseInt(s, 10, 0)
	return uint32(i64), err
}

func Int32ToStr(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

func Uint32ToStr(n uint32) string {
	return strconv.FormatUint(uint64(n), 10)
}

//int64 convert
func StrToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 0)
}

func StrToUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 0)
}

func Int64ToStr(n int64) string {
	return strconv.FormatInt(n, 10)
}

func Uint64ToStr(n uint64) string {
	return strconv.FormatUint(n, 10)
}

//float convert
func StrToFloat32(s string) (float32, error) {
	f, err := strconv.ParseFloat(s, 32)
	return float32(f), err
}

func StrToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func Floa32ToStr(f float32) string {
	return strconv.FormatFloat(float64(f), 'f', -1, 32)
}

func Floa64ToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func BytesToStr(buff []byte) string {
	index := bytes.IndexByte(buff, 0)
	return string(buff[0:index])
	//return *(*string)(unsafe.Pointer(&buff))
}
