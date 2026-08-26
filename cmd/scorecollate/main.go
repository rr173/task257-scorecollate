package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"task257-scorecollate/internal/demo"
	"task257-scorecollate/internal/httpapi"
	"task257-scorecollate/internal/service"
	"task257-scorecollate/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP 监听地址")
	dbPath := flag.String("db", "task257-scorecollate.db", "SQLite 数据库路径")
	smoke := flag.Bool("smoke-test", false, "运行离线端到端自检（含 DB 关闭重开恢复验证）后退出")
	flag.Parse()

	if *smoke {
		if err := demo.RunSmokeTest(*dbPath); err != nil {
			fmt.Fprintln(os.Stderr, "smoke-test failed:", err)
			os.Exit(1)
		}
		fmt.Println("smoke-test passed")
		return
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := store.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	svc := service.New(store.New(db))
	api := httpapi.New(svc)
	log.Printf("task257-scorecollate listening on %s (db=%s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, api); err != nil {
		log.Fatal(err)
	}
}
