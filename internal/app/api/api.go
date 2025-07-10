package api

import (
	"reflect"

	"github.com/lzh-1625/go_process_manager/internal/app/constants"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"

	"github.com/gin-gonic/gin"
)

func getRole(ctx *gin.Context) constants.Role {
	if v, ok := ctx.Get(constants.CTXFLG_ROLE); ok {
		return v.(constants.Role)
	}
	return constants.ROLE_GUEST
}

func getUserName(ctx *gin.Context) string {
	return ctx.GetString(constants.CTXFLG_USER_NAME)
}

func isAdmin(ctx *gin.Context) bool {
	return getRole(ctx) <= constants.ROLE_ADMIN
}

func hasOprPermission(ctx *gin.Context, uuid int, op constants.OprPermission) bool {
	return isAdmin(ctx) || reflect.ValueOf(repository.PermissionRepository.GetPermission(getUserName(ctx), uuid)).FieldByName(string(op)).Bool()
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
	} else {
		if err, ok := msg.(error); ok {
			r.Msg = err.Error()
		}
	}
	return r
}
