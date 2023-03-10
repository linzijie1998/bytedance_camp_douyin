// Code generated by hertz generator. DO NOT EDIT.

package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	douyin_base "github.com/linzijie1998/bytedance_camp_douyin/biz/router/douyin/base"
	douyin_interact "github.com/linzijie1998/bytedance_camp_douyin/biz/router/douyin/interact"
	douyin_relation "github.com/linzijie1998/bytedance_camp_douyin/biz/router/douyin/relation"
)

// GeneratedRegister registers routers generated by IDL.
func GeneratedRegister(r *server.Hertz) {
	//INSERT_POINT: DO NOT DELETE THIS LINE!
	douyin_relation.Register(r)

	douyin_interact.Register(r)

	douyin_base.Register(r)
}
