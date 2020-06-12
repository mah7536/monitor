package log

import (
	"fmt"
	"os"
	"time"
)

type LogFile struct {
	FileName string
	File *os.File
}

var (
	LogExt string = ".log"
	LogPath = "log/"
	CurrentLog *LogFile
)

const TIME_LAYOUT = "2006-01-02-15-04"

func init() {
	file, err := os.OpenFile(GetLogName(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	CurrentLog = &LogFile{
		FileName: GetLogName(),
		File: file,
	}
}


func GetNowTime() string {
	return time.Now().Format(TIME_LAYOUT)
}

func GetLogName() string {
	return LogPath + GetNowTime() + LogExt
}

func CheckLogName() {
	if CurrentLog.FileName != GetLogName() {
		CurrentLog.File.Close()
		file, err := os.OpenFile(GetLogName(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		CurrentLog.FileName = GetLogName()
		CurrentLog.File = file
	}
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func Log(content string) {
	CheckLogName() 
	if CurrentLog.File == nil {
		fmt.Println("file is nil! content is ",content)
	}
	CurrentLog.File.Write([]byte(content+"\n"))
}

