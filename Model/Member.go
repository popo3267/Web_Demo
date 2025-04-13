package model

type MemberInfo struct {
	MemberId   string `json:"memberid"`   //會員信箱(ID)
	MemberName string `json:"membername"` //會員名稱
	Birth      string `json:"birth"`      //會員生日
	Sex        string `json:"sex"`        //會員性別
	Phone      string `json:"phone"`      //會員電話
	Address    string `json:"address"`    //會員地址
}
type Member struct {
	MemberId    string  `json:"memberid"`    //會員信箱(ID)
	Password    *string `json:"password"`    //會員密碼
	LoginBy     int     `json:"loginby"`     //登入方式 會員=1,Google=2,Facebook=3
	EmailVerify bool    `json:"emailverify"` //信箱驗證確認
	InfoVerify  bool    `json:"infoverify"`  //個人資料確認
	SocialId    *string `json:"socialid"`    //社交平台ID(Google/Facebook)
}
