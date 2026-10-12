package app

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"wells-risk-backend/docs"
	"wells-risk-backend/internal/app/ds"
	"wells-risk-backend/internal/app/role"
)

func (a *Application) StartServer() {
	log.Println("Server start up")

	r := gin.Default()

	// Swagger отправляет запросы на тот порт, на котором запущен сервис.
	if port := os.Getenv("PORT"); port != "" {
		docs.SwaggerInfo.Host = "localhost:" + port
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	h := a.handler

	api := r.Group("/api")
	{
		criteria := api.Group("/criteria")
		{
			// чтение доступно гостю, с токеном врача заполняются is_mine и is_liked
			criteria.GET("", a.WithOptionalAuth, h.GetCriteria)
			criteria.GET("/feed", a.WithOptionalAuth, h.GetCriterionFeed)

			// методы создателя — только с авторизацией
			criteria.GET("/draft", a.WithAuthCheck(role.Physician, role.Moderator), h.GetCriterionDraft)
			criteria.POST("", a.WithAuthCheck(role.Physician, role.Moderator), h.CreateCriterion)
			criteria.PUT("/:id/publish", a.WithAuthCheck(role.Physician, role.Moderator), h.PublishCriterion)
			criteria.DELETE("/:id", a.WithAuthCheck(role.Physician, role.Moderator), h.DeleteCriterion)
			criteria.POST("/:id/like", a.WithAuthCheck(role.Physician, role.Moderator), h.LikeCriterion)
		}

		physicians := api.Group("/physicians")
		{
			physicians.POST("/register", a.Register)
			physicians.POST("/login", a.Login)
			physicians.POST("/logout", a.WithAuthCheck(role.Physician, role.Moderator), a.Logout)
		}
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}

type loginReq struct {
	Login    string `json:"login" example:"kuznetsov"`
	Password string `json:"password" example:"secret123"`
}

type loginResp struct {
	ExpiresIn   time.Duration `json:"expires_in" swaggertype:"integer" example:"3600000000000"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
}

// Login godoc
// @Summary      Аутентификация врача
// @Description  Проверка логина и пароля, в ответе JWT для заголовка Authorization: Bearer <access_token>
// @Tags         Врачи
// @Accept       json
// @Produce      json
// @Param        body  body      loginReq  true  "логин и пароль"
// @Success      200   {object}  loginResp
// @Failure      400
// @Failure      403
// @Failure      500
// @Router       /api/physicians/login [post]
func (a *Application) Login(gCtx *gin.Context) {
	cfg := a.config
	req := &loginReq{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	physician, err := a.repo.GetPhysicianByLogin(req.Login)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if req.Login == physician.Login && physician.Password == generateHashString(req.Password) {
		// значит проверка пройдена
		// генерируем ему jwt
		token := jwt.NewWithClaims(cfg.JWT.SigningMethod, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(cfg.JWT.ExpiresIn).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "wells-admin",
			},
			UserUUID: physician.PhysicianUUID,
			Role:     physician.Role,
		})
		if token == nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
			return
		}

		strToken, err := token.SignedString([]byte(cfg.JWT.Token))
		if err != nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token"))
			return
		}

		gCtx.JSON(http.StatusOK, loginResp{
			ExpiresIn:   cfg.JWT.ExpiresIn,
			AccessToken: strToken,
			TokenType:   "Bearer",
		})

		return
	}

	gCtx.AbortWithStatus(http.StatusForbidden) // отдаем 403 ответ в знак того что доступ запрещен
}

type registerReq struct {
	Login    string `json:"login" example:"kuznetsov"`
	Password string `json:"password" example:"secret123"`
	FullName string `json:"full_name" example:"Кузнецов Андрей Николаевич"`
}

type registerResp struct {
	Ok bool `json:"ok"`
}

// Register godoc
// @Summary      Регистрация врача
// @Description  Новый врач с ролью создателя, пароль хранится в виде sha1-хеша
// @Tags         Врачи
// @Accept       json
// @Produce      json
// @Param        body  body      registerReq  true  "данные врача"
// @Success      200   {object}  registerResp
// @Failure      400
// @Failure      409
// @Failure      500
// @Router       /api/physicians/register [post]
func (a *Application) Register(gCtx *gin.Context) {
	req := &registerReq{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Password == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("password is empty"))
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("login is empty"))
		return
	}

	exists, err := a.repo.PhysicianExists(req.Login)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if exists {
		gCtx.AbortWithError(http.StatusConflict, fmt.Errorf("login is taken"))
		return
	}

	err = a.repo.RegisterPhysician(&ds.Physician{
		PhysicianUUID: uuid.New(),
		Role:          role.Physician,
		Login:         req.Login,
		FullName:      strings.TrimSpace(req.FullName),
		Password:      generateHashString(req.Password), // пароли делаем в хешированном виде и далее будем сравнивать хеши, чтобы их не угнали с базой вместе
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	gCtx.JSON(http.StatusOK, &registerResp{
		Ok: true,
	})
}

func generateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// Logout godoc
// @Summary      Деавторизация врача
// @Description  Токен из заголовка Authorization попадает в blacklist Redis до истечения срока действия
// @Tags         Врачи
// @Security     ApiKeyAuth
// @Success      200
// @Failure      400
// @Failure      403
// @Failure      500
// @Router       /api/physicians/logout [post]
func (a *Application) Logout(gCtx *gin.Context) {
	// получаем заголовок
	jwtStr := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) { // если нет префикса то нас дурят!
		gCtx.AbortWithStatus(http.StatusBadRequest) // отдаем что нет доступа

		return // завершаем обработку
	}

	// отрезаем префикс
	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.config.JWT.Token), nil
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)

		return
	}

	// сохраняем в блеклист редиса
	err = a.redis.WriteJWTToBlacklist(gCtx.Request.Context(), jwtStr, a.config.JWT.ExpiresIn)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)

		return
	}

	gCtx.Status(http.StatusOK)
}
