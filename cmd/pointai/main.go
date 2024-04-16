package main

import (
	"github.com/ahmdrz/goinsta/v2"
	"log"
)

func main() {
	//cfg := config.Load()
	//mongodb := db.ConnectToDB(cfg)
	//str := storage.ConnectToStorage(cfg)
	//
	//server.RunHTTPServer(cfg, mongodb, str)

	insta := goinsta.New("mika_shamshidov", "Milkshake12)")
	err := insta.Login()
	if err != nil {
		log.Fatal("failed logging in")
	}

	insta.Account.Following()

}
