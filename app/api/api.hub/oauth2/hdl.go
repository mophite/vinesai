package oauth2

import (
	"errors"
	"github.com/gogo/protobuf/proto"
	"net/http"
	"strings"
	"vinesai/internel/ava"
	"vinesai/internel/lib"
	"vinesai/internel/x"
	"vinesai/proto/phub"
)

/*
2.1 获取OAuth 2.0令牌 请求URL: /api/v1/token
方法: POST参数:
client_id: 应用的客户端ID
client_secret: 应用的客户端密钥
grant_type: 授权类型，password
code: 授权码(如果使用授权码流程)
返回:
access_token: 访问令牌
expires_in: 令牌的有效期(秒)
https://cloud.tencent.com/developer/article/1532760
*/
var cacheClientSecretTmp = map[string]string{"498715320649678": "JYNFA9OHGQBL5IU62ZWMXKPS1TRD73VC"}

//var codeTmp = map[string]string{"498715320649678": "2071635489279878"}

const (
	grantTypeClientCredentials = "client_credentials"
	grantTypeCode              = "code"
)

type Oauth2 struct{}

func (d *Oauth2) Token(c *ava.Context, req *phub.TokenReq, rsp *phub.TokenRsp) {
	if req.HomeId == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "homeId不能为空"
		return
	}

	if req.GrantType != grantTypeCode && req.GrantType != grantTypeClientCredentials {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "GrantType错误"
		return
	}

	s, ok := cacheClientSecretTmp[req.ClientId]
	if !ok {
		rsp.Code = http.StatusUnauthorized
		rsp.Msg = "商户不存在"
		return
	}

	if req.ClientSecret != s {
		rsp.Code = http.StatusUnauthorized
		rsp.Msg = "商户信息不正确"
		return
	}

	jwtToken, expiry := generateJWToken(c, req.HomeId)

	rsp.Code = http.StatusOK
	rsp.Msg = "获取token成功"
	rsp.Data = &phub.TokenData{
		AccessToken: jwtToken,
		ExpiresIn:   expiry,
		TokenType:   "Bearer",
	}
}

var whiteHttpPathList = map[string]bool{
	"/hub/oauth/token": true,
}

func Oauth(c *ava.Context) (proto.Message, error) {

	if _, ok := whiteHttpPathList[c.Metadata.Method()]; ok {
		return nil, nil
	}

	var rsp phub.CommonData
	//处理bearar
	bearer := c.GetHeader("Authorization")
	if bearer == "" {
		rsp.Code = 401
		rsp.Msg = "请求头Bearer信息缺失"
		return &rsp, errors.New("请求头Bearer信息缺失")
	}

	bearer = strings.TrimPrefix(bearer, "Bearer ")

	t, err := parseJWToken(bearer)
	if err != nil {
		rsp.Code = 401
		rsp.Msg = "身份验证错误或已过期，请重新尝试"
		return &rsp, errors.New("身份认证失败")
	}

	c.Infof("Oauth |data=%v", x.MustMarshal2String(t))

	lib.SetHomeId(c, t.Audience[0])

	return nil, nil
}
