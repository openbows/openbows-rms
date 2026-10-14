package purchase

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	purchasesvc "openbows-rms/internal/service/purchase"
)

type Controller struct {
	svc *purchasesvc.Service
}

func New(svc *purchasesvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (ctl *Controller) op(c *gin.Context) purchasesvc.Operator {
	return purchasesvc.Operator{UserID: middleware.GetUserID(c)}
}

func (ctl *Controller) CreateSupplier(c *gin.Context) {
	var in purchasesvc.SupplierInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	m, err := ctl.svc.CreateSupplier(c.Request.Context(), middleware.GetStoreID(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *Controller) UpdateSupplier(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in purchasesvc.SupplierInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	m, err := ctl.svc.UpdateSupplier(c.Request.Context(), middleware.GetStoreID(c), id, in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *Controller) DeleteSupplier(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	if err := ctl.svc.DeleteSupplier(c.Request.Context(), middleware.GetStoreID(c), id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

func (ctl *Controller) GetSupplier(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	m, err := ctl.svc.GetSupplier(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *Controller) ListSuppliers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	var status *bool
	if s := c.Query("status"); s != "" {
		v := s == "true" || s == "1"
		status = &v
	}
	list, total, err := ctl.svc.ListSuppliers(c.Request.Context(), middleware.GetStoreID(c), c.Query("keyword"), status, page, pageSize)
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

func (ctl *Controller) CreateOrder(c *gin.Context) {
	var in purchasesvc.CreateOrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.CreateOrder(c.Request.Context(), middleware.GetStoreID(c), ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.GetOrder(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	supplierID, _ := strconv.ParseInt(c.Query("supplier_id"), 10, 64)
	list, total, err := ctl.svc.ListOrders(c.Request.Context(), middleware.GetStoreID(c), supplierID, c.Query("status"), page, pageSize)
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

func (ctl *Controller) SubmitOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.SubmitOrder(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) CancelOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.CancelOrder(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) CreateReceipt(c *gin.Context) {
	var in purchasesvc.CreateReceiptInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.CreateReceipt(c.Request.Context(), middleware.GetStoreID(c), ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) GetReceipt(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.GetReceipt(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) ListReceipts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	poID, _ := strconv.ParseInt(c.Query("purchase_order_id"), 10, 64)
	list, total, err := ctl.svc.ListReceipts(c.Request.Context(), middleware.GetStoreID(c), poID, c.Query("status"), page, pageSize)
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

func (ctl *Controller) PostReceipt(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.PostReceipt(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}
