package handler

import (
	"chainmscan/server"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		method := c.Request.Method
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,Content-Type")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

type PageReq struct {
	PageSize int32  `json:"pageSize"`
	Page     int32  `json:"page"`
	SortType string `json:"sortType"`
}

type StandardResp struct {
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type StandardRespWithPage struct {
	StandardResp
	Total int64 `json:"total"`
}

const (
	RespCodeSuccess = 200
	RespCodeFailed  = 500
)

// SuccessfulJSONResp gin成功返回数据包装（不带分页）
func SuccessfulJSONResp(data interface{}, msg string, c *gin.Context) {
	resp := StandardResp{
		Code: RespCodeSuccess,
		Msg:  msg,
		Data: data,
	}
	c.JSON(http.StatusOK, resp)
}

// SuccessfulJSONRespWithPage gin成功返回数据包装（带分页）
func SuccessfulJSONRespWithPage(data interface{}, total int64, c *gin.Context) {
	resp := StandardRespWithPage{
		StandardResp: StandardResp{
			Code: RespCodeFailed,
			Data: data,
		},
		Total: total,
	}
	c.JSON(http.StatusOK, resp)
}

// FailedJSONResp gin失败返回数据包装
func FailedJSONResp(msg string, c *gin.Context) {
	resp := StandardResp{
		Code: RespCodeFailed,
		Msg:  msg,
	}
	c.JSON(http.StatusOK, resp)
}

// Handler gin业务处理器
type Handler interface {
	Handle(s *server.Server) gin.HandlerFunc
}

// TestHandler 测试处理器
type TestHandler struct {
}

func (th *TestHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		SuccessfulJSONResp("Hello,World!", "", ctx)
	}
}

const (
	RespMsgParamsTypeError = "错误的参数类型！"

	RespMsgParamsMissing = "缺少必要参数！"

	RespMsgLogServerError = "日志服务错误！"

	RespMsgServerError = "服务内部错误，请检查日志！"
)

func checkStringParamsEmpty(params ...string) error {
	for _, p := range params {
		if len(p) == 0 {
			err := errors.New("missing parameters")
			return err
		}
	}
	return nil
}

func checkPageReq(p *PageReq) {
	if p.Page <= 0 {
		p.Page = 1
	}

	if p.PageSize <= 0 {
		p.PageSize = 10
	}
}
