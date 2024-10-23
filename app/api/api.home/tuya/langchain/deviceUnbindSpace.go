package langchain

import (
	"context"
	"vinesai/internel/ava"
	"vinesai/internel/db"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// 设备绑定
type deviceUnbindSpace struct {
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

func (o *deviceUnbindSpace) Call(c *ava.Context) error {

	//todo 同步redis
	//查询数据
	//同步redis数据
	//更新离线设备数据库
	filter := bson.M{"_id": o.BizData.DevID}
	_, err := db.Mgo.Collection(mgoCollectionNameDevice).DeleteOne(context.Background(), filter)
	if err != nil {
		c.Error(err)
		return err
	}

	return nil
}
