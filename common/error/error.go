package gterror

type Error struct {
	ErrorMsg  string `json:"error"`
	ErrorCode int    `json:"errorcode"`
}
