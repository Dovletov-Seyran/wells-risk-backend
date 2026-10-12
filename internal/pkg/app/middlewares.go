package app

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"

	"wells-risk-backend/internal/app/ds"
	"wells-risk-backend/internal/app/handler"
	"wells-risk-backend/internal/app/role"
)

const jwtPrefix = "Bearer "

func (a *Application) WithAuthCheck(assignedRoles ...role.Role) func(ctx *gin.Context) {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtPrefix) { // если нет префикса то нас дурят!
			gCtx.AbortWithStatus(http.StatusForbidden) // отдаем что нет доступа

			return // завершаем обработку
		}

		// отрезаем префикс
		jwtStr = jwtStr[len(jwtPrefix):]
		// проверяем jwt в блеклист редиса
		err := a.redis.CheckJWTInBlacklist(gCtx.Request.Context(), jwtStr)
		if err == nil { // значит что токен в блеклисте
			gCtx.AbortWithStatus(http.StatusForbidden)

			return
		}
		if !errors.Is(err, redis.Nil) { // значит что это не ошибка отсуствия - внутренняя ошибка
			gCtx.AbortWithError(http.StatusInternalServerError, err)

			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.config.JWT.Token), nil
		})
		if err != nil {
			gCtx.AbortWithStatus(http.StatusForbidden)
			log.Println(err)

			return
		}

		myClaims := token.Claims.(*ds.JWTClaims)

		for _, oneOfAssignedRole := range assignedRoles {
			if myClaims.Role == oneOfAssignedRole {
				// создатель критерия берётся из JWT: кладём ид врача в контекст обработчика
				if !a.setPhysician(gCtx, myClaims) {
					gCtx.AbortWithStatus(http.StatusForbidden)

					return
				}

				gCtx.Next()

				return
			}
		}
		gCtx.AbortWithStatus(http.StatusForbidden)
		log.Printf("role %v is not assigned in %v", myClaims.Role, assignedRoles)

		return

	}

}

// WithOptionalAuth — для методов чтения (список и лента): гостю они доступны без токена,
// а если передан JWT врача, по нему вычисляются признаки is_mine и is_liked.
func (a *Application) WithOptionalAuth(gCtx *gin.Context) {
	jwtStr := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		return
	}

	jwtStr = jwtStr[len(jwtPrefix):]

	// разлогиненный токен — как гость
	if err := a.redis.CheckJWTInBlacklist(gCtx.Request.Context(), jwtStr); !errors.Is(err, redis.Nil) {
		return
	}

	token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.config.JWT.Token), nil
	})
	if err != nil {
		return
	}

	a.setPhysician(gCtx, token.Claims.(*ds.JWTClaims))
}

// setPhysician находит врача по uuid из JWT и кладёт его ид в контекст gin.
func (a *Application) setPhysician(gCtx *gin.Context, claims *ds.JWTClaims) bool {
	physician, err := a.repo.GetPhysicianByUUID(claims.UserUUID)
	if err != nil {
		log.Println(err)

		return false
	}

	gCtx.Set(handler.PhysicianIDKey, physician.PhysicianID)

	return true
}
