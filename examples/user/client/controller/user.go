package controller

import (
	"context"
	"time"

	"github.com/atliliw/microg/examples/user/client/service"
	"github.com/atliliw/microg/pkg/common/core"
	"github.com/atliliw/microg/pkg/log"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	svc service.UserService
}

func NewUserController(svc service.UserService) *UserController {
	return &UserController{svc: svc}
}

func (c *UserController) List(ctx *gin.Context) {
	log.InfoC(ctx.Request.Context(), "HTTP请求: List用户列表", log.String("path", ctx.Request.URL.Path))
	cxt, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	list, err := c.svc.List(cxt, 1, 10)
	if err != nil {
		log.ErrorC(ctx.Request.Context(), "List用户列表失败", log.Err(err))
	}
	core.WriteResponse(ctx, err, list)
}

func (c *UserController) GetByID(ctx *gin.Context) {
	var req struct {
		ID int32 `uri:"id" binding:"required"`
	}
	if err := ctx.ShouldBindUri(&req); err != nil {
		log.WarnC(ctx.Request.Context(), "参数绑定失败", log.Err(err), log.String("path", ctx.Request.URL.Path))
		core.WriteResponse(ctx, err, nil)
		return
	}

	log.InfoC(ctx.Request.Context(), "HTTP请求: GetByID", log.Int32("id", req.ID))
	cxt, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := c.svc.GetByID(cxt, req.ID)
	if err != nil {
		log.ErrorC(ctx.Request.Context(), "GetByID失败", log.Err(err), log.Int32("id", req.ID))
	}
	core.WriteResponse(ctx, err, user)
}

func (c *UserController) Create(ctx *gin.Context) {
	var req struct {
		NickName string `json:"nickName" binding:"required"`
		PassWord string `json:"passWord" binding:"required"`
		Mobile   string `json:"mobile" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.WarnC(ctx.Request.Context(), "参数绑定失败", log.Err(err), log.String("path", ctx.Request.URL.Path))
		core.WriteResponse(ctx, err, nil)
		return
	}

	log.InfoC(ctx.Request.Context(), "HTTP请求: Create用户",
		log.String("nickname", req.NickName),
		log.String("mobile", req.Mobile),
	)
	cxt, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := c.svc.Create(cxt, req.NickName, req.PassWord, req.Mobile)
	if err != nil {
		log.ErrorC(ctx.Request.Context(), "Create用户失败", log.Err(err))
	} else {
		log.InfoC(ctx.Request.Context(), "Create用户成功", log.Int32("id", user.ID))
	}
	core.WriteResponse(ctx, err, user)
}

func (c *UserController) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("", c.List)
	r.GET("/:id", c.GetByID)
	r.POST("", c.Create)
}
