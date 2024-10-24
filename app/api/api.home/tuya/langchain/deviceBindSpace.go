package langchain

import (
	"vinesai/internel/ava"
	"vinesai/internel/x"
)

// 设备绑定
type deviceBindSpace struct {
	BizCode string `json:"bizCode"`
	BizData struct {
		DevID     string `json:"devId"`
		UID       string `json:"uid"`
		SpaceID   string `json:"spaceId"`
		ProductID string `json:"productId"`
		UUID      string `json:"uuid"`
		Token     string `json:"token"`
	} `json:"bizData"`
	Ts int64 `json:"ts"`
}

func (o *deviceBindSpace) Call(c *ava.Context) error {

	//todo 同步redis
	//查询数据
	//同步redis数据
	//更新离线设备数据库
	//var tmpDevicesResp = &deviceResp{}
	//
	//err := tuyago.Get(c, fmt.Sprintf("/v1.0/devices/%s", d.ID), tmpDevicesResp)
	//
	//if err != nil {
	//	ava.Error(err)
	//	return
	//}
	//tuyago.Get(c, "")
	//filter := bson.M{"_id": o.BizData.DevID}
	//_, err := db.Mgo.Collection(mgoCollectionDevice).insert(context.Background(), filter)
	//if err != nil {
	//	c.Error(err)
	//	return err
	//}

	//暂时不做
	c.Debugf("deviceBindSpace |data=%s", x.MustMarshal2String(o))

	return nil
}
