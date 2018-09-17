package main

import (
	"crypto/md5"
	"net/http"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
)

var (
	loginUser = make(map[string]string)
)

func vaildLogin(ct *gin.Context) bool {
	auth, err := ct.Cookie("auth")
	if err != nil {
		return false
	}

	if _, ok := loginUser[auth]; ok != false {
		return true
	}

	return false
}

func userLogin(ct *gin.Context) {
	name, err := ct.GetPostForm("name")
	if err == false {
		ct.String(http.StatusBadRequest, "ok")
		return
	}

	auth := fmt.Sprintf("%x", md5.Sum([]byte(name+time.Now().String())))
	loginUser[auth] = name

	ct.SetCookie("auth", auth, 3600, "", "", false, false)
	ct.String(http.StatusOK, "ok")
}

func userSign(ct *gin.Context) {
	name, err := ct.GetPostForm("name")
	if err == false {
		ct.String(http.StatusBadRequest, "ok")
		return
	}

	_, sqlerr := mysqlDB.Exec("insert into user(userID,createDate) values(?,now())", name)
	if sqlerr != nil {
		ct.AbortWithError(http.StatusBadGateway, sqlerr)
		return
	}

	ct.String(http.StatusOK, "ok")
	return
}

func userFollow(ct *gin.Context) {

}
