package middle

import (
	"github.com/lzh-1625/go_process_manager/internal/app/eum"

	"github.com/gin-gonic/gin"
)

func RolePermission(needPermission eum.Role) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if v, ok := ctx.Get(eum.CtxRole); !ok || v.(eum.Role) > needPermission {
			rErr(ctx, -1, "Insufficient permissions; please check your access rights!", nil)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
