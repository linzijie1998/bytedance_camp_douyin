// Code generated by hertz generator.

package relation

import (
	"context"
	"fmt"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/cache"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/dal"
	"github.com/linzijie1998/bytedance_camp_douyin/biz/model/douyin/base"
	"github.com/linzijie1998/bytedance_camp_douyin/global"
	"github.com/linzijie1998/bytedance_camp_douyin/util"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	relation "github.com/linzijie1998/bytedance_camp_douyin/biz/model/douyin/relation"
)

const (
	relationActionFollow = 1
	relationActionCancel = 2
)

// RelationAction .
// @router /douyin/relation/action/ [POST]
func RelationAction(ctx context.Context, c *app.RequestContext) {
	// 1. 解析Token，获取UserID
	// 2. 检查userID和toUserID的关注状态
	// 3. 根据关注状态和actionType进行不同的逻辑处理
	var err error
	var req relation.RelationActionReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	resp := new(relation.RelationActionResp)

	if req.Token == "" {
		global.DOUYIN_LOGGER.Debug("未携带Token")
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}
	j := util.NewJWT()
	claims, err := j.ParseToken(req.Token)
	if err != nil {
		global.DOUYIN_LOGGER.Debug("Token解析错误")
		resp.StatusCode = 1
		return
	}
	userID := int64(claims.UserInfo.ID)

	// 用户不能对自己进行关注或者取关操作
	if userID == req.ToUserID {
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	// 查询是否关注了该用户
	isFollow, err := cache.GetFollowState(userID, req.ToUserID)
	if err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关系数据查询失败 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	if req.ActionType == relationActionFollow {
		if isFollow {
			global.DOUYIN_LOGGER.Info(fmt.Sprintf("ID为%d的用户尝试重复关注ID为%d的用户", userID, req.ToUserID))
			resp.StatusCode = 1
			c.JSON(consts.StatusBadRequest, resp)
			return
		}
		// 此处需要更新多个key, 可能出现数据不一致的情况
		if err := cache.UpdateFollowState(userID, req.ToUserID, true); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关系数据更新失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		if err := cache.UpdateFollowerState(req.ToUserID, userID, true); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关系数据更新失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		if err := cache.UpdateFollowCount(userID, true); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关系数据更新失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		if err := cache.UpdateFollowerCount(req.ToUserID, true); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关系数据更新失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
	} else if req.ActionType == relationActionCancel {
		if !isFollow {
			global.DOUYIN_LOGGER.Info(fmt.Sprintf("ID为%d的用户尝试取消关注未关注的ID为%d的用户", userID, req.ToUserID))
			resp.StatusCode = 1
			c.JSON(consts.StatusBadRequest, resp)
			return
		}
		if err := cache.UpdateFollowState(userID, req.ToUserID, false); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关系数据更新失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		if err := cache.UpdateFollowerState(req.ToUserID, userID, false); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关系数据更新失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		if err := cache.UpdateFollowCount(userID, false); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关系数据更新失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		if err := cache.UpdateFollowerCount(req.ToUserID, false); err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关系数据更新失败 err: %v", err))
			resp.StatusCode = 1
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
	} else {
		global.DOUYIN_LOGGER.Info(fmt.Sprintf("错误的关系操作 action_type: %d", req.ActionType))
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}

	c.JSON(consts.StatusOK, resp)
}

// RelationFollowList .
// @router /douyin/relation/follow/list/ [GET]
func RelationFollowList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req relation.RelationFollowListReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(relation.RelationFollowListResp)

	if req.Token == "" {
		c.JSON(consts.StatusOK, resp)
		return
	}
	j := util.NewJWT()
	claim, err := j.ParseToken(req.Token)
	if err != nil {
		global.DOUYIN_LOGGER.Info(fmt.Sprintf("Token解析失败 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}
	userID := int64(claim.UserInfo.ID)

	followIDs, err := cache.QueryFollowByUserID(userID)
	if err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("关注者数据查询失败 err: %v", err))
	}

	userList := make([]*base.User, len(followIDs))

	for i, followID := range followIDs {
		userInfos, err := dal.QueryUserInfoByUserID(followID)
		if err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("查询用户信息失败: %v", err))
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		if len(userInfos) != 1 {
			global.DOUYIN_LOGGER.Warn(fmt.Sprintf("查询到%d条的ID为%d的用户信息", len(userInfos), req.UserID))
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}

		followCnt, _ := cache.GetFollowCount(int64(userInfos[0].ID))
		followerCnt, _ := cache.GetFollowerCount(int64(userInfos[0].ID))
		isFollow, _ := cache.GetFollowState(userID, int64(userInfos[0].ID))

		var user base.User
		user.ID = int64(userInfos[0].ID)
		user.Name = userInfos[0].Name
		user.FollowCount = &followCnt
		user.FollowerCount = &followerCnt
		user.IsFollow = isFollow
		userList[i] = &user
	}

	resp.UserList = userList
	c.JSON(consts.StatusOK, resp)
}

