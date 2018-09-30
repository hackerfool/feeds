package main

import (
	"crypto/md5"
	"log"
	"net/http"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
)

var (
	loginUser = make(map[string]string)
)

func vaildLogin(ct *gin.Context) {
	auth, err := ct.Cookie("auth")
	if err != nil {
		ct.Redirect(http.StatusFound, "/web/login.html")
		return
	}

	if _, ok := loginUser[auth]; ok != false {
		return
	}

	ct.Redirect(http.StatusFound, "/web/login.html")
	ct.Abort()
	return
}

func userLogin(ct *gin.Context) {
	name, err := ct.GetQuery("name")
	if err == false {
		ct.String(http.StatusBadRequest, "ok")
		return
	}

	auth := fmt.Sprintf("%x", md5.Sum([]byte(name+time.Now().String())))
	loginUser[auth] = name

	ct.SetCookie("auth", auth, 3600, "", "", false, false)
	ct.String(http.StatusOK, "ok")

	return
}

func userSign(ct *gin.Context) {
	name, err := ct.GetQuery("name")
	if err == false {
		ct.String(http.StatusBadRequest, "ok")
		return
	}

	user := makeFeedUser(name)
	sqlerr := user.New()
	if sqlerr != nil {
		ct.AbortWithError(http.StatusBadGateway, sqlerr)
		return
	}

	ct.String(http.StatusOK, "ok")
	return
}

func userFollow(ct *gin.Context) {
	var (
		followsID = make([]string, 0, 255)
		userID    string
		userName  string
		ok        bool
	)
	auth, err := ct.Cookie("auth")
	if err != nil {
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}

	if userName, ok = loginUser[auth]; !ok {
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}

	action := ct.DefaultQuery("action", "list")
	switch action {
	case "list":
		rows, err := mysqlDB.Query("select watchID from watch where userID=?", userName)
		if err != nil {
			log.Println(err)
			ct.AbortWithStatus(http.StatusBadGateway)
			return
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&userID)
			followsID = append(followsID, userID)
		}

		ct.JSON(http.StatusOK, gin.H{
			"followsID": followsID,
		})
	case "follow":
		id, ok := ct.GetQuery("id")
		if ok == false {
			ct.String(http.StatusBadRequest, "follow id error")
			break
		}
		_, err := mysqlDB.Exec("insert into watch(userID,watchID,watchDate) values(?,?,now())", userName, id)
		if err != nil {
			log.Println(err)
			ct.String(http.StatusBadGateway, "ok")
			break
		}
		ct.String(http.StatusOK, "ok")
	case "unfollow":
		id, ok := ct.GetQuery("id")
		if ok == false {
			ct.String(http.StatusBadRequest, "follow id error")
			break
		}

		_, err := mysqlDB.Exec("delete from watch where userID=? and watchID=?", userName, id)
		if err != nil {
			log.Println(err)
			ct.String(http.StatusBadGateway, "ok")
			break
		}
		ct.String(http.StatusOK, "ok")
	}

	return
}

func userFans(ct *gin.Context) {
	var (
		fansID   = make([]string, 0, 255)
		userID   string
		userName string
		ok       bool
	)
	auth, err := ct.Cookie("auth")
	if err != nil {
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}

	if userName, ok = loginUser[auth]; !ok {
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}

	rows, err := mysqlDB.Query("select userID from watch where watchID=?", userName)
	if err != nil {
		log.Println(err)
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&userID)
		fansID = append(fansID, userID)
	}

	ct.JSON(http.StatusOK, gin.H{
		"fansID": fansID,
	})
	return
}

func userPost(ct *gin.Context) {
	var (
		text     string
		userName string
		ok       bool
	)
	auth, err := ct.Cookie("auth")
	if err != nil {
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}

	if userName, ok = loginUser[auth]; !ok {
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}
	text, ok = ct.GetPostForm("post")
	if !ok {
		ct.AbortWithStatus(http.StatusBadRequest)
		return
	}

	user := makeFeedUser(userName)
	if user.Post(text) < 0 {
		log.Println(err)
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}

	ct.String(http.StatusOK, "ok")

	return
}
