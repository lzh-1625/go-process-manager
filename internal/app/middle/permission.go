package middle

import (
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/log"

	"github.com/gin-gonic/gin"
)

func RolePermission(needPermission eum.Role) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if v, ok := ctx.Get(eum.CtxRole); !ok || v.(eum.Role) > needPermission {
			log.Logger.Errorw("Insufficient permissions", "needPermission", needPermission, "role", v)
			rErr(ctx, -1, "Insufficient permissions", nil)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
