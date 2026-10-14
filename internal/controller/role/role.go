package role

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	rolesvc "openbows-rms/internal/service/role"
)

type Controller struct {
	svc *rolesvc.Service
}

func New(svc *rolesvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (ctl *Controller) Create(c *gin.Context) {
	var in rolesvc.CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if in.StoreID == 0 {
		in.StoreID = middleware.GetStoreID(c)
	}
	if in.StoreID != middleware.GetStoreID(c) {
		response.Fail(c, 403, "无权在其他门店创建角色")
		return
	}
	m, err := ctl.svc.Create(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *Controller) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in rolesvc.UpdateInput
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
	m, err := ctl.svc.Get(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (ctl *Controller) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	list, total, err := ctl.svc.List(c.Request.Context(), middleware.GetStoreID(c), page, pageSize, keyword)
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

func (ctl *Controller) AssignPermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in rolesvc.AssignPermissionsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := ctl.svc.AssignPermissions(c.Request.Context(), middleware.GetStoreID(c), id, in.PermissionIDs); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

func (ctl *Controller) ListPermissionsOfRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	ids, err := ctl.svc.ListPermissionIDs(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"permission_ids": ids})
}

func (ctl *Controller) ListAllPermissions(c *gin.Context) {
	list, err := ctl.svc.ListAllPermissions(c.Request.Context())
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, list)
}
