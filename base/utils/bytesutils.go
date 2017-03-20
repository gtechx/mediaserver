package utils

import (
	"bytes"
	"encoding/binary"
)

func NumToBytes(num interface{}) ([]byte, error) {
	//bytesBuffer := bytes.NewBuffer([]byte{})
	bytesBuffer := new(bytes.Buffer)

	err := binary.Write(bytesBuffer, binary.LittleEndian, num)
	if err != nil {
		return bytesBuffer.Bytes(), err
	}

	return bytesBuffer.Bytes(), nil
}

func BytesToNum(b []byte, num interface{}) error {
	buf := bytes.NewReader(b)
	return binary.Read(buf, binary.LittleEndian, num)
}

func AppendNumBytes(dst []byte, num interface{}) ([]byte, error) {
	buf, err := NumToBytes(num)

	if err != nil {
		return dst, err
	} else {
		return append(dst, buf...), nil
	}
}

// func Int32ToBytes(i int32) []byte {
// 	var buf []byte
// 	bytesBuffer := bytes.NewBuffer([]byte{})
// 	binary.Write(bytesBuffer, binary.LittleEndian, i)
// 	buf = append(buf, bytesBuffer.Bytes()...)

// 	return buf
// }

// func Int64ToBytes(i int64) []byte {
// 	var buf []byte
// 	bytesBuffer := bytes.NewBuffer([]byte{})
// 	binary.Write(bytesBuffer, binary.LittleEndian, i)
// 	buf = append(buf, bytesBuffer.Bytes()...)

// 	return buf
// }
