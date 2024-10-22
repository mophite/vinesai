package langchain

import (
	"context"
	"fmt"
	"vinesai/internel/db"
	"vinesai/internel/langchaingo/llms"
	"vinesai/internel/lib/tuyago"
	"vinesai/internel/x"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// 场景设置
type scene struct{ CallbacksHandler LogHandler }

func (s *scene) Name() string {
	return "setting_one_click_scene"
}

func (s *scene) Description() string {
	return "创建设置智能家居一键执行场景"
}

// 场景背景图，用于查询默认场景图⽚列表。
//
// URL
// GET AY/v1.0/scenes/default-pictures
// 写死一张发送，其他的都很丑
var defaultBackgroudPicture = "https://images.tuyacn.com/smart/rule/cover/starry.png"

func (s *scene) Call(ctx context.Context, input string) (string, error) {

	var c = fromCtx(ctx)
	var homeId = getHomeId(c)
	var msg = "请告诉我更详细的规则"
	input = getFirstInput(c)

	//获取支持场景的设备列表
	var ssResp supportSceneDevices
	err := tuyago.Get(c, fmt.Sprintf("/v1.0/homes/%s/scene/devices", homeId), &ssResp)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	if len(ssResp.Result) == 0 {
		return "你没有可以创建场景的设备", err
	}

	////通过ai返回需要的设备信息
	//mcList := []llms.MessageContent{
	//	{
	//		Role:  llms.ChatMessageTypeSystem,
	//		Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(onClickSceneDevicesPrompts, x.MustMarshal2String(ssResp.Result)))},
	//	},
	//	{
	//		Role:  llms.ChatMessageTypeHuman,
	//		Parts: []llms.ContentPart{llms.TextPart(input)},
	//	},
	//}
	//
	//var resultDevices struct {
	//	Result []struct {
	//		Name     string `json:"name"`
	//		DeviceId string `json:"device_id"`
	//	} `json:"result"`
	//	FailureMsg string `json:"failure_Msg"`
	//}
	//
	//err = GenerateContentWithout(c, mcList, &resultDevices)
	//if err != nil {
	//	c.Error(err)
	//	return msg, err
	//}
	//
	//if len(resultDevices.Result) == 0 {
	//	return resultDevices.FailureMsg, err
	//}
	//
	//ids := make([]string, 0, len(resultDevices.Result))
	//for _, result := range resultDevices.Result {
	//	ids = append(ids, result.DeviceId)
	//}

	filter := bson.M{"homeid": homeId}

	//获取筛选后的设备支持的联动规则，指令
	cur, err := db.Mgo.Collection(mgoCollectionNameCodes).Find(context.Background(), filter)
	if err != nil {
		c.Error(err)
		return "开了点小差，重试一下", err
	}

	var codesResp []homeFunction
	err = cur.All(ctx, &codesResp)
	if err != nil {
		c.Error(err)
		return "开了点小差，重试一下", err
	}
	fmt.Println("----1--", codesResp)

	//通过ai返回创建一键场景的数据

	//通过ai返回需要的设备信息
	mcList1 := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(fmt.Sprintf(onClickSceneCreatePrompts, removeWhitespace(x.MustMarshal2String(codesResp))))},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(input)},
		},
	}

	var resultAction createScene

	err = GenerateContentWithout(c, mcList1, &resultAction)
	if err != nil {
		c.Error(err)
		return msg, err
	}

	if len(resultAction.Actions) == 0 {
		return "创建场景失败了", fmt.Errorf("创建场景失败了 ｜homeid=%s |input=%s", homeId, input)
	}

	var createSceneResp struct {
		Success bool   `json:"success"`
		T       int64  `json:"t"`
		Result  string `json:"result"` //返回场景id
	}

	resultAction.Background = defaultBackgroudPicture

	//添加场景
	err = tuyago.Post(c, fmt.Sprintf("/v1.0/homes/%s/scenes", homeId), resultAction, &createSceneResp)
	if err != nil || !createSceneResp.Success {
		c.Error(err)
		return "场景创建失败了", err
	}

	return resultAction.Content, nil
}

var onClickSceneCreatePrompts = `分析我的意图，从指令数据信息中选择指令，并严格按照返回的JSON格式返回我即将调用创建场景接口的数据
### 设备以及指令数据信息: %s

### 返回json格式：
{
 "content":"根据我的意图和创建成功或失败后用人性化的语言告诉我",
 "name": "关闭灯光",
 "actions": [
   {
	 "executor_property": { "switch_1": true },
     "action_executor": "dpIssue",
     "entity_id": "6c3f4cb6c5899478efrgea"
   }
 ]
}

说明：
1.entity_id就是device_id
2.actions：数据对象，一个对象元素只能有一个指令，不能在一个对象中出现多个指令

### 示例：
设备以及指令数据： {"device_id":"6c3f4cb6c5899478efrgea","functions":[{"values":{},"code":"switch_1","name":"开关1","type":"Boolean","value_range_json":[[true,"开启"],[false,"关闭"]]},{"type":"Boolean","value_range_j[true,"开启"],[false,"关闭"]],"values":{},"code":"switch_2","name":"开关2"},{"value_range_json":[[true,"开启"],[false,"关闭"]],"values":{},"code":"switch_3","name":"开关3","type":"Boolean"},{"value_ratrue,"开启"],[false,"关闭"]],"values":{},"code":"switch_4","name":"开关4","type":"Boolean"},{"name":"开关5","type":"Boolean","value_range_json":[[true,"开启"],[false,"关闭"]],"values":{},"code":"switc":"开关6","type":"Boolean","value_range_json":[[true,"开启"],[false,"关闭"]],"values":{},"code":"switch_6"},{"values":{},"code":"child_lock","name":"童锁","type":"Boolean","value_range_json":[[true,"开se,"关闭"]]}]}

输入：创建一个关闭客厅插座场景

返回：
{
 "content":"关闭客厅插座场景创建成功",
 "name": "关闭客厅插座",
 "actions": [
        {
          "action_executor": "dpIssue",
          "entity_id": "6c3f4cb6c5899478efrgea",
          "executor_property": {
            "switch_1": true
          }
        },
        {
          "action_executor": "dpIssue",
          "entity_id": "6c3f4cb6c5899478efrgea",
          "executor_property": {
            "switch_2": true
          }
        },
        {
          "action_executor": "dpIssue",
          "entity_id": "6c3f4cb6c5899478efrgea",
          "executor_property": {
            "switch_3": true
          }
        },
        {
          "action_executor": "dpIssue",
          "entity_id": "6c3f4cb6c5899478efrgea",
          "executor_property": {
            "switch_4": true
          }
        },
        {
          "action_executor": "dpIssue",
          "entity_id": "6c3f4cb6c5899478efrgea",
          "executor_property": {
            "switch_5": true
          }
        },
        {
          "action_executor": "dpIssue",
          "entity_id": "6c3f4cb6c5899478efrgea",
          "executor_property": {
            "switch_6": true
          }
        }
 ]
}`
