package tg

import (
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	APIEndpoint  = "https://api.telegram.org/bot%s/%s"
	FileEndpoint = "https://api.telegram.org/file/bot%s/%s"
	Notify_Group = 327835980
	WEB_Delay    = 10
)


var (
	Bot *tgbotapi.BotAPI
	Select_dblist_or_not chan bool
	Fresh_tb chan bool
)


func init() {
	bot, err := tgbotapi.NewBotAPI("841452677:AAFTaj5bkrKMcE04m3SZe06fiGnn9vJe6n0")
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	Bot = bot
	// 是否重新整理db_list及web_list的chan
	Select_dblist_or_not = make(chan bool, 1)

	// 拿來判斷是否更新監控list的chan
	Fresh_tb = make(chan bool, 1)
}
// telegram 接收指令用的function
func Telegram() {
	// 針對telegram 的細部做設定
	Bot.Debug = false
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := Bot.GetUpdatesChan(u)

	// 監聽各種傳給bot的指令 (目前尚未指定chat_id)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() == true {
			// 此為設置 當 telegram 收到 re 的指令時 則重新從 監控主機的db中抓取要監控的名單
			if update.Message.Command() == "re" {
				Fresh_tb <- true
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "重新整理Montior顯示清單")
				msg.ReplyToMessageID = update.Message.MessageID
				Bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "請確認該指令有動作")
				msg.ReplyToMessageID = update.Message.MessageID
				Bot.Send(msg)
			}

		}
		// *power = !*power
		// if *power == false {
		// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "okok")
		// 	msg.ReplyToMessageID = update.Message.MessageID
		// 	bot.Send(msg)
		// } else {
		// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "nono")
		// 	msg.ReplyToMessageID = update.Message.MessageID
		// 	bot.Send(msg)
		// }

	}
}

// 此為固定傳送到指定群組訊息用
func Tgmessage(message string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(Notify_Group, message)
	return msg
}
