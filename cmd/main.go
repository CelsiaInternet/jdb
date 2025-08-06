package main

import (
	"github.com/celsiainternet/elvis/cache"
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/event"
	_ "github.com/celsiainternet/jdb/drivers/sqlite"
	jdb "github.com/celsiainternet/jdb/jdb"
)

func main() {
	_, err := cache.Load()
	if err != nil {
		panic(err)
	}

	_, err = event.Load()
	if err != nil {
		panic(err)
	}

	db, err := jdb.Load()
	if err != nil {
		panic(err)
	}

	console.Debug("db:", db.Name)
}
