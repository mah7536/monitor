package maindb

import (
	"database/sql"
	"fmt"
	"monitor/monitor/mydb"
	"monitor/monitor/rs"
	"monitor/monitor/web"
	"os"
)

const (
	database            = "performance_schema"
	user                = "arther"
	password            = "1qazxsw23edcXZASWQ!@"
	main_host           = "192.168.0.156"
	main_database       = "test"
	main_user           = "monitor"
	main_password       = "1qaz"
)

var (
	Main_db *sql.DB
)

func init() {
	mainDb,err := Create_New_DB(main_user, main_password, main_host, "3306",main_database)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	// 判斷 監控主機用DB連線是否正常
	err = mainDb.Ping()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	// 設置監控用主機連線timeout 值 -1 代表不限制
	mainDb.SetConnMaxLifetime(-1)
	
	fmt.Println(Main_db)
	Main_db = mainDb
}

func Create_New_DB(user string, password string, host string, port string, database string) (*sql.DB, error ) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?allowNativePasswords=true&timeout=3s", user, password, host, port, database)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil,err
	}
	return db,nil
}

// 此為專門抓取要監控的各個資料 如要監控的DB清單 web清單等等
func Select_all_db() {
	fmt.Println("更新列表")
	go Get_All_DB_List()
	go Get_All_Redis_List()
	go Get_All_Web_List()	
}

// 取得需要偵測的db的列表
func Get_All_DB_List() {
	results, err := Main_db.Query("Select `Id`,`Group`,`Name`,`Ip`,`Port`,`Rule` from monitor.all_db_instance ;")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer results.Close()
	for results.Next() {
		newDb:=&mydb.DBInfo{}
		results.Scan(
			&newDb.ID,
			&newDb.GROUP,
			&newDb.NAME,
			&newDb.IP,
			&newDb.PORT,
			&newDb.RULE,
		)
		db,err := Create_New_DB(user, password, newDb.IP, newDb.PORT, database)
		if err != nil {
			newDb.STATUS = false
			newDb.NOTIFY = false
			mydb.DBList[newDb.ID] = newDb
			continue
		}
		err = db.Ping()
		if err != nil {
			newDb.STATUS = false
			newDb.NOTIFY = false
			mydb.DBList[newDb.ID] = newDb
			continue
		}
		db.SetConnMaxLifetime(-1)
		newDb.DB = db
		newDb.STATUS = true
		newDb.NOTIFY = false
		mydb.DBList[newDb.ID] = newDb
	}
}

// 取得需要偵測的web的列表
func Get_All_Web_List() {
	web_results, web_err := Main_db.Query("Select `Id`,`Domain`,`Webname` from monitor.all_website ;")
	if web_err != nil {
		fmt.Println(web_err)
		return
	}
	defer web_results.Close()
	for web_results.Next() {
		newWeb := &web.Web_data_struct{}
		web_results.Scan(
			&newWeb.ID,
			&newWeb.DOMAIN,
			&newWeb.WEBNAME,
		)
		web.WebList[newWeb.ID] = newWeb
	}
}


// 取得需要偵測的redis的列表
func Get_All_Redis_List() {
	redis_results, redis_err := Main_db.Query("Select `Id`,`IP`,`PORT`,`REDISNAME` from monitor.all_redis ;")
	if redis_err != nil {
		fmt.Println(redis_err)
		return
	}
	defer redis_results.Close()
	for redis_results.Next() {
		newRedis := &rs.Redis_data_struct{}
		redis_results.Scan(
			&newRedis.ID,
			&newRedis.IP,
			&newRedis.PORT,
			&newRedis.REDISNAME,
		)
		rs.RedisList[newRedis.ID] = newRedis
	}
}