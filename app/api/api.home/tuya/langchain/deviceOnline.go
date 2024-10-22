package langchain

import (
	"context"
	"vinesai/internel/ava"
	"vinesai/internel/db"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// 设备离线
type deviceOnline struct {
	BizCode string `json:"bizCode"`
	BizData struct {
		DevID     string `json:"devId"`
		UID       string `json:"uid"`
		ProductID string `json:"productId"`
		Time      int64  `json:"time"`
	} `json:"bizData"`
	Ts int64 `json:"ts"`
}

func (o *deviceOnline) Call(c *ava.Context) error {

	//更新离线设备数据库
	filter := bson.M{"_id": o.BizData.DevID}
	update := bson.M{"$set": bson.M{"online": true}}
	_, err := db.Mgo.Collection(mgoCollectionNameDevice).UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.Error(err)
		return err
	}

	return nil
}
