package main

import (
	"Web_Demo/Controller"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	enverr := godotenv.Load() // 載入 .env 檔案
	if enverr != nil {
		log.Fatal("Error loading .env file")
	}
	router := Controller.Router()
	router.Static("static", "./static")
	router.LoadHTMLGlob("View/*")
	fmt.Println("http://localhost:8080/Member/PasswordforgetSend")
	err := router.Run(":8080")
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
