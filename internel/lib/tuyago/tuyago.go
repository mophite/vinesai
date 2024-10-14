package tuyago

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/lib"
	"vinesai/internel/x"

	uuid "github.com/satori/go.uuid"
)

var defaultClientID = "55tvajraw3edkr78qqa9"
var defaultKey = "9ad81ab89b384354bf6ad3f1bb0b7e2e"
var defaultApiHost = "https://openapi.tuyacn.com"
var defaultMsgHost = "pulsar://mqe.tuyacn.com:7285"

func generateSignature4Token(method, uri, nonce, ts string) string {
	return strings.ToUpper(hS256Sign(
		defaultKey,
		defaultClientID+ts+nonce+StringToSign(method, uri, nil),
	))
}

func generateSignature(method, uri, nonce, ts, accessToken string, body []byte) string {
	return strings.ToUpper(
		hS256Sign(
			defaultKey,
			defaultClientID+accessToken+ts+nonce+StringToSign(method, uri, body),
		))
}

func StringToSign(method, uri string, body []byte) string {

	contentSha256 := ""
	if body != nil {
		contentSha256 = getSha256(body)
	} else {
		contentSha256 = getSha256([]byte(""))
	}

	keys := make([]string, 0, 10)
	// 解析URL
	parsedURL, err := url.Parse(uri)
	if err != nil {
		ava.Error(err)
		return ""
	}

	p := parsedURL.Path

	form, err := url.ParseQuery(parsedURL.RawQuery)
	if err == nil {
		for key := range form {
			keys = append(keys, key)
		}
	}

	if len(keys) > 0 && err == nil {
		p += "?"
		sort.Strings(keys)
		for _, keyName := range keys {
			value := form.Get(keyName)
			p += keyName + "=" + value + "&"
		}
		p = strings.TrimSuffix(p, "&")
	}

	return method + "\n" + contentSha256 + "\n" + "" + "\n" + p
}

func getSha256(data []byte) string {
	sha256Contain := sha256.New()
	sha256Contain.Write(data)
	return hex.EncodeToString(sha256Contain.Sum(nil))
}

func hS256Sign(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

var defaultTokenURL = "https://openapi.tuyacn.com/v1.0/token"

type token struct {
	Result struct {
		AccessToken  string `json:"access_token"`
		ExpireTime   int64  `json:"expire_time"`
		RefreshToken string `json:"refresh_token"`
		UID          string `json:"uid"`
		ExpireTimeAt int64  `json:"expire_time_at"`
	} `json:"result"`
	Success bool   `json:"success"`
	T       int64  `json:"t"`
	Tid     string `json:"tid"`
}

var gToken *token
var mux = new(sync.Mutex)

// 获取token，快过期也重新获取
func instanceToken() (*token, error) {
	mux.Lock()
	defer mux.Unlock()

	var uri = defaultTokenURL + "?grant_type=1"
	//if gToken != nil && gToken.Result.ExpireTimeAt-time.Now().Unix() < 200 {
	//	uri = defaultTokenURL + "/" + gToken.Result.RefreshToken
	//}

	if gToken == nil || gToken.Result.ExpireTimeAt-time.Now().Unix() < 200 {
		//if gToken == nil || gToken.Result.ExpireTimeAt-time.Now().Unix() < 200 {

		var nonce = getUUID()
		var ts = getTs()

		var header = map[string]string{
			"client_id":    defaultClientID,
			"sign":         generateSignature4Token(http.MethodGet, uri, nonce, ts),
			"nonce":        nonce,
			"t":            ts,
			"sign_method":  "HMAC-SHA256",
			"Content-Type": "application/json",
		}

		b, err := lib.GetWithout(uri, header)
		if err != nil {
			return nil, err
		}

		ava.Debugf("FROM=%s |new_token |gToken=%s |uri=%s", string(b), x.MustMarshal2String(gToken), uri)
		gToken = new(token)
		err = x.MustUnmarshal(b, gToken)
		if err != nil {
			ava.Error(err)
			return nil, err
		}

		if !gToken.Success {
			return nil, errors.New("access_token is invalid")
		}

		gToken.Result.ExpireTimeAt = time.Now().Unix() + gToken.Result.ExpireTime
	}

	return gToken, nil
}

func getTs() string {
	return strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
}

func Get(c *ava.Context, uri string, v interface{}) error {
	var now = time.Now()

	uri = defaultApiHost + uri
	var nonce = getUUID()
	var ts = getTs()
	accessToken, err := instanceToken()
	if err != nil {
		c.Error(err)
		return err
	}

	var header = map[string]string{
		"client_id":    defaultClientID,
		"sign":         generateSignature(http.MethodGet, uri, nonce, ts, accessToken.Result.AccessToken, nil),
		"nonce":        nonce,
		"t":            ts,
		"sign_method":  "HMAC-SHA256",
		"access_token": accessToken.Result.AccessToken,
		"Content-Type": "application/json",
	}

	b, err := lib.Get(c, uri, header)
	if err != nil {
		c.Error(err)
		return err
	}

	c.Debugf("latency=%v秒 |uri=%v |FROM=%v", time.Now().Sub(now).Seconds(), uri, string(b))

	return x.MustNativeUnmarshal(b, v)
}

func getUUID() string {
	u2 := uuid.NewV4()
	return u2.String()
}

func Post(c *ava.Context, uri string, data, v interface{}) error {
	var now = time.Now()

	uri = defaultApiHost + uri
	var nonce = getUUID()
	var ts = getTs()
	accessToken, err := instanceToken()
	if err != nil {
		c.Error(err)
		return err
	}

	var body = x.MustMarshal(data)

	var header = map[string]string{
		"client_id":    defaultClientID,
		"sign":         generateSignature(http.MethodPost, uri, nonce, ts, accessToken.Result.AccessToken, body),
		"nonce":        nonce,
		"t":            ts,
		"sign_method":  "HMAC-SHA256",
		"access_token": accessToken.Result.AccessToken,
		"Content-Type": "application/json",
	}

	b, err := lib.POST(c, uri, body, header)
	if err != nil {
		c.Error(err)
		return err
	}

	c.Debugf("latency=%v秒 ｜uri=%s |TO=%v |FROM=%v", time.Now().Sub(now).Seconds(), uri, string(body), string(b))

	return x.MustNativeUnmarshal(b, v)
}

func Put(c *ava.Context, uri string, data, v interface{}) error {
	var now = time.Now()

	uri = defaultApiHost + uri
	var nonce = getUUID()
	var ts = getTs()
	accessToken, err := instanceToken()
	if err != nil {
		c.Error(err)
		return err
	}

	var body = x.MustMarshal(data)

	var header = map[string]string{
		"client_id":    defaultClientID,
		"sign":         generateSignature(http.MethodPut, uri, nonce, ts, accessToken.Result.AccessToken, body),
		"nonce":        nonce,
		"t":            ts,
		"sign_method":  "HMAC-SHA256",
		"access_token": accessToken.Result.AccessToken,
		"Content-Type": "application/json",
	}

	b, err := lib.PUT(c, uri, body, header)
	if err != nil {
		c.Error(err)
		return err
	}

	c.Debugf("latency=%v秒 ｜uri=%s |TO=%v |FROM=%v", time.Now().Sub(now).Seconds(), uri, string(body), string(b))

	return x.MustNativeUnmarshal(b, v)
}
