package main

import (
	"encoding/json"
	"fmt"
	"monitor/api"
	"monitor/maindb"
	mydb "monitor/monitor/mydb"
	rs "monitor/monitor/rs"
	web "monitor/monitor/web"
	tg "monitor/tg"
	ws "monitor/ws"
	_ "net/http/pprof"

	_ "github.com/go-sql-driver/mysql"

	// "runtime"
	// "strconv"
	"time"
)

const (

	host                = "127.0.0.1"
	web_socket_password = "1qaz2wsx#EDCXZASWQ@!"
	APIEndpoint         = "https://api.telegram.org/bot%s/%s"
	FileEndpoint        = "https://api.telegram.org/file/bot%s/%s"
	Notify_Group        = tg.Notify_Group
	DB_Delay            = 1
	WEB_Delay           = 10
	
)

var (
	// 拿來判斷是否更新監控list的chan
	fresh_tb 	chan bool
	// 拿來判斷是否更新傳送 db_list的chan
	send_db 	chan bool
	// 拿來判斷是否更新傳送 redis_list的chan
	send_redis chan bool
	// 拿來判斷是否更新傳送 web_list的chan
	send_web 	chan bool
)

type row_data_struct struct {
	id      int
	group   string
	name    string
	ip      string
	port    string
	rule    string
	status  bool
	message string
	
}



func bToMb(b uint64) uint64 {
	return b / 1024
}

// func PrintMemUsage() {
// 	var m runtime.MemStats
// 	runtime.ReadMemStats(&m)
// 	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
// 	fmt.Printf("Alloc = %v KiB", bToMb(m.Alloc))
// 	fmt.Printf("\tHeapInuse = %v", bToMb(m.HeapInuse))
// 	fmt.Printf("\tHeapInuse = %v", bToMb(m.HeapInuse))
// 	fmt.Printf("\tTotalAlloc = %v KiB", bToMb(m.TotalAlloc))
// 	fmt.Printf("\tSys = %v KiB", bToMb(m.Sys))
// 	fmt.Printf("\totherSys = %v KiB", bToMb(m.OtherSys))
// 	fmt.Printf("\tHeapReleased = %v KiB", bToMb(m.HeapReleased))
// 	fmt.Printf("\tNumGC = %v\n", m.NumGC)

// }
func init() {

	// 拿來判斷是否更新傳送 db_list的chan
	send_db = make(chan bool, 1)

	// 拿來判斷是否更新傳送 redis_list的chan
	send_redis = make(chan bool, 1)

	// 拿來判斷是否更新傳送 web_list的chan
	send_web = make(chan bool, 1)

	
}


func main() {

	// 讓抓取各個需要監控的列表 自己跑一個thread
	maindb.Select_all_db()
	
	// 讓telegram 機器人在接收command 及發送 訊息 自己跑一個thread
	go tg.Telegram()

	// 讓監控頁面用的websocket 自己跑一個thread
	go ws.Start_websocket()

	// 讓抓取DB資訊用的功能 自己跑一個thread
	go mydb.Get_db_inf(send_db)
	// 讓抓取監控網站回應時間用的功能 自己跑一個thread
	go web.Get_web_response(send_web)
	// 讓抓取監控redis回應時間用的功能 自己跑一個thread
	go rs.Get_redis_response(send_redis)
	
	// 啟動web server
	go api.StartServer()
	// pprof
	// go func() {
	// 	http.ListenAndServe("0.0.0.0:8080", nil)
	// }()

	//golang 在跑main 時 需要跑一個無限迴圈 才能一直持續執行
	for {
		select {
		case <-tg.Fresh_tb:
			web.Fresh_tb <- true
			mydb.Fresh_tb <-true
			rs.Fresh_tb <- true
			time.Sleep(3*time.Second)
			maindb.Select_all_db()
		case <-send_redis:
			data,_:=json.Marshal(rs.RedisList)
			ws.Send_ws_data(data)
		case <-send_web:
			data,_:=json.Marshal(web.WebList)
			ws.Send_ws_data(data)
		case <-send_db:
			data,_:=json.Marshal(mydb.DBList)
			ws.Send_ws_data(data)
		default:
			time.Sleep(1*time.Second)
			if len(ws.LoginList) != 0 {
				fmt.Println(ws.LoginList)
				fmt.Println(len(ws.Get_manager().Clients))
			}
		}
	}
}
