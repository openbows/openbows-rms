package member

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	membersvc "openbows-rms/internal/service/member"
)

type Controller struct {
	svc *membersvc.Service
}

func New(svc *membersvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (ctl *Controller) op(c *gin.Context) membersvc.Operator {
	return membersvc.Operator{UserID: middleware.GetUserID(c)}
}

func (ctl *Controller) Create(c *gin.Context) {
	var in membersvc.CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.Create(c.Request.Context(), middleware.GetStoreID(c), ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in membersvc.UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	m, err := ctl.svc.Update(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *Controller) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	if err := ctl.svc.Delete(c.Request.Context(), middleware.GetStoreID(c), id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

func (ctl *Controller) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.Get(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	var status *bool
	if s := c.Query("status"); s != "" {
		v := s == "true" || s == "1"
		status = &v
	}
	list, total, err := ctl.svc.List(c.Request.Context(), middleware.GetStoreID(c), keyword, status, page, pageSize)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	response.Page(c, list, total, page, pageSize)
}

func (ctl *Controller) Recharge(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in membersvc.AmountInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.Recharge(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) Consume(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in membersvc.AmountInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.Consume(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) EarnPoints(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in membersvc.PointsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.EarnPoints(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) SpendPoints(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in membersvc.PointsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.SpendPoints(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) ListWalletLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := ctl.svc.ListWalletLogs(c.Request.Context(), middleware.GetStoreID(c), id, page, pageSize)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	response.Page(c, list, total, page, pageSize)
}

func (ctl *Controller) ListPointLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := ctl.svc.ListPointLogs(c.Request.Context(), middleware.GetStoreID(c), id, page, pageSize)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	response.Page(c, list, total, page, pageSize)
}
