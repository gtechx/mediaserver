package gtprotocol

import (
	"bytes"
	"reflect"
	"unsafe"
	"utils"
)

type Protocol struct {
	DataSize int32
	MsgType  int32
	RoomName [32]byte
}

type ReqLoginProtocol struct {
	Protocol
	SessionId uint64
}

type RetLoginProtocol struct {
	Protocol
	Result int8
}

type DataTransProtocol struct {
	Protocol
	Data []byte
}

func GetRoomName(buff []byte) string {
	strbuff := buff[8:40]
	index := bytes.IndexByte(strbuff, 0)
	return string(strbuff[0:index])
}

func GetMsgType(buff []byte) int {
	var msgtype int16
	utils.BytesToNum(buff[4:8], &msgtype)

	return int(msgtype)
}

func (p *Protocol) Parse(buff []byte) error {
	err := utils.BytesToNum(buff[:4], &p.DataSize)

	if err != nil {
		return err
	}

	err = utils.BytesToNum(buff[4:8], &p.MsgType)

	if err != nil {
		return err
	}

	copy(p.RoomName[:], buff[8:40])

	return nil
}

func (r *ReqLoginProtocol) Parse(buff []byte) error {
	//fmt.Println("ReqLoginProtocol Parse buff:", buff)
	err := r.Protocol.Parse(buff)
	if err != nil {
		return err
	}
	//fmt.Println("ReqLoginProtocol Parse buff[38:38+8]:", buff[40:40+8])
	err = utils.BytesToNum(buff[40:40+8], &r.SessionId)
	if err != nil {
		return err
	}
	// fmt.Println("ReqLoginProtocol Parse SessionId:", r.SessionId)
	// fmt.Println("ReqLoginProtocol Parse RoomName:", r.RoomName)
	// fmt.Println("ReqLoginProtocol Parse MsgType:", r.MsgType)
	// fmt.Println("ReqLoginProtocol Parse DataSize:", r.DataSize)
	return nil
}

func (r *RetLoginProtocol) Parse(buff []byte) error {
	err := r.Protocol.Parse(buff)
	if err != nil {
		return err
	}
	err = utils.BytesToNum(buff[40:40+1], &r.Result)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataTransProtocol) Parse(buff []byte) error {
	err := d.Protocol.Parse(buff)
	if err != nil {
		return err
	}
	d.Data = append(d.Data, buff[40:40+d.DataSize]...)
	return nil
}

func (r *ReqLoginProtocol) ToBytes() []byte {
	var x reflect.SliceHeader
	x.Len = int(unsafe.Sizeof(ReqLoginProtocol{}))
	x.Cap = int(unsafe.Sizeof(ReqLoginProtocol{}))
	x.Data = uintptr(unsafe.Pointer(r))
	return *(*[]byte)(unsafe.Pointer(&x))
}

func (r *RetLoginProtocol) ToBytes() []byte {
	var x reflect.SliceHeader
	x.Len = int(unsafe.Sizeof(RetLoginProtocol{}))
	x.Cap = int(unsafe.Sizeof(RetLoginProtocol{}))
	x.Data = uintptr(unsafe.Pointer(r))
	return *(*[]byte)(unsafe.Pointer(&x))
}

func (p *Protocol) ToBytes() []byte {
	var x reflect.SliceHeader
	x.Len = int(unsafe.Sizeof(Protocol{}))
	x.Cap = int(unsafe.Sizeof(Protocol{}))
	x.Data = uintptr(unsafe.Pointer(p))
	return *(*[]byte)(unsafe.Pointer(&x))
}

func (d *DataTransProtocol) ToBytes() []byte {
	buff := make([]byte, 0)
	buff = append(buff, d.Protocol.ToBytes()...)
	buff = append(buff, d.Data...)
	return buff
}
