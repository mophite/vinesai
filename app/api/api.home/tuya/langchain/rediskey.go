package langchain

import "fmt"

var defaultActionCategoryListKey = "TUYA__ACTION_CATEGORY_"
var defaultActionShortDevicesKey = "TUYA_ACTION_SHORT_DEVICES_"
var defaultActionCommandKey = "TUYA_ACTION_COMMAND_%s_%s_%s_%s_%s" //action+productId+valueDesc
var defalutActionNameDeviceKey = "TUYA_ACTION_NAME_SHORT_DEVICES_"

func getActionCategoryListKey(category, productId, action, valueDesc, value string) string {
	return fmt.Sprintf(defaultActionCommandKey, category, productId, action, valueDesc, value)
}
