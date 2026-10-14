package menu

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	menusvc "openbows-rms/internal/service/menu"
)

type DishController struct {
	svc *menusvc.DishService
}

func NewDishController(svc *menusvc.DishService) *DishController {
	return &DishController{svc: svc}
}

func (ctl *DishController) Create(c *gin.Context) {
	var in menusvc.DishInput
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

func (ctl *DishController) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in menusvc.DishInput
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

func (ctl *DishController) Delete(c *gin.Context) {
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

func (ctl *DishController) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	m, err := ctl.svc.GetDetail(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *DishController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	categoryID, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	stationID, _ := strconv.ParseInt(c.Query("station_id"), 10, 64)
	keyword := c.Query("keyword")
	var isOnSale *bool
	if v := c.Query("is_on_sale"); v == "true" || v == "1" {
		b := true
		isOnSale = &b
	} else if v == "false" || v == "0" {
		b := false
		isOnSale = &b
	}
	list, total, err := ctl.svc.List(c.Request.Context(), middleware.GetStoreID(c), categoryID, stationID, isOnSale, keyword, page, pageSize)
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

func (ctl *DishController) ReplaceSpecs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in menusvc.ReplaceSpecsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	list, err := ctl.svc.ReplaceSpecs(c.Request.Context(), middleware.GetStoreID(c), id, in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, list)
}

func (ctl *DishController) ReplaceOptions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in menusvc.ReplaceOptionsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	list, err := ctl.svc.ReplaceOptions(c.Request.Context(), middleware.GetStoreID(c), id, in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, list)
}

func (ctl *DishController) ReplaceMaterials(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in menusvc.ReplaceMaterialsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	list, err := ctl.svc.ReplaceMaterials(c.Request.Context(), middleware.GetStoreID(c), id, in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, list)
}
