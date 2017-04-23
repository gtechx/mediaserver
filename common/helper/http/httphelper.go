package httphelper

import (
	"io"
	"net/http"
	"utils"
)

func WriteError(rw http.ResponseWriter, errcode int, errmsg string) {
	io.WriteString(rw, "{\"errorcode\":"+utils.IntToStr(errcode)+", \"error\":\""+errmsg+"\"}")
}
