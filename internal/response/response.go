package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 200, Msg: "操作成功", Data: data})
}

func OKMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 200, Msg: msg, Data: data})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Body{Code: code, Msg: msg, Data: nil})
}

func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	OK(c, PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}
