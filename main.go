package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLConf struct {
	Mysql    string `json:"mysql"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

var defautlConf = &MySQLConf{
	"localhost:3306",
	"root",
	"",
	"log",
}

var db *sql.DB

func loadJsonFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&defautlConf)
	return err
}

var (
	mysqlPath = flag.String("c", "conf.json", "指定读取Mysql配置的文件")
	lPort     = flag.String("p", ":1234", "指定http的端口")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	flag.Parse()
	var err error
	if err = loadJsonFile(*mysqlPath); err != nil {
		panic(err)
	}

	mysql_url := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		defautlConf.User, defautlConf.Password,
		defautlConf.Mysql, defautlConf.Database)
	log.Println(mysql_url)
	db, err = sql.Open("mysql", mysql_url)
	if err != nil {
		log.Fatalf("Open database error: %s\n", err)
	}
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(16)

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ping database error: %s\n", err)
	}
	log.Println("Connect Database OK!")
	defer db.Close()
	srv := NewServer(*lPort)
	srv.Start()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	for sig := range c {
		log.Println("Catched Signal", sig)
		switch sig {
		case syscall.SIGINT:
			srv.Stop()
		// case SIG_STATUS:
		// 	log.Println("catch sigstatus, ignore")
		case syscall.SIGTERM:
			log.Println("catch sigterm, ignore")
		}
		return
	}
}
