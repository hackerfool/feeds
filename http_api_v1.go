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
	// ct.String(http.StatusOK, "ok")
	ct.Redirect(http.StatusFound, "/")

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

	user := makeFeedUser(userName)

	action := ct.DefaultQuery("action", "list")
	switch action {
	case "list":
		followsID, err := user.WatchList()
		if err != nil {
			fmt.Println(err)
			ct.String(http.StatusBadGateway, "error")
			return
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
		err := user.Follow(id)
		if err != nil {
			log.Println(err)
			ct.String(http.StatusBadGateway, "error")
			break
		}
		ct.String(http.StatusOK, "ok")
	case "unfollow":
		id, ok := ct.GetQuery("id")
		if ok == false {
			ct.String(http.StatusBadRequest, "follow id error")
			break
		}

		err := user.UnFollw(id)
		if err != nil {
			log.Println(err)
			ct.String(http.StatusBadGateway, "error")
			break
		}
		ct.String(http.StatusOK, "ok")
	}
	return
}

func userFans(ct *gin.Context) {
	var (
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

	user := makeFeedUser(userName)
	fansIDs, err := user.FanList()
	if err != nil {
		log.Println(err)
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}

	ct.JSON(http.StatusOK, gin.H{
		"fansID": fansIDs,
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

func userPostList(ct *gin.Context) {
	var (
		userName    string
		ok          bool
		limit, page int
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

	limit = getParmInt(ct, "limit")
	page = getParmInt(ct, "page")
	if limit < 0 {
		limit = 100
	}
	if page < 0 {
		page = 1
	}

	user := makeFeedUser(userName)
	texts, err := user.PostList(limit, page)
	if err != nil {
		log.Println(err)
		ct.AbortWithStatus(http.StatusBadGateway)
		return
	}

	ct.JSON(http.StatusOK, gin.H{"texts": texts, "limit": limit, "page": page})
}
