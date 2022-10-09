package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"ipdog/cip"
	"ipdog/fakegeo"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/mattn/go-sqlite3"
)

// ctx = context.Background()

type Config struct {
	Token   string `json:"token"`
	Dbpath  string `json:"dbpath"`
	Redis   string `json:"redis"`
	Maxnum  int    `json:"maxnum"`
	Iplim   int    `json:"iplim"`
	Port    int    `json:"port"`
	Address string `json:"address"`
}

type userinfo struct {
	Ip    string `json:"ip"`
	Token string `json:"token"`
}

var config Config
var ctx = context.Background()

func init() {
	configfile, _ := os.ReadFile("config.json")
	// jsonstr := string(configfile)
	err := json.Unmarshal(configfile, &config)
	if err != nil {
		panic(err)
	}
	// ctx := context.Background()

}

func main() {
	rc, err := redis.ParseURL(config.Redis)
	if err != nil {
		panic(err)
	}
	rcl := redis.NewClient(rc)
	fmt.Println(config.Dbpath)
	db, err := sql.Open("sqlite3", config.Dbpath)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		qm := c.Query("token")
		q := c.ClientIP()
		if qm == config.Token {
			res := cip.GetIp_c(rcl, db, q)
			c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
		} else {
			if !Numlim(rcl) {
				res := cip.GetFIp_c(rcl, q)
				c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
			} else {
				res := cip.GetIp_c(rcl, db, q)
				c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
			}
		}
	})
	r.GET("/ip/:ip", func(c *gin.Context) {
		qm := c.Query("token")
		q := c.Params.ByName("ip")
		if qm == config.Token {
			res := cip.GetIp_c(rcl, db, c.Params.ByName("ip"))
			c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
		} else {
			if !Numlim(rcl) {
				res := cip.GetFIp_c(rcl, q)
				c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
			} else {
				res := cip.GetIp_c(rcl, db, c.Params.ByName("ip"))
				c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
			}
		}
	})

	r.GET("/fakeip/:ip", func(c *gin.Context) {
		q := c.Params.ByName("ip")
		if !cip.IsIp(q) {
			c.JSON(http.StatusBadGateway, gin.H{"err": "ip address is not valid"})
			return
		}
		qidex := cip.Ip2index(q)
		uij, err := rcl.Get(ctx, qidex).Result()
		if err == redis.Nil {
			a := fakegeo.Fakeip()
			jsoa, err := json.Marshal(a)
			if err != nil {
				panic(err)
			}
			rcl.Set(ctx, qidex, jsoa, time.Hour*2)
			fmt.Println(jsoa)
			c.JSON(http.StatusOK, gin.H{"ip": q, "res": a})
		} else if err != nil {
			panic(err)
		} else {
			var a map[string]string
			err := json.Unmarshal([]byte(uij), &a)
			if err != nil {
				panic(err)
			}
			c.JSON(http.StatusOK, gin.H{"ip": q, "res": a})
		}

	})
	r.GET("/fakeip/", func(c *gin.Context) {
		q := c.ClientIP()
		res := cip.GetFIp_c(rcl, q)
		c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
	})

	r.POST("/ip/", func(c *gin.Context) {
		var u userinfo
		if err := c.ShouldBindJSON(&u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		q := u.Ip
		if u.Token != config.Token {
			if !Numlim(rcl) {
				res := cip.GetFIp_c(rcl, q)
				c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
			} else {
				res := cip.GetIp_c(rcl, db, c.Params.ByName("ip"))
				c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
			}
		} else {
			res := cip.GetIp_c(rcl, db, c.Params.ByName("ip"))
			c.JSON(http.StatusOK, gin.H{"ip": q, "res": res})
		}

	})

	r.Run(fmt.Sprint(config.Address) + ":" + fmt.Sprint(config.Port))

}
