package kds

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"openbows-rms/internal/middleware"
	"openbows-rms/internal/response"
	kdssvc "openbows-rms/internal/service/kds"
)

type Controller struct {
	svc *kdssvc.Service
}

func New(svc *kdssvc.Service) *Controller {
	return &Controller{svc: svc}
}

func (ctl *Controller) List(c *gin.Context) {
	stationID, _ := strconv.ParseInt(c.Query("station_id"), 10, 64)
	status := c.Query("status")
	list, err := ctl.svc.List(c.Request.Context(), middleware.GetStoreID(c), stationID, status)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, list)
}

func (ctl *Controller) Start(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	v, err := ctl.svc.Start(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, v)
}

func (ctl *Controller) Done(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, "无效ID")
		return
	}
	v, err := ctl.svc.Done(c.Request.Context(), middleware.GetStoreID(c), id)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, v)
}
