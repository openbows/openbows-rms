package online

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/response"
	onlinesvc "openbows-rms/internal/service/online"
)

/*
顾客端接口鉴权说明：
- POST/GET /apis/v1/online/orders* 不校验员工 JWT
- 下单靠 store_id+table_id；查单靠 access_token（query 或 X-Order-Access-Token）
*/

type Controller struct {
	svc *onlinesvc.Service
}

func New(svc *onlinesvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (ctl *Controller) Create(c *gin.Context) {
	var in onlinesvc.CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	res, err := ctl.svc.Create(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, res)
}

func (ctl *Controller) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	token := c.Query("access_token")
	if token == "" {
		token = c.GetHeader("X-Order-Access-Token")
	}
	res, err := ctl.svc.GetProgress(c.Request.Context(), id, token)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, res)
}
