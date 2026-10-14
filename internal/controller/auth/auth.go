package auth

import (
	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	authsvc "openbows-rms/internal/service/auth"
)

type Controller struct {
	svc *authsvc.Service
}

func New(svc *authsvc.Service) *Controller {
	return &Controller{svc: svc}
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ctl *Controller) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	res, err := ctl.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, res)
}

func (ctl *Controller) Me(c *gin.Context) {
	res, err := ctl.svc.Me(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, res)
}
