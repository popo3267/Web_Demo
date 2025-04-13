package Controller

import (
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type TokenInfo struct {
	JWTToken  string
	ExpiresAt time.Time
}

// JWT Claims 結構
type Claims struct {
	jwt.StandardClaims
}

// JWT簽證密鑰，需妥善處理
var jwtSecret = []byte("secret")

// 短期 token 映射
var tokenMapping = make(map[string]TokenInfo)

// 產生時長1小時的JWT
func generateJWT(user string) (string, error) {
	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			Audience:  user,
			ExpiresAt: time.Now().Add(time.Minute * 3).Unix(), // 1 小時後過期
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 將JWT轉成較短token
func generateShortToken(user string) string {
	// 生成 UUID
	shortToken := uuid.New().String()

	// 生成 JWT
	jwtToken, err := generateJWT(user)
	if err != nil {
		fmt.Println(err)
	}
	// 儲存 UUID 和對應的 JWT
	tokenMapping[shortToken] = TokenInfo{
		JWTToken:  jwtToken,
		ExpiresAt: time.Now().Add(time.Minute * 3),
	}
	return shortToken
}

// 驗證較短token
func validateShortToken(c *gin.Context) {
	shortToken := c.Param("token")

	// 根據短token 找到對應的 JWT
	jwtToken, exists := tokenMapping[shortToken]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// 解碼 JWT
	token, _ := jwt.ParseWithClaims(jwtToken.JWTToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		c.JSON(http.StatusOK, gin.H{"Hello": claims.StandardClaims.Audience})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
	}
}

// 忘記密碼時發送信件
func PasswordforgetSend(c *gin.Context) {
	user := c.PostForm("MemberId")
	if user == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "user required"})
		return
	}
	shorttoken := generateShortToken(user)
	fmt.Println(user)
	fmt.Println(shorttoken)
	resetLink := "http://localhost:8080/Member/PasswordUpdate/" + shorttoken
	fmt.Println(resetLink)
	// 寄送郵件
	// 設定寄件者的 Gmail 帳號和應用程式密碼
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	fmt.Println("Host:", host)
	fmt.Println("Port:", port)
	auth := smtp.PlainAuth(
		"",
		os.Getenv("SMTP_EMAIL"),    // 寄件者 Gmail
		os.Getenv("SMTP_PASSWORD"), // Gmail 應用程式密碼
		os.Getenv("SMTP_HOST"),
	)
	msg := []byte("Subject: 重設密碼連結\r\n\r\n重設密碼的連結為: " + resetLink)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT")), // SMTP 伺服器位址和埠號
		auth,
		os.Getenv("SMTP_EMAIL"), // 寄件者 Gmail
		[]string{user},          // 收件者的電子郵件地址
		[]byte(msg),
	)
	if err != nil {
		c.JSON(400, gin.H{"msg": err})
		fmt.Println("寄失敗")
	}
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"msg": "send"})
		fmt.Println("寄成功")
	}
}

// func sendmailTest(c *gin.Context) {
// 	user := c.Query("user") // 或從 body email資訊
// 	if user == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"msg": "user required"})
// 		return
// 	}

// 	shortToken := generateShortToken(user)
// 	resetLink := "http://localhost/PasswordUpdate/" + shortToken

// 	// 寄送郵件
// 	// 設定寄件者的 Gmail 帳號和應用程式密碼
// 	auth := smtp.PlainAuth(
// 		"",
// 		os.Getenv("SMTP_EMAIL"),    // 寄件者 Gmail
// 		os.Getenv("SMTP_PASSWORD"), // Gmail 應用程式密碼
// 		os.Getenv("SMTP_HOST"),
// 	)
// 	msg := []byte("Subject: 重設密碼連結\r\n\r\n重設密碼的連結為: " + resetLink)
// 	err := smtp.SendMail(
// 		fmt.Sprintf("%s:%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT")), // SMTP 伺服器位址和埠號
// 		auth,
// 		os.Getenv("SMTP_EMAIL"), // 寄件者 Gmail
// 		[]string{""},            // 收件者的電子郵件地址
// 		[]byte(msg),
// 	)
// 	if err != nil {
// 		c.JSON(400, gin.H{"msg": err})
// 	}
// 	if err == nil {
// 		c.JSON(http.StatusOK, gin.H{"msg": "send"})
// 	}
// }

func printAllTokens(c *gin.Context) {
	fmt.Println("Current tokens in memory:")
	for key, tokenInfo := range tokenMapping {
		fmt.Printf("Token: %s, ExpiresAt: %s\n", key, tokenInfo.ExpiresAt.Format(time.RFC3339))
	}
	fmt.Println("A--------------------------------------A")
}

// 過期token清理
func cleanupExpiredTokens() {
	for key, tokenInfo := range tokenMapping {
		if time.Now().After(tokenInfo.ExpiresAt) {
			// 如果目前時間超過 token 的過期時間，則刪除token
			delete(tokenMapping, key)
		}
	}
}

// token清理計時器
func startCleanupRoutine() {
	ticker := time.NewTicker(3 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				cleanupExpiredTokens()
			}
		}
	}()
}
