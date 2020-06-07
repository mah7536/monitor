package web

import (
	"crypto/tls"
	"monitor/tg"
	"net/http"
	"strconv"
	"time"
)

const (
	WEB_Delay = 10
)

type Web_data_struct struct {
	ID           int    `json:"webid" db:"Id"`
	Type    	 string `json:"type" `
	DOMAIN       string `json:"domain" db:"Domain"`
	WEBNAME      string `json:"webname" db:"Webname"`
	STATUS       int    `json:"status"`
	RESPONSETIME int    `json:"responsetime"`
	ERR          string `json:"err"`
}
type WEB_Websocket_data map[int]*Web_data_struct

var (
	WebList WEB_Websocket_data
	Fresh_tb chan bool
	tr *http.Transport
	httpc *http.Client
)

func init(){
	WebList = make(map[int]*Web_data_struct)
	// 拿來判斷是否更新監控list的chan
	Fresh_tb = make(chan bool, 1)
	// 監控web用設定
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	httpc = &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}
}

// 針對要監控的網站 做response time的動作 並將時間記錄到web_list中再透過 websocker傳送
func Get_web_response(send_web chan bool) {
	ticker := time.NewTicker(WEB_Delay * time.Second)
	for {
		select {
		case <-Fresh_tb:
			for k, _ := range WebList {
				delete(WebList, k)
			}
		case <- ticker.C:
			for _, row := range WebList {
				start := time.Now()
				response, err := httpc.Get(row.DOMAIN)
				response_time := time.Since(start)
				if err != nil {
					
					WebList[row.ID].STATUS = -1
					WebList[row.ID].RESPONSETIME = -1
					WebList[row.ID].ERR = err.Error()
					tg.Bot.Send(tg.Tgmessage("Web:" + row.DOMAIN + " is down\nError Message:" + err.Error()))
					continue
				} 
				
				defer response.Body.Close()
				if response.StatusCode != 200 {
					tg.Bot.Send(tg.Tgmessage("Web:" + row.DOMAIN + " is ok\nBut Status is " + strconv.Itoa(response.StatusCode)))
				} else if response.StatusCode == 200 && row.STATUS == -1 {
					tg.Bot.Send(tg.Tgmessage("Web:" + row.DOMAIN + " come back to life\n"))
				}
				WebList[row.ID].STATUS = response.StatusCode
				WebList[row.ID].RESPONSETIME = int(response_time / time.Millisecond)
				WebList[row.ID].ERR = ""
	
				response = nil
				err = nil
			}
			send_web <- true
		}
		
	}
}