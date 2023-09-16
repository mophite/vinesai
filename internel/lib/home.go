package lib

import "vinesai/internel/ava"

const OauthKey2HomeId = "OauthKey2HomeId"

func SetHomeId(c *ava.Context, homeId string) {
	c.Set(OauthKey2HomeId, homeId)
}

func GetHomeId(c *ava.Context) string {
	return c.GetString(OauthKey2HomeId)
}
