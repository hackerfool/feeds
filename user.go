package main

import (
	"fmt"
	"log"
)

type feedUser struct {
	userID string
}

func makeFeedUser(name string) *feedUser {
	return &feedUser{userID: name}
}

func (u *feedUser) New() error {
	_, sqlerr := mysqlDB.Exec("insert into user(userID,createDate) values(?,now())", u.userID)
	return sqlerr
}

func (u *feedUser) Post(text string) int32 {
	var postID int32

	tx, err := mysqlDB.Begin()
	if err != nil {
		log.Println(err)
		return postID
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	row, err := tx.Query("select postIDincr from user where userID=?", u.userID)
	if err != nil {
		log.Println(err)
		return -1
	}
	defer row.Close()

	for row.Next() {
		row.Scan(&postID)
	}

	result, err := tx.Exec("insert into post(text,date,stars,comments,userID,postID)  values(?,now(),0,0,?,?)", text, u.userID, postID)
	if err != nil {
		log.Println(err)
		return -1
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return -1
	}

	log.Println(u.userID, "post :", text, " PostID:", id)

	_, err = tx.Exec("update user set postIDincr=postIDincr+1 where userID=?", u.userID)
	if err != nil {
		log.Println(err)
		return -1
	}

	return postID
}

func (u *feedUser) Delete(postID uint32) {
	_, err := mysqlDB.Exec("delete from post where userID=? and postID=?", u.userID, postID)
	if err != nil {
		log.Println(err)
		return
	}
}

func (u *feedUser) Watch(watchID string) {
	_, err := mysqlDB.Exec("insert into watch(userID,watchID,watchDate) values(?,?,now())", u.userID, watchID)
	if err != nil {
		log.Println(err)
		return
	}
}

func (u *feedUser) UnWatch(watchID string) {
	_, err := mysqlDB.Exec("delete from watch where userID=? and watchID=?", u.userID, watchID)
	if err != nil {
		log.Println(err)
		return
	}
}

func (u *feedUser) MakePostMessage(postID int32) string {
	return fmt.Sprintf("POST:%s:%d", u.userID, postID)
}
