package mydb

import (
	"database/sql"
	"fmt"
	"monitor/tg"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DB_Delay = 1 
)



type DBInfo struct {
	ID      int    `json:"id" db:"Id"`
	GROUP   string `json:"group" db:"Group"`
	NAME    string `json:"name" db:"Name"`
	IP      string `json:"ip" db:"Ip"`
	PORT    string `json:"port" db:"Port"`
	RULE    string `json:"rule" db:"Rule"`
	THREAD  string `json:"thread"`
	SLAVE   string `json:"slave"`
	STATUS  bool   `json:"status"`
	MESSAGE string `json:"message"`
	NOTIFY  bool   `json:"notify"`
	DB      *sql.DB
}

type Websocket_data map[int]*DBInfo

var (
	DBList Websocket_data
	Fresh_tb chan bool
	thread string 
	buff []interface{}
	data []string
)

func init(){
	DBList=make(map[int]*DBInfo)
	// 拿來判斷是否更新監控list的chan
	Fresh_tb = make(chan bool, 1)
	buff = make([]interface{}, 50) // 临时slice，用来通过类型检查
	data = make([]string, 50)      // 真正存放数据的slice
	for i, _ := range buff {
		buff[i] = &data[i] // 把两个slice关联起来
	}
}

// 針對要監控的db list 做 connection 數作query 並且針對slave 做判斷是否落後 並將此資料機錄到db_data 中 當有問題時 會直接透過telegram作通知 等到全部的資料都抓取監控完後 才會透過websocket傳送該資料
func Get_db_inf(send_db chan bool) {
	ticker := time.NewTicker(DB_Delay * time.Second)
	for {
		select {
		case <-Fresh_tb:
			for k, _ := range DBList {
				DBList[k].DB.Close()
				delete(DBList, k)
			}
		case <-ticker.C:
			for _, row := range DBList {
				message := ""
				if row.STATUS == true {
					err := row.DB.Ping()
					if err != nil {
						fmt.Println(err)
						if DBList[row.ID].NOTIFY == false {
							tg.Bot.Send(tg.Tgmessage("mysql IP:" + row.IP + " PORT:" + row.PORT + " can't connect!"))
							DBList[row.ID].NOTIFY = true
						}
						DBList[row.ID].STATUS = false
						DBList[row.ID].DB.Close()
						DBList[row.ID].MESSAGE = "無法連線 請確認"
						continue
					}
					processResult, err := row.DB.Query("Select count(1) from performance_schema.threads where PROCESSLIST_HOST is not null;")
					if err != nil {
						if DBList[row.ID].NOTIFY == false {
							tg.Bot.Send(tg.Tgmessage("mysql IP:" + row.IP + " PORT:" + row.PORT + " is connect but it can't not show processlist ,please check grant privileges!"))
							DBList[row.ID].NOTIFY = true
						}
						DBList[row.ID].STATUS = false
						DBList[row.ID].DB.Close()
						DBList[row.ID].MESSAGE = "請確認連入帳號的權限"
						continue
					}
					defer processResult.Close()
					for processResult.Next() {
						processResult.Scan(&thread)
	
						slave := "0"
						io_running := ""
						sql_running := ""
						second := ""
						sql_error := ""
						if row.RULE == "Slave" {
							slaveResult, err := row.DB.Query("show slave status")
							defer slaveResult.Close()
							if err != nil {
	
							}
							cols, _ := slaveResult.Columns()
							for slaveResult.Next() {
								slaveResult.Scan(buff...) // ...是必须的
							}
							for k, col := range data {
								if cols[k] == "Slave_IO_Running" {
									io_running = col
								} else if cols[k] == "Slave_SQL_Running" {
									sql_running = col
								} else if cols[k] == "Seconds_Behind_Master" {
									second = col
								} else if cols[k] == "Last_Errno" {
									sql_error = col
								}
							}
	
							if io_running == "Yes" && sql_running == "Yes" {
								slave = second
								if second != "0" && second != "1" && second != "2" {
									message = "slave 落後"
									tg.Bot.Send(tg.Tgmessage("mysql IP:" + DBList[row.ID].IP + " PORT:" + DBList[row.ID].PORT + " 目前落後秒數為: " + second))
								}
	
							} else if io_running == "No" && sql_running == "No" {
								if DBList[row.ID].NOTIFY == false {
									tg.Bot.Send(tg.Tgmessage("mysql IP:" + DBList[row.ID].IP + " PORT:" + DBList[row.ID].PORT + " 請確認是否有開啟Slave"))
									DBList[row.ID].NOTIFY = true
								}
								slave = sql_error
								row.STATUS = false
								message = "replication is not start"
							} else {
								if DBList[row.ID].NOTIFY == false {
									tg.Bot.Send(tg.Tgmessage("mysql IP:" + DBList[row.ID].IP + " PORT:" + DBList[row.ID].PORT + " slave error! error no:" + sql_error))
									DBList[row.ID].NOTIFY = true
								}
								slave = sql_error
								row.STATUS = false
								message = "slave replication error"
							}
						}
						if DBList[row.ID].STATUS == true && DBList[row.ID].NOTIFY == true {
							tg.Bot.Send(tg.Tgmessage("mysql IP:" + row.IP + " PORT:" + row.PORT + " 恢復正常"))
							DBList[row.ID].NOTIFY = false
						}
						DBList[row.ID].THREAD = thread
						DBList[row.ID].SLAVE = slave
						DBList[row.ID].MESSAGE = message
	
					}
					processResult = nil
				} else {
					if DBList[row.ID].NOTIFY == false {
						tg.Bot.Send(tg.Tgmessage("mysql IP:" + row.IP + " PORT:" + row.PORT + " can't connect!"))
						DBList[row.ID].NOTIFY = true
					}
					DBList[row.ID].MESSAGE =  "無法連線 請確認"
				}
			
		}
		send_db <- true
		}
	}
	
}