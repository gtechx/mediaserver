package gtroom

import (
	"net"
)

type Room struct {
	Name string
	ID   uint32
}

type RSRoom struct {
	Room
	ClientList map[string]*net.UDPAddr
	BSList     []*net.UDPConn
}

type BSRoom struct {
	Room
	ClientList map[string]*net.UDPAddr
}

type SCRoom struct {
	Room
	BSList      []string
	RSList      []string
	IsPublic    int8
	HasPassword int8
	Password    string
	RoomType    string
}

func NewSCRoom(name string, ispublic int8, haspassword int8, password string, roomtype string) *SCRoom {
	return &SCRoom{Room: Room{name, 0}, BSList: make([]string, 0), RSList: make([]string, 0), IsPublic: ispublic, HasPassword: haspassword, Password: password, RoomType: roomtype}
}

func NewRSRoom(name string) *RSRoom {
	return &RSRoom{Room: Room{name, 0}, ClientList: make(map[string]*net.UDPAddr), BSList: make([]*net.UDPConn, 0)}
}

func NewBSRoom(name string) *BSRoom {
	return &BSRoom{Room: Room{name, 0}, ClientList: make(map[string]*net.UDPAddr)}
}

func (r *RSRoom) AddClient(strsid string, raddr *net.UDPAddr) {
	r.ClientList[strsid] = raddr
}

func (r *RSRoom) AddBS(bsconn *net.UDPConn) {
	r.BSList = append(r.BSList, bsconn)
}

func (r *BSRoom) AddClient(strsid string, raddr *net.UDPAddr) {
	r.ClientList[strsid] = raddr
}

// func (r *Room) ProcessMsg(proto gtprotocol.Protocol, packet gtnet.GTUDPPacket) {
// 	if proto.MsgType == common.MSG_REQ_LOGIN {
// 		a := 0
// 	} else if proto.MsgType == common.MSG_DATA_TRANS {
// 		rsroom
// 	}
// }

// func (r *RSRoom) ProcessMsg(proto gtprotocol.Protocol, packet gtnet.GTUDPPacket) {
// 	r.Room.ProcessMsg(proto, packet)

// 	if proto.MsgType == common.MSG_REQ_LOGIN {
// 		a := 0
// 	} else if proto.MsgType == common.MSG_DATA_TRANS {
// 		rsroom
// 	}
// }
