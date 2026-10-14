package table

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	tablesvc "openbows-rms/internal/service/table"
)

type AreaController struct {
	svc *tablesvc.AreaService
}

func NewAreaController(svc *tablesvc.AreaService) *AreaController {
	return &AreaController{svc: svc}
}

func (ctl *AreaController) Create(c *gin.Context) {
	var in tablesvc.AreaInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	m, err := ctl.svc.Create(c.Request.Context(), middleware.GetStoreID(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *AreaController) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in tablesvc.AreaInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	m, err := ctl.svc.Update(c.Request.Context(), middleware.GetStoreID(c), id, in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *AreaController) Delete(c *gin.Context) {
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

func (ctl *AreaController) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	m, err := ctl.svc.Get(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *AreaController) List(c *gin.Context) {
	list, err := ctl.svc.List(c.Request.Context(), middleware.GetStoreID(c))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, list)
}
