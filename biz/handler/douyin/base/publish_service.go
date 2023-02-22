// Code generated by hertz generator.

package base

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/cache"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/dal"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/handler/douyin"
	base "github.com/linzijie1998/bytedance_camp_douyin/biz/model/douyin/base"
	"github.com/linzijie1998/bytedance_camp_douyin/global"
	"github.com/linzijie1998/bytedance_camp_douyin/model"
	"github.com/linzijie1998/bytedance_camp_douyin/util"
	"path/filepath"
)

// PublishAction 需要JWT鉴权, 验证通过会在请求上下文中添加"token_user_id"信息.
// @router /douyin/publish/action/ [POST]
func PublishAction(ctx context.Context, c *app.RequestContext) {
	resp := new(base.PublishActionResp)
	// 从请求上下文中提取"token_user_id"
	title := c.PostForm("title")
	rawID, exists := c.Get("token_user_id")
	if !exists {
		global.DOUYIN_LOGGER.Debug("未从请求上下文中解析到userID")
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}
	userID := int64(rawID.(uint))

	file, err := c.FormFile("data")
	if err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("未接收到视频数据 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	filename, err := util.GenerateFilenameByUploadDatetime(file.Filename)
	if err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("文件名生成错误 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	video := model.Video{
		UserInfoID: userID,
		Title:      title,
		VideoPath:  filename,
		CoverPath:  fmt.Sprintf("%s.jpg", filename),
	}

	videoPath := filepath.Join(global.DOUYIN_CONFIG.Upload.UploadRoot, "videos", video.VideoPath)
	coverPath := filepath.Join(global.DOUYIN_CONFIG.Upload.UploadRoot, "covers", video.CoverPath)

	if err := c.SaveUploadedFile(file, videoPath); err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("视频保存失败 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	if err := util.GetCover(videoPath, coverPath, "00:00:02"); err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("视频封面截取失败 err: %v", err))
		video.CoverPath = ""
	}

	if err := dal.CreateVideoInfo(&video); err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("视频信息存储失败 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	// redis 用户计数
	if err := cache.UpdateWorkCount(userID); err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("用户作品计数失败 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	c.JSON(consts.StatusOK, resp)
	global.DOUYIN_LOGGER.Info(fmt.Sprintf("ID为%d的用户上传了视频%s", userID, videoPath))
}

// PublishList .
// @router /douyin/publish/list/ [GET]
func PublishList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req base.PublishListReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	resp := new(base.PublishListResp)

	var userID int64
	// 登录状态下查看发布视频的列表
	if req.Token != "" {
		j := util.NewJWT()
		claim, err := j.ParseToken(req.Token)
		if err != nil {
			global.DOUYIN_LOGGER.Info(fmt.Sprintf("Token解析失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusBadRequest, resp)
			return
		}
		userID = int64(claim.UserInfo.ID)
	}

	// 查询req.UserID发布的视频
	videoInfos, err := dal.QueryVideoInfoByUserInfoID(req.UserID)
	if err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("ID为%d的用户上传的视频信息查询失败 err: %v", userID, err))
		resp.StatusCode = 1
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	// 查询发布者的用户信息
	userInfos, err := dal.QueryUserInfoByUserID(req.UserID)
	if err != nil || len(userInfos) > 1 {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("ID为%d的用户信息查询失败 err: %v", userID, err))
		resp.StatusCode = 1
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	if len(userInfos) != 1 {
		global.DOUYIN_LOGGER.Info(fmt.Sprintf("未找到ID为%d的用户信息", req.UserID))
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	var user = new(base.User)
	if err = douyin.UserInfoSupplement(userID, user, &userInfos[0]); err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("查询用户信息补充失败 err:%v", err))
		return
	}

	videoList := make([]*base.Video, len(videoInfos))
	for i, info := range videoInfos {

		var video = new(base.Video)
		video.ID = int64(info.ID)
		video.Author = user
		if err = douyin.VideoInfoSupplement(userID, video, &info); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("查询视频信息补充失败 err:%v", err))
			return
		}

		videoList[i] = video
	}
	resp.VideoList = videoList
	c.JSON(consts.StatusOK, resp)
}
