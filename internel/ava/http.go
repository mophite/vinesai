package ava

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

func handlerServerHttp(c *Context, s *Server, w http.ResponseWriter, r *http.Request) {
	c.Metadata.SetMethod(r.URL.Path)

	c.Writer = w
	c.Request = r

	for k, v := range r.Header {
		if len(v) == 0 {
			continue
		}
		c.SetHeader(k, v[0])
	}

	if trace := c.GetHeader(defaultHeaderTrace); trace != "" {
		c.trace.WithTrace(trace)
	}

	c.ContentType = c.GetHeader(defaultHeaderContentType)
	c.setCodec()

	c.RemoteAddr = r.RemoteAddr

	w.Header().Set(defaultHeaderContentType, c.ContentType)
	w.Header().Set(defaultHeaderTrace, c.trace.TraceId())

	for i := range s.opts.Dog {
		rsp, err := s.opts.Dog[i](c)
		if err != nil {
			c.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write(c.Codec().MustEncode(rsp))
			return
		}
	}

	var err error
	var req *Packet
	var rsp *Packet
	defer func() {
		if req != nil {
			Recycle(req)
		}

		if rsp != nil {
			Recycle(rsp)
		}
	}()

	switch r.Method {
	case http.MethodPost, http.MethodDelete:
		req, rsp = payloadIo(r.Body), newPacket()

		_ = r.Body.Close()

		err = s.route.RR(c, req, rsp)

	case http.MethodPut:
		var f multipart.File
		var h *multipart.FileHeader
		f, h, err = r.FormFile("file")
		if err != nil {
			c.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write(s.opts.Err.Error400(c))
			return
		}

		var buf = bytes.NewBuffer(make([]byte, 0, 10485760))

		io.Copy(buf, f)

		var fileReq = &HttpFileReq{}
		fileReq.Body = buf.Bytes()
		fileReq.FileSize = h.Size
		fileReq.FileName = h.Filename
		fileReq.Extra = r.FormValue("extra")

		var fb []byte
		fb, err = c.Codec().Encode(fileReq)
		if err != nil {
			c.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write(s.opts.Err.Error400(c))
			return
		}

		req, rsp = payload2avaPacket(fb), newPacket()

		c.isPutFile = true
		err = s.route.RR(c, req, rsp)

		_ = r.Body.Close()

	case http.MethodGet:

		if strings.Count(r.URL.Path, "/") == 1 && s.opts.HttpGetRootPath != "" {
			c.Metadata.M = s.opts.HttpGetRootPath
		}

		values := r.URL.Query()

		var apiReq = &HttpApiReq{Params: make(map[string]string, len(values))}
		for k, v := range values {
			if len(v) > 0 {
				apiReq.Params[k] = v[0]
			}
		}

		var fb []byte
		fb, err = c.Codec().Encode(apiReq)
		if err != nil {
			c.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write(s.opts.Err.Error400(c))
			return
		}

		req, rsp = payload2avaPacket(fb), newPacket()

		err = s.route.RR(c, req, rsp)

	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(s.opts.Err.Error405(c))
		return
	}

	for i := range s.opts.Hijacker {
		if s.opts.Hijacker[i](c, r, w, req, rsp) {
			return
		}
	}

	if err != nil {
		c.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(s.opts.Err.Error500(c))
		return
	}

	if len(rsp.Bytes()) > 0 {
		w.WriteHeader(http.StatusOK)
		w.Write(rsp.Bytes())
	}
}
