package order

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	ordersvc "openbows-rms/internal/service/order"
)

type Controller struct {
	svc *ordersvc.Service
}

func New(svc *ordersvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (ctl *Controller) op(c *gin.Context) ordersvc.Operator {
	return ordersvc.Operator{
		UserID:   middleware.GetUserID(c),
		Username: middleware.GetUsername(c),
	}
}

func (ctl *Controller) Create(c *gin.Context) {
	var in ordersvc.CreateInput
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

func (ctl *Controller) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	tableID, _ := strconv.ParseInt(c.Query("table_id"), 10, 64)
	status := c.Query("status")
	list, total, err := ctl.svc.List(c.Request.Context(), middleware.GetStoreID(c), tableID, status, page, pageSize)
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

func (ctl *Controller) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.GetDetail(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) AddItems(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in ordersvc.AddItemsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.AddItems(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) ReturnItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	itemID, err := strconv.ParseInt(c.Param("item_id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效明细ID")
		return
	}
	var in ordersvc.ReturnItemInput
	_ = c.ShouldBindJSON(&in)
	d, err := ctl.svc.ReturnItem(c.Request.Context(), middleware.GetStoreID(c), id, itemID, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) Advance(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in ordersvc.AdvanceInput
	_ = c.ShouldBindJSON(&in)
	d, err := ctl.svc.Advance(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in ordersvc.CancelInput
	_ = c.ShouldBindJSON(&in)
	d, err := ctl.svc.Cancel(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}