// RelationFollowerList .
// @router /douyin/relation/follower/list/ [GET]
func RelationFollowerList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req relation.RelationFollowerListReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(relation.RelationFollowerListResp)
	if req.Token == "" {
		c.JSON(consts.StatusOK, resp)
		return
	}
	j := util.NewJWT()
	claim, err := j.ParseToken(req.Token)
	if err != nil {
		global.DOUYIN_LOGGER.Info(fmt.Sprintf("Token解析失败 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}
	userID := int64(claim.UserInfo.ID)

	followIDs, err := cache.QueryFollowerByUserID(userID)
	if err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("粉丝数据查询失败 err: %v", err))
	}

	userList := make([]*base.User, len(followIDs))

	for i, followID := range followIDs {
		userInfos, err := dal.QueryUserInfoByUserID(followID)
		if err != nil {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("查询用户信息失败: %v", err))
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		if len(userInfos) != 1 {
			global.DOUYIN_LOGGER.Warn(fmt.Sprintf("查询到%d条的ID为%d的用户信息", len(userInfos), req.UserID))
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}

		followCnt, _ := cache.GetFollowCount(int64(userInfos[0].ID))
		followerCnt, _ := cache.GetFollowerCount(int64(userInfos[0].ID))
		isFollow, _ := cache.GetFollowState(userID, int64(userInfos[0].ID))

		var user base.User
		user.ID = int64(userInfos[0].ID)
		user.Name = userInfos[0].Name
		user.FollowCount = &followCnt
		user.FollowerCount = &followerCnt
		user.IsFollow = isFollow
		userList[i] = &user
	}

	resp.UserList = userList
	c.JSON(consts.StatusOK, resp)
}

// RelationFriendList .
// @router /douyin/relation/friend/list/ [GET]
func RelationFriendList(ctx context.Context, c *app.RequestContext) {
	var err error
	var req relation.RelationFriendListReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}

	resp := new(relation.RelationFriendListResp)

	if req.Token == "" {
		c.JSON(consts.StatusOK, resp)
		return
	}
	j := util.NewJWT()
	claim, err := j.ParseToken(req.Token)
	if err != nil {
		global.DOUYIN_LOGGER.Info(fmt.Sprintf("Token解析失败 err: %v", err))
		resp.StatusCode = 1
		c.JSON(consts.StatusBadRequest, resp)
		return
	}
	userID := int64(claim.UserInfo.ID)

	followUserIDs, err := cache.QueryFollowByUserID(userID)
	if err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("查询关注者信息失败: %v", err))
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}
	followerUserIDs, err := cache.QueryFollowerByUserID(userID)
	if err != nil {
		global.DOUYIN_LOGGER.Debug(fmt.Sprintf("查询粉丝信息失败: %v", err))
		c.JSON(consts.StatusInternalServerError, resp)
		return
	}

	// 求并集
	friendIDs := util.GetIntersection(followUserIDs, followerUserIDs)
	for _, id := range friendIDs {
		userInfos, err := dal.QueryUserInfoByUserID(id)
		if err != nil || len(userInfos) != 1 {
			global.DOUYIN_LOGGER.Debug(fmt.Sprintf("查询用户信息失败: %v", err))
			c.JSON(consts.StatusInternalServerError, resp)
			return
		}
		friend := new(relation.FriendUser)
		friend.ID = int64(userInfos[0].ID)
		friend.IsFollow = true
	}

	c.JSON(consts.StatusOK, resp)
}
