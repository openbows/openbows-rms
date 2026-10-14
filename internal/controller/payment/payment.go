package payment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	ordersvc "openbows-rms/internal/service/order"
	paymentsvc "openbows-rms/internal/service/payment"
)

type Controller struct {
	svc *paymentsvc.Service
}

func New(svc *paymentsvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (ctl *Controller) op(c *gin.Context) ordersvc.Operator {
	return ordersvc.Operator{
		UserID:   middleware.GetUserID(c),
		Username: middleware.GetUsername(c),
	}
}

func (ctl *Controller) Pay(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in paymentsvc.PayInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	res, err := ctl.svc.Pay(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, res)
}

func (ctl *Controller) List(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	list, err := ctl.svc.ListByOrder(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, list)
}
