package api

import (
	"fmt"
	"monitor/ws"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func StartServer() {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowHeaders:     []string{"Origin","Content-Type","Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))
	// server.LoadHTMLGlob("view/*")
	// server.GET("/", GetIndex)
	server.POST("/login", Login)
	server.GET("/monitor/:loginId", Websocket)



	server.Run("0.0.0.0:1234")
}


func GetIndex(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", nil)
}


// 將其升級成websocket
func Websocket(ctx *gin.Context) {
	loginId := ctx.Param("loginId")
	if _,exist := ws.LoginList[loginId];exist {
		delete(ws.LoginList,loginId)
		ws.NewWebClient(ctx.Writer,ctx.Request)
		return
	}
	ctx.JSON( http.StatusOK, gin.H{"code":-1,"message":"錯誤"})
}


type LoginParam struct {
	Account  string `json:"account" binding:"required"`
	Password string `josn:"password" binding:"required"`
}

// 登入
func Login(ctx *gin.Context) {
	var request LoginParam
	if err:=ctx.ShouldBindBodyWith(&request,binding.JSON);err != nil {
		fmt.Println(err)
		ctx.JSON( http.StatusOK, gin.H{"code":-1,"message":"錯誤"})
		return
	}
	loginId := GetRandString()
	ws.LoginList[loginId] = true
	ctx.JSON( http.StatusOK, gin.H{"code":0,"message":"登入成功","loginId":loginId})
}