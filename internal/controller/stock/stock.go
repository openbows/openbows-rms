package stock

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	stocksvc "openbows-rms/internal/service/stock"
)

type Controller struct {
	svc *stocksvc.Service
}

func New(svc *stocksvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (ctl *Controller) op(c *gin.Context) stocksvc.Operator {
	return stocksvc.Operator{UserID: middleware.GetUserID(c)}
}

func (ctl *Controller) CreateMaterial(c *gin.Context) {
	var in stocksvc.MaterialInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	m, err := ctl.svc.CreateMaterial(c.Request.Context(), middleware.GetStoreID(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *Controller) UpdateMaterial(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in stocksvc.MaterialInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	m, err := ctl.svc.UpdateMaterial(c.Request.Context(), middleware.GetStoreID(c), id, in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *Controller) DeleteMaterial(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	if err := ctl.svc.DeleteMaterial(c.Request.Context(), middleware.GetStoreID(c), id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

func (ctl *Controller) GetMaterial(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	m, err := ctl.svc.GetMaterial(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *Controller) ListMaterials(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	var status *bool
	if s := c.Query("status"); s != "" {
		v := s == "true" || s == "1"
		status = &v
	}
	list, total, err := ctl.svc.ListMaterials(c.Request.Context(), middleware.GetStoreID(c), c.Query("keyword"), status, page, pageSize)
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

func (ctl *Controller) ListBalances(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	materialID, _ := strconv.ParseInt(c.Query("material_id"), 10, 64)
	list, total, err := ctl.svc.ListBalances(c.Request.Context(), middleware.GetStoreID(c), materialID, page, pageSize)
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

func (ctl *Controller) CreateDoc(c *gin.Context) {
	var in stocksvc.CreateDocInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.CreateDoc(c.Request.Context(), middleware.GetStoreID(c), ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) GetDoc(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.GetDoc(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) ListDocs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	list, total, err := ctl.svc.ListDocs(c.Request.Context(), middleware.GetStoreID(c), c.Query("doc_type"), c.Query("status"), page, pageSize)
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

func (ctl *Controller) PostDoc(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.Post(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) VoidDoc(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in stocksvc.VoidInput
	_ = c.ShouldBindJSON(&in)
	d, err := ctl.svc.Void(c.Request.Context(), middleware.GetStoreID(c), id, ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}
