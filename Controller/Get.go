package Controller

import (
	model "Web_Demo/Model"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	GoogleOAuth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

func PingWeb(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// 讀取頁首
func HeaderWeb(c *gin.Context) {
	c.File("./View/header.html")
}

// 讀取頁尾
func FooterWeb(c *gin.Context) {
	c.File("./View/footer.html")
}

// 首頁頁面
func IndexWeb(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

// 登入頁面
func LoginWeb(c *gin.Context) {
	c.HTML(http.StatusOK, "Login.html", nil)
}

// 註冊頁面
func RegisterWeb(c *gin.Context) {
	c.HTML(http.StatusOK, "Register.html", nil)
}

// 忘記密碼頁面
func PasswordforgetSendWeb(c *gin.Context) {
	c.HTML(http.StatusOK, "PasswordforgetSend.html", nil)
}

// 密碼更新頁面
func PasswordUpdateWeb(c *gin.Context) {
	usertoken := c.Param("token")
	// 根據短token 找到對應的 JWT
	jwtToken, exists := tokenMapping[usertoken]
	if !exists {
		c.JSON(404, nil)
		return
	}

	// 解碼 JWT
	token, _ := jwt.ParseWithClaims(jwtToken.JWTToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		c.HTML(http.StatusOK, "PasswordUpdate.html", gin.H{
			"token": usertoken,
			"email": claims.StandardClaims.Audience},
		)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "憑證失效，請重新發送重設密碼信件"})
	}
}

var conf = &oauth2.Config{
	ClientID:     os.Getenv("Client_ID"),
	ClientSecret: os.Getenv("Client_Secret"),
	RedirectURL:  "http://localhost:8080/Member/auth/callback",
	Scopes: []string{
		"profile",
		"email",
	},
	Endpoint: google.Endpoint,
}

// Google登入
func CallGoogle(c *gin.Context) {
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
	//fmt.Printf("Visit the URL for the auth dialog: %v", url+"\n")
}

// Google登入後處理
func CallbackGoogle(c *gin.Context) {
	ctx := context.Background()
	code := c.DefaultQuery("code", "")
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		fmt.Println(err)
		return
	}
	oauth2Service, err := GoogleOAuth2.NewService(ctx, option.WithTokenSource(conf.TokenSource(ctx, token)))
	if err != nil {
		fmt.Println("Error creating OAuth2 service:", err)
		return
	}
	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		fmt.Println("Error getting user info:", err)
		return
	}
	filter := bson.M{"memberid": userInfo.Email, "loginby": 2}
	var result bson.M
	err = MemberCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		//資料庫無此會員時，繼續進行註冊判斷
		if err == mongo.ErrNoDocuments {
			member := model.Member{
				MemberId:    userInfo.Email,
				Password:    nil,
				LoginBy:     2,
				EmailVerify: true,
				InfoVerify:  false,
				SocialId:    &userInfo.Id,
			}
			//儲存至MongoDB
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, insertErr := MemberCollection.InsertOne(ctx, member)
			if insertErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "資料儲存失敗"})
				return
			}
			session := sessions.Default(c)
			session.Set("Member", userInfo.Name)
			session.Set("MemberEmail", userInfo.Email)
			session.Save()
			c.Redirect(http.StatusFound, "/Member/Profile-Setup")
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "伺服器錯誤"})
		return
	}
	//如有會員則顯示已註冊過
	session := sessions.Default(c)
	session.Set("Member", userInfo.Name)
	session.Set("MemberEmail", userInfo.Email)
	session.Save()
	fmt.Println(result["infoverify"])
	if result["infoverify"] == false {
		c.Redirect(http.StatusFound, "/Member/Profile-Setup")
		return
	}
	c.Redirect(http.StatusFound, "/Member/Index")
}

// 頁首登入確認
func LoginCheck(c *gin.Context) {
	session := sessions.Default(c)

	user := session.Get("Member")
	if user == nil {
		c.JSON(http.StatusOK, gin.H{"logged_in": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"logged_in": true, "member": user})
	}
}

// 登出
func Logout(c *gin.Context) {
	// 清除session
	session := sessions.Default(c)
	session.Clear()
	fmt.Println(session.Get("Member"))
	fmt.Println(session.Get("MemberEmail"))
	session.Save()

	c.Redirect(http.StatusFound, "/Member/Index")
}

// 外部註冊填寫資料頁面
func ProfileSetupWeb(c *gin.Context) {
	c.HTML(http.StatusOK, "G-Profile-Setup.html", nil)
}

// 註冊信箱確認
func EmailVerifyCheck(c *gin.Context) {
	usertoken := c.Param("token")
	// 根據短token 找到對應的 JWT
	jwtToken, exists := tokenMapping[usertoken]
	if !exists {
		c.JSON(404, nil)
		return
	}

	// 解碼 JWT
	token, _ := jwt.ParseWithClaims(jwtToken.JWTToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		filter := bson.M{
			"memberid": claims.StandardClaims.Audience,
		}
		memberupdate := bson.M{
			"$set": bson.M{
				"emailverify": true,
			},
		}
		//儲存至MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		//更新member的個人資料填寫狀況
		_, updateErr := MemberCollection.UpdateOne(ctx, filter, memberupdate)
		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "伺服器錯誤"})
			return
		}
		c.Redirect(http.StatusFound, "/Member/Index")
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "憑證失效，請重新發送重設密碼信件"})
	}
}
