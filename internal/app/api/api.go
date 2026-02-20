package api

import (
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"

	"github.com/gin-gonic/gin"
)

func getRole(ctx *gin.Context) eum.Role {
	if v, ok := ctx.Get(eum.CtxRole); ok {
		return v.(eum.Role)
	}
	return eum.RoleGuest
}

func getUserName(ctx *gin.Context) string {
	return ctx.GetString(eum.CtxUserName)
}

func isAdmin(ctx *gin.Context) bool {
	return getRole(ctx) <= eum.RoleAdmin
}

func hasOprPermission(ctx *gin.Context, uuid int, op eum.OprPermission) bool {
	if isAdmin(ctx) {
		return true
	}
	per := repository.PermissionRepository.GetPermission(getUserName(ctx), uuid)
	if per == nil {
		return false
	}
	switch op {
	case eum.OperationLog:
		return per.Log
	case eum.OperationTerminal:
		return per.Terminal
	case eum.OperationStart:
		return per.Start
	case eum.OperationStop:
		return per.Stop
	case eum.OperationTerminalWrite:
		return per.Write
	}
	return false
}

type Response struct {
	StatusCode int
	Code       int
	Data       any
	Msg        string
}

func NewResponse() *Response {
	return &Response{StatusCode: 200}
}

func (r *Response) SetStatusCode(code int) *Response {
	r.StatusCode = code
	return r
}

func (r *Response) SetDate(data any) *Response {
	r.Data = data
	return r
}

func (r *Response) SetCode(code int) *Response {
	r.Code = code
	return r
}

func (r *Response) SetMessage(msg any) *Response {
	if str, ok := msg.(string); ok {
		r.Msg = str
	} else if err, ok := msg.(error); ok {
		r.Msg = err.Error()
	}
	return r
}
