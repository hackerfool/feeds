package main

import (
	"log"
)

type user struct {
	userID string
}

func (u *user) Post(text string) {
	tx, err := mysqlDB.Begin()
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	result, err := tx.Exec("insert into post(text,date,stars,comments)  values(?,now(),0,0)", text)
	if err != nil {
		log.Println(err)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(u.userID, "post :", text, " PostID:", id)

	row, err := tx.Query("select post from user where userID=?", u.userID)
	if err != nil {
		log.Println(err)
		return
	}
	defer row.Close()

	postIDs := make([]uint32, 0)
	for row.Next() {
		row.Scan(&postIDs)
	}

	postIDs = append(postIDs, uint32(id))

	_, err = tx.Exec("update user set post=? where userID=?", postIDs, u.userID)
	if err != nil {
		log.Println(err)
	}
}

func (u *user) Delete(textID uint32) {

}

func (u *user) Watch(userID string) {

}

func (u *user) UnWatch(userID string) {

}
