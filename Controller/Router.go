package Controller

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type Gin struct {
	eng *gin.Engine
}

func setGin() *Gin {
	gin.SetMode(gin.ReleaseMode) //設定Gin的模式為DebugMode/ReleaseMode/TestMode
	r := gin.Default()
	// 初始化session
	store := cookie.NewStore([]byte("secret")) //session key
	store.Options(sessions.Options{
		Path:     "/Member", //需設定路徑，避免session儲存在不同路徑無法讀取
		MaxAge:   int(24 * time.Hour / time.Second),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("mysession", store)) //session名稱
	return &Gin{r}
}

// 將所有函式集中在該func裡，並寫到main執行

func Router() *gin.Engine {
	r := setGin()
	r.eng.Use()
	{
		//r.eng.GET("/ping", PingWeb)
		r.eng.GET("/Member/header", HeaderWeb)
		r.eng.GET("/Member/Index", IndexWeb)
		r.eng.GET("/Member/Login", LoginWeb)
		r.eng.GET("/Member/Register", RegisterWeb)
		r.eng.GET("/Member/PasswordforgetSend", PasswordforgetSendWeb)
		r.eng.GET("/Member/PasswordUpdate/:token", PasswordUpdateWeb)
		r.eng.GET("/Member/auth", CallGoogle)
		r.eng.GET("/Member/auth/callback", CallbackGoogle)
		r.eng.GET("/Member/Profile-Setup", ProfileSetupWeb)
		r.eng.GET("/Member/footer", FooterWeb)
		r.eng.GET("/Member/status", LoginCheck)
		r.eng.GET("/Member/Logout", Logout)
		r.eng.GET("/Member/EmailVerify/:token", EmailVerifyCheck)

		r.eng.POST("/Member/Register", MemberRegister)
		r.eng.POST("/Member/Login", MemberLogin)
		r.eng.POST("/Member/Profile-Setup", ProfileSetup)
		r.eng.POST("/Member/PasswordforgetSend", PasswordforgetSend)
		r.eng.POST("/Member/PasswordUpdate/:token", PasswordUpdate)
	}
	return r.eng
}
