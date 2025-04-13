package Controller

import (
	model "Web_Demo/Model"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// 會員註冊
func MemberRegister(c *gin.Context) {
	//獲取網頁參數
	MemberId := c.PostForm("MemberId")
	MemberName := c.PostForm("MemberName")
	Password := c.PostForm("Password")
	Birth := c.PostForm("MemberBirth")
	Sex := c.PostForm("MemberSex")
	Phone := c.PostForm("MemberPhone")
	Address := c.PostForm("MemberAddress")
	ConfirmPassword := c.PostForm("ConfirmPassword")
	//註冊判斷
	// 查詢資料庫是否已有該會員
	filter := bson.M{"memberid": MemberId}
	var result bson.M
	err := MemberCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		//資料庫無此會員時，繼續進行註冊判斷
		if err == mongo.ErrNoDocuments {
			//判斷密碼是否等於密碼確認
			if Password != ConfirmPassword {
				c.HTML(http.StatusBadRequest, "Register.html", gin.H{"ConfirmResult": "用戶密碼與密碼確認不一致"})
				return
			}
			//API參數
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "資料儲存失敗"})
				return
			}
			memberInfo := model.MemberInfo{
				MemberId:   MemberId,
				MemberName: MemberName,
				//Password:   string(hashedPassword),
				Birth:   Birth,
				Sex:     Sex,
				Phone:   Phone,
				Address: Address,
			}
			hashPassword := string(hashedPassword)
			member := model.Member{
				MemberId:    MemberId,
				Password:    &hashPassword,
				LoginBy:     1,
				EmailVerify: false,
				InfoVerify:  true,
				SocialId:    nil,
			}
			//儲存至MongoDB
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, insertErr := MemberInfoCollection.InsertOne(ctx, memberInfo)
			if insertErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "資料儲存失敗"})
				return
			}
			_, insertErr = MemberCollection.InsertOne(ctx, member)
			if insertErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "資料儲存失敗"})
				return
			}
			shorttoken := generateShortToken(MemberId)
			fmt.Println(shorttoken)
			c.JSON(http.StatusOK, gin.H{
				"Message": MemberId + "註冊成功。\n已發送信件驗證，請至信箱確認",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "伺服器錯誤"})
		return
	}
	//如有會員則顯示已註冊過
	c.HTML(http.StatusConflict, "Register.html", gin.H{"IdResult": "該會員信箱已註冊過"})

}

// 會員登入
func MemberLogin(c *gin.Context) {
	MemberId := c.PostForm("MemberId")
	Password := c.PostForm("Password")
	//註冊判斷
	//查詢資料庫是否已有該會員
	filter := bson.M{"memberid": MemberId}
	var result model.Member
	err := MemberCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		//資料庫無此會員
		if err == mongo.ErrNoDocuments {
			c.HTML(http.StatusBadRequest, "Login.html", gin.H{"Result": "請輸入正確的帳號及密碼"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "伺服器錯誤"})
		return
	}

	var DBpassword = result.Password
	//如有會員則繼續進行判斷
	err = bcrypt.CompareHashAndPassword([]byte(*DBpassword), []byte(Password))
	if err != nil {
		// 當密碼不匹配時
		c.HTML(http.StatusBadRequest, "Login.html", gin.H{"Result": "請輸入正確的帳號及密碼"})
		return
	}
	var result2 model.MemberInfo
	err2 := MemberInfoCollection.FindOne(context.Background(), filter).Decode(&result2)
	if err2 != nil {
		fmt.Println("Login?")
	}
	//session儲存
	session := sessions.Default(c)
	session.Set("Member", result2.MemberName)
	session.Set("MemberEmail", MemberId)
	session.Save()
	fmt.Println(session)

	c.Redirect(http.StatusFound, "/Member/Index")
}

// 忘記密碼後密碼更新
func PasswordUpdate(c *gin.Context) {
	MemberId := c.PostForm("MemberId")
	Password := c.PostForm("Password")
	ConfirmPassword := c.PostForm("ConfirmPassword")
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

	if _, ok := token.Claims.(*Claims); ok && token.Valid {
		// 查詢資料庫是否已有該會員
		filter := bson.M{"memberid": MemberId}
		var result bson.M
		err := MemberCollection.FindOne(context.Background(), filter).Decode(&result)
		if err != nil {
			//資料庫無此會員
			if err == mongo.ErrNoDocuments {
				c.HTML(http.StatusBadRequest, "PasswordUpdate.html", gin.H{"Result": "?"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"message": "伺服器錯誤"})
			return
		}
		fmt.Printf("pw:%s,cpw:%s", Password, ConfirmPassword)
		//判斷密碼是否等於密碼確認
		if Password != ConfirmPassword {
			c.HTML(http.StatusBadRequest, "PasswordUpdate.html", gin.H{
				"Result": "用戶密碼與密碼確認不一致",
				"email":  MemberId,
			})
			return
		}
		//API參數
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "資料修改失敗"})
			return
		}

		//儲存至MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		memberupdate := bson.M{
			"$set": bson.M{
				"password": string(hashedPassword),
			},
		}
		_, updateErr := MemberCollection.UpdateOne(ctx, filter, memberupdate)
		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "伺服器錯誤"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Message": "資料修改成功",
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "憑證失效，請重新發送重設密碼信件"})
	}

}

// 外部登入後的會員資料填寫
func ProfileSetup(c *gin.Context) {
	//獲取網頁參數
	MemberName := c.PostForm("MemberName")
	Birth := c.PostForm("MemberBirth")
	Sex := c.PostForm("MemberSex")
	Phone := c.PostForm("MemberPhone")
	Address := c.PostForm("MemberAddress")

	session := sessions.Default(c)
	user := session.Get("MemberEmail")
	memberid := fmt.Sprintf("%s", user)
	// 查詢資料庫是否已有該會員
	filter := bson.M{"memberid": user}

	//API參數
	memberInfo := model.MemberInfo{
		MemberId:   memberid,
		MemberName: MemberName,
		Birth:      Birth,
		Sex:        Sex,
		Phone:      Phone,
		Address:    Address,
	}
	memberupdate := bson.M{
		"$set": bson.M{
			"infoverify": true,
		},
	}
	//儲存至MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, insertErr := MemberInfoCollection.InsertOne(ctx, memberInfo)
	if insertErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "資料儲存失敗"})
		return
	}
	//更新member的個人資料填寫狀況
	_, updateErr := MemberCollection.UpdateOne(ctx, filter, memberupdate)
	if updateErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "伺服器錯誤"})
		return
	}
	c.Redirect(http.StatusFound, "/Member/Index")
}
