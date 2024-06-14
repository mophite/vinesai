package user

import (
	"net/http"
	"strconv"
	"vinesai/internel/ava"
	"vinesai/internel/db"
	"vinesai/internel/x"
	"vinesai/proto/pha"

	"gorm.io/gorm"
)

type User struct {
}

func (u User) Login(c *ava.Context, req *pha.AuthReq, rsp *pha.AuthRsp) {
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

	remoteIp, err := x.RemoteIp()
	if err != nil {
		c.Error(err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = "获取外网ip失败"
		return
	}

	var data = &DBUser{
		Phone: req.Phone,
	}

	err = db.GMysql.Transaction(func(tx *gorm.DB) error {

		err = tx.Table(TableUser).Create(data).Error
		if err != nil {
			c.Error(err)
			return err
		}
		port := strconv.Itoa(data.ID)
		var updates = make(map[string]interface{}, 2)
		updates["ha_address_1"] = remoteIp + ":" + port
		updates["ha_address_2"] = localIp + ":" + port
		err = tx.Table(TableUser).Where("id=?", data.ID).Updates(updates).Error
		if err != nil {
			c.Error(err)
			return err
		}

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
