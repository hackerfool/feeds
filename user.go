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

//PostInfo struct
type PostInfo struct {
	ID       int
	Text     string
	Date     string
	Stars    int
	Comments int
}

func (u *feedUser) PostList(limit, page int) (texts []PostInfo, err error) {
	var (
		text = PostInfo{}
	)

	start := limit * (page - 1)

	row, err := mysqlDB.Query("select id,text,date,stars,comments from post where userID=? order by date desc limit ?,?", u.userID, start, limit)
	if err != nil {
		return
	}
	defer row.Close()

	texts = make([]PostInfo, 0, limit)
	for row.Next() {
		row.Scan(&text.ID, &text.Text, &text.Date, &text.Stars, &text.Comments)
		texts = append(texts, text)
	}

	return
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

	_, err = result.LastInsertId()
	if err != nil {
		log.Println(err)
		return -1
	}

	_, err = tx.Exec("update user set postIDincr=postIDincr+1 where userID=?", u.userID)
	if err != nil {
		log.Println(err)
		return -1
	}

	return postID
}

func (u *feedUser) Delete(postID uint32) (sqlerr error) {
	_, sqlerr = mysqlDB.Exec("delete from post where userID=? and postID=?", u.userID, postID)
	return
}

func (u *feedUser) Follow(watchID string) (sqlerr error) {
	_, sqlerr = mysqlDB.Exec("insert into watch(userID,watchID,watchDate) values(?,?,now())", u.userID, watchID)
	return
}

func (u *feedUser) UnFollw(watchID string) (sqlerr error) {
	_, sqlerr = mysqlDB.Exec("delete from watch where userID=? and watchID=?", u.userID, watchID)
	return
}

func (u *feedUser) MakePostMessage(postID int32) string {
	return fmt.Sprintf("POST:%s:%d", u.userID, postID)
}

func (u *feedUser) WatchList() (followsID []string, sqlerr error) {
	var (
		userID string
	)

	rows, sqlerr := mysqlDB.Query("select watchID from watch where userID=?", u.userID)
	if sqlerr != nil {
		return
	}
	defer rows.Close()

	followsID = make([]string, 0, 256)
	for rows.Next() {
		rows.Scan(&userID)
		followsID = append(followsID, userID)
	}
	return
}

func (u *feedUser) FanList() (IDs []string, sqlerr error) {
	var (
		userID string
	)

	rows, sqlerr := mysqlDB.Query("select userID from watch where watchID=?", u.userID)
	if sqlerr != nil {
		return
	}
	defer rows.Close()

	IDs = make([]string, 0, 256)
	for rows.Next() {
		rows.Scan(&userID)
		IDs = append(IDs, userID)
	}
	return
}
