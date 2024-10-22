package langchain

import (
	"context"
	"strings"
	"vinesai/internel/ava"
	"vinesai/internel/db"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// 设备离线
type devicePropertyMessage struct {
	BizCode string `json:"bizCode"`
	BizData struct {
		DataID     string `json:"dataId"`
		DevID      string `json:"devId"`
		ProductID  string `json:"productId"`
		Properties []struct {
			Code  string      `json:"code"`
			DpID  int         `json:"dpId"`
			Time  int64       `json:"time"`
			Value interface{} `json:"value"`
		} `json:"properties"`
	} `json:"bizData"`
	Ts int64 `json:"ts"`
}

func (o *devicePropertyMessage) Call(c *ava.Context) error {

	for i := range o.BizData.Properties {
		status := o.BizData.Properties[i]
		// 使用有效的占位符，如 elem1, elem2 作为数组元素的过滤器
		// 在过滤器中匹配 code 为实际数据中的值
		filter := bson.D{
			{"_id", o.BizData.DevID},
			{"status.code", bson.D{{"$regex", "^" + status.Code}}},
		}

		// 设置需要更新的字段和值
		update := bson.D{
			{"$set", bson.D{
				{"status.$.value", status.Value}, // 假设要将 value 更新为 true
			}},
		}

		err := db.Mgo.Collection(mgoCollectionNameDevice).
			FindOneAndUpdate(context.Background(), filter, update).Err()
		if err != nil {
			c.Error(err)
			return err
		}
	}

	return nil
}

func toCamelCase(s string) string {
	components := strings.Split(s, "_")
	for i, component := range components[1:] { // 从第二个元素开始处理
		components[i+1] = strings.Title(component) // 首字母大写
	}
	return strings.Join(components, "")
}
