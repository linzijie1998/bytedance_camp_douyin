// Code generated by hertz generator.

package interact

import (
	"context"
	"fmt"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/cache"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/handler/douyin"
	"github.com/linzijie1998/bytedance_camp_douyin/global"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/dal"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/model/douyin/base"
	interact "github.com/linzijie1998/bytedance_camp_douyin/biz/model/douyin/interact"
	"github.com/linzijie1998/bytedance_camp_douyin/model"
	"github.com/linzijie1998/bytedance_camp_douyin/util"
)

const (
	commentActionPublish = 1
	commentActionDelete  = 2
)

// CommentAction .
// @router /douyin/comment/action/ [POST]
func CommentAction(ctx context.Context, c *app.RequestContext) {
	var err error
	var req interact.CommentActionReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(interact.CommentActionResp)

	rawID, exists := c.Get("token_user_id")
	if !exists {
		// 未找到user_id
		global.DOUYIN_LOGGER.Debug("未从请求上下文中解析到USERID")
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}
	userID := int64(rawID.(uint))

	if req.ActionType == commentActionPublish {
		// 发布评论
		// 1. 创建评论实例
		// 2. 添加评论数据
		// 3. 更新评论计数
		if req.CommentText == nil || *req.CommentText == "" {
			global.DOUYIN_LOGGER.Debug("未接收到评论信息")
			resp.StatusCode = 1
			c.JSON(consts.StatusBadRequest, resp)
			return
		}
		comment := model.Comment{
			UserInfoID:  userID,
			VideoID:     req.VideoID,
			Content:     *req.CommentText,
			PublishDate: util.GetSysDatetime(),
		}

		if err := dal.CreateComment(&comment); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("评论信息添加失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}

		if err := cache.UpdateCommentCount(req.VideoID, true); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("评论计数添加失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}

		var user = new(base.User)
		user.ID = comment.UserInfoID
		if err = douyin.UserInfoSupplement(userID, user, nil); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("用户信息补充失败 err:%v", err))
			return
		}

		var com = new(interact.Comment)
		com.User = user
		com.ID = int64(comment.ID)
		com.CreateDate = comment.PublishDate
		com.Content = comment.Content

		resp.Comment = com

	} else if req.ActionType == commentActionDelete {
		// 删除评论
		if req.CommentID == nil {
			resp.StatusCode = 1
			c.JSON(consts.StatusBadRequest, resp)
			return
		}
		if err := dal.DeleteCommentByID(*req.CommentID); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("评论信息删除失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		if err := cache.UpdateCommentCount(req.VideoID, false); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("评论计数更新失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
	} else {
		global.DOUYIN_LOGGER.Info(fmt.Sprintf("错误的评论操作 action_type: %d", req.ActionType))
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// CommentList .
// @router /douyin/comment/list/ [GET]
func CommentList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req interact.CommentListReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(interact.CommentListResp)
	var userID int64
	// 登录状态下查看用户信息
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

	comments, err := dal.QueryCommentByVideoID(req.VideoID)
	if err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("评论信息查询失败 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}
	if len(comments) != 0 {
		commentList := make([]*interact.Comment, len(comments))
		for i, comment := range comments {

			var user = new(base.User)
			user.ID = comment.UserInfoID
			if err = douyin.UserInfoSupplement(userID, user, nil); err != nil {
				global.DOUYIN_LOGGER.Debug(fmt.Sprintf("用户信息补充失败 err:%v", err))
				return
			}

			var com = new(interact.Comment)
			com.User = user
			com.ID = int64(comment.ID)
			com.CreateDate = comment.PublishDate
			com.Content = comment.Content
			commentList[i] = com

		}
		resp.CommentList = commentList
	}

	c.JSON(consts.StatusOK, resp)
}
