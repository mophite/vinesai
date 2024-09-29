package connector

import (
	"vinesai/internel/ava"
	_ "vinesai/internel/lib/connector/header"
	_ "vinesai/internel/lib/connector/message"
	_ "vinesai/internel/lib/connector/sign"
	_ "vinesai/internel/lib/connector/token"

	"context"
	"net/http"

	"vinesai/internel/lib/connector/constant"
	"vinesai/internel/lib/connector/env"
	"vinesai/internel/lib/connector/env/extension"
	"vinesai/internel/lib/connector/httplib"
)

type ParamFunc func(*httplib.ProxyHttp)

// init env config
// init handler
func InitWithOptions(opts ...env.OptionFunc) {
	env.Config = env.NewEnv()
	for _, v := range opts {
		v(env.Config)
	}
	env.Config.Init()
}

// make request api
func makeRequest(ctx context.Context, params ...ParamFunc) error {
	defer func() {
		if v := recover(); v != nil {
			ava.Errorf("unknown error, err=%+v", v)
		}
	}()
	ph := httplib.NewProxyHttp()
	params = append(params, WithErrProc(constant.TOKEN_EXPIRED, &tokenError{ph: ph}))
	for _, p := range params {
		p(ph)
	}
	ctx = context.WithValue(ctx, constant.REQ_INFO, ph.GetReqHandler())
	// set header
	ph.SetHeader(extension.GetHeader(constant.TUYA_HEADER).Do(ctx))

	//get req
	return ph.DoRequest(ctx)
}

// GET
func MakeGetRequest(ctx context.Context, params ...ParamFunc) error {
	params = append(params, withMethod(http.MethodGet))
	return makeRequest(ctx, params...)
}

// POST
func MakePostRequest(ctx context.Context, params ...ParamFunc) error {
	params = append(params, withMethod(http.MethodPost))
	return makeRequest(ctx, params...)
}

// PUT
func MakePutRequest(ctx context.Context, params ...ParamFunc) error {
	params = append(params, withMethod(http.MethodPut))
	return makeRequest(ctx, params...)
}

// DELETE
func MakeDeleteRequest(ctx context.Context, params ...ParamFunc) error {
	params = append(params, withMethod(http.MethodDelete))
	return makeRequest(ctx, params...)
}

// PATCH
func MakePatchRequest(ctx context.Context, params ...ParamFunc) error {
	params = append(params, withMethod(http.MethodPatch))
	return makeRequest(ctx, params...)
}

// HEAD
func MakeHeadRequest(ctx context.Context, params ...ParamFunc) error {
	params = append(params, withMethod(http.MethodHead))
	return makeRequest(ctx, params...)
}

func WithHeader(h map[string]string) ParamFunc {
	return func(v *httplib.ProxyHttp) {
		v.SetHeader(h)
	}
}

func withMethod(method string) ParamFunc {
	return func(v *httplib.ProxyHttp) {
		v.SetMethod(method)
	}
}

func WithAPIUri(uri string) ParamFunc {
	return func(v *httplib.ProxyHttp) {
		v.SetAPIUri(env.Config.GetApiHost() + uri)
	}
}

func WithPayload(body []byte) ParamFunc {
	return func(v *httplib.ProxyHttp) {
		v.SetPayload(body)
	}
}

func WithResp(res interface{}) ParamFunc {
	return func(v *httplib.ProxyHttp) {
		v.SetResp(res)
	}
}

func WithErrProc(code int, f extension.IError) ParamFunc {
	return func(v *httplib.ProxyHttp) {
		v.SetErrProc(code, f)
	}
}

type tokenError struct {
	ph *httplib.ProxyHttp
}

func (t *tokenError) Process(ctx context.Context, code int, msg string) {
	if code == constant.TOKEN_EXPIRED {
		_, _ = extension.GetToken(constant.TUYA_TOKEN).Refresh(ctx)
		t.ph.SetPayload(t.ph.GetPayload())
		t.ph.SetHeader(extension.GetHeader(constant.TUYA_HEADER).Do(ctx))
		t.ph.DoRequest(ctx)
	}
}
