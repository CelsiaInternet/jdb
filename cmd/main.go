package main

import (
	"github.com/celsiainternet/elvis/cache"
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/event"
	_ "github.com/celsiainternet/jdb/drivers/postgres"
	jdb "github.com/celsiainternet/jdb/jdb"
)

func main() {
	_, err := cache.Load()
	if err != nil {
		console.Panic(err)
	}

	_, err = event.Load()
	if err != nil {
		console.Panic(err)
	}

	db, err := jdb.Load()
	if err != nil {
		console.Panic(err)
	}

	_, err = jdb.From("users").
		Where("phone").Eq("").
		Debug().
		All()
	if err != nil {
		console.Panic(err)
	}

	console.Debug("db:", db.Name)
}
