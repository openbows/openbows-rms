package hr

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	hrsvc "openbows-rms/internal/service/hr"
)

type Controller struct {
	svc *hrsvc.Service
}

func New(svc *hrsvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (ctl *Controller) op(c *gin.Context) hrsvc.Operator {
	return hrsvc.Operator{UserID: middleware.GetUserID(c)}
}

// ---------- employees ----------

func (ctl *Controller) CreateEmployee(c *gin.Context) {
	var in hrsvc.EmployeeCreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.CreateEmployee(c.Request.Context(), middleware.GetStoreID(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) UpdateEmployee(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in hrsvc.EmployeeUpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.UpdateEmployee(c.Request.Context(), middleware.GetStoreID(c), id, in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) DeleteEmployee(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	if err := ctl.svc.DeleteEmployee(c.Request.Context(), middleware.GetStoreID(c), id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

func (ctl *Controller) GetEmployee(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.GetEmployee(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) ListEmployees(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	var status *bool
	if s := c.Query("status"); s != "" {
		v := s == "true" || s == "1"
		status = &v
	}
	list, total, err := ctl.svc.ListEmployees(c.Request.Context(), middleware.GetStoreID(c), keyword, status, page, pageSize)
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

// ---------- shifts ----------

func (ctl *Controller) CreateShift(c *gin.Context) {
	var in hrsvc.ShiftCreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.CreateShift(c.Request.Context(), middleware.GetStoreID(c), ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) UpdateShift(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	var in hrsvc.ShiftUpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.UpdateShift(c.Request.Context(), middleware.GetStoreID(c), id, in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) CancelShift(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.CancelShift(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) DeleteShift(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	if err := ctl.svc.DeleteShift(c.Request.Context(), middleware.GetStoreID(c), id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

func (ctl *Controller) GetShift(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.GetShift(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) ListShifts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	employeeID, _ := strconv.ParseInt(c.Query("employee_id"), 10, 64)
	list, total, err := ctl.svc.ListShifts(
		c.Request.Context(), middleware.GetStoreID(c), employeeID,
		c.Query("shift_date"), c.Query("status"), page, pageSize,
	)
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

// ---------- attendances ----------

func (ctl *Controller) Punch(c *gin.Context) {
	var in hrsvc.PunchInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.Punch(c.Request.Context(), middleware.GetStoreID(c), ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) Makeup(c *gin.Context) {
	var in hrsvc.MakeupInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	d, err := ctl.svc.Makeup(c.Request.Context(), middleware.GetStoreID(c), ctl.op(c), in)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}

func (ctl *Controller) ListAttendances(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	employeeID, _ := strconv.ParseInt(c.Query("employee_id"), 10, 64)
	list, total, err := ctl.svc.ListAttendances(
		c.Request.Context(), middleware.GetStoreID(c), employeeID, c.Query("work_date"), page, pageSize,
	)
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

func (ctl *Controller) GetAttendance(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	d, err := ctl.svc.GetAttendance(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, d)
}
