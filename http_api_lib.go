package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func getParmInt(ct *gin.Context, key string) (value int) {
	var (
		v    string
		hasK bool
		err  error
	)
	if v, hasK = ct.GetQuery(key); !hasK {
		return -1
	}

	if value, err = strconv.Atoi(v); err != nil {
		return -1
	}
	return
}
