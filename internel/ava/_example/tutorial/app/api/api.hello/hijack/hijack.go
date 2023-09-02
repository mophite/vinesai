package hijack

import (
	"net/http"
	"vinesai/internel/ava"
)

func HijackWriter(c *ava.Context, r *http.Request, w http.ResponseWriter, req, rsp *ava.Packet) bool {
	w.Write(rsp.Bytes())
	return true
}
