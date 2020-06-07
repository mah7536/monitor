package rs

import (
	tg "monitor/tg"
	"time"

	"github.com/go-redis/redis"
)

const (
	REDIS_Delay         = 5
	REDIS_Password      = "foobared"
)

type Redis_data_struct struct {
	ID           int    `json:"rsid" db:"Id"`
	Type		 string `json:"type" `
	IP           string `json:"ip" db:"IP"`
	PORT         string `json:"port" db:"PORT"`
	REDISNAME    string `json:"redisname" db:"REDISNAME"`
	STATUS       int    `json:"status"`
	RESPONSETIME int    `json:"responsetime"`
	ERR          string `json:"err"`
}

type Redis_Websocket_data map[int]*Redis_data_struct

var (
	RedisList Redis_Websocket_data
	thread string 
	Fresh_tb chan bool
)

func init(){
	RedisList = make(map[int]*Redis_data_struct)
	// 拿來判斷是否更新監控list的chan
	Fresh_tb = make(chan bool, 1)
}

// 針對要監控的redis 做response time的動作 並將時間記錄到redis_list中再透過 websocker傳送
func Get_redis_response(send_redis chan bool) {
	ticker := time.NewTicker(REDIS_Delay * time.Second)
	for {
		select {
		case <-Fresh_tb:
			for k, _ := range RedisList {
				delete(RedisList, k)
			}
		case <-ticker.C:
			for _, row := range RedisList {
				client := redis.NewClient(&redis.Options{
					Addr:     row.IP + ":" + row.PORT,
					Password: REDIS_Password, // no password set
					DB:       0,              // use default DB
				})
				start := time.Now()
				defer client.Close()
				_, err := client.Ping().Result()
				response_time := time.Since(start)
				// fmt.Println(pong, err, response_time)

				if err != nil {
					RedisList[row.ID].STATUS = -1
					RedisList[row.ID].RESPONSETIME = -1
					RedisList[row.ID].ERR = err.Error()
					tg.Bot.Send(tg.Tgmessage("Redis:" + row.IP + ":" + row.PORT + " is down\nError Message:" + err.Error()))

				} else {
					RedisList[row.ID].STATUS = 1
					RedisList[row.ID].RESPONSETIME = int(response_time / time.Millisecond)
					RedisList[row.ID].ERR = ""
				}
				client.Close()
			}
			send_redis <- true
		}
	}
}