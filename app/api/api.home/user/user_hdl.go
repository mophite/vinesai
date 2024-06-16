package user

import (
	"net/http"
	"strconv"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/proto/pha"

	"gorm.io/gorm"
)

type User struct {
	RemoteIp string
}

func (u *User) Code(c *ava.Context, req *pha.CodeReq, rsp *pha.CodeRsp) {
	//todo 检验手机号码是否正确
	//todo 判断用户是否已经存在
	//todo 发送短信验证码

	rsp.Code = http.StatusOK
	rsp.Msg = "短信验证码发送成功"
}

func (u *User) Login(c *ava.Context, req *pha.AuthReq, rsp *pha.AuthRsp) {
	if req.Phone == "" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "请输入手机号"
		return
	}

	localIp, err := ava.LocalIp()
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "获取内网ip失败"
		return
	}

	//remoteIp, err := x.RemoteIp()
	//if err != nil {
	//	c.Error(err)
	//	rsp.Code = http.StatusInternalServerError
	//	rsp.Msg = "获取外网ip失败"
	//	return
	//}

	var data = &DBUser{
		Phone: req.Phone,
	}

	err = db.GMysql.Transaction(func(tx *gorm.DB) error {

		//todo 判断用户是否存在，如果存在就只登录
		err = tx.Table(TableUser).Create(data).Error
		if err != nil {
			c.Error(err)
			return err
		}
		port := strconv.Itoa(data.ID)

		data.HaAddress1 = u.RemoteIp + ":" + port
		data.HaAddress2 = localIp + ":" + port

		var updates = make(map[string]interface{}, 2)
		updates["ha_address_1"] = data.HaAddress1
		updates["ha_address_2"] = data.HaAddress2
		err = tx.Table(TableUser).Where("id=?", data.ID).Updates(updates).Error
		if err != nil {
			c.Error(err)
			return err
		}

		c.Debugf("USER |login |prot=%v", port)
		//创建docker
		err = createDocker(c, port)
		if err != nil {
			c.Error(err)
			return err
		}

		return nil
	})

	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "实例创建失败"
		return
	}

	redirect := data.HaAddress1

	token, _ := generateJWToken(c, req.Phone, redirect)

	rsp.Code = http.StatusOK
	rsp.Msg = "登录成功"
	rsp.Data = &pha.AuthRspData{
		RedirectUrl: redirect,
		AccessToken: token,
	}
}
