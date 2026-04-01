package controller

import (
	"context"
	"time"

	"microg/examples/user/client/service"
	"microg/pkg/common/core"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	svc service.UserService
}

func NewUserController(svc service.UserService) *UserController {
	return &UserController{svc: svc}
}

func (c *UserController) List(ctx *gin.Context) {
	cxt, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	list, err := c.svc.List(cxt, 1, 10)
	core.WriteResponse(ctx, err, list)
}

func (c *UserController) GetByID(ctx *gin.Context) {
	var req struct {
		ID int32 `uri:"id" binding:"required"`
	}
	if err := ctx.ShouldBindUri(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	cxt, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := c.svc.GetByID(cxt, req.ID)
	core.WriteResponse(ctx, err, user)
}

func (c *UserController) Create(ctx *gin.Context) {
	var req struct {
		NickName string `json:"nickName" binding:"required"`
		PassWord string `json:"passWord" binding:"required"`
		Mobile   string `json:"mobile" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}

	cxt, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := c.svc.Create(cxt, req.NickName, req.PassWord, req.Mobile)
	core.WriteResponse(ctx, err, user)
}

func (c *UserController) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("", c.List)
	r.GET("/:id", c.GetByID)
	r.POST("", c.Create)
}
