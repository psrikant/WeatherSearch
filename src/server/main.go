// Author: Pirakalan

package main

import (
	"log"
	"net/http"
	"handler"
	"helper"
	"session"
	"time"
	"os/signal"
	"os"
	"syscall"
	"flag"
)

var (
	command = flag.String("display", "", "history")
)

func main() {
	// Display user session history in command line
	// Usage: --display=history
	flag.Parse()
	helper.PrintSessionHistory(*command)

	quit := make(chan struct{})
	go session.CleanSessions(1*time.Hour, quit)

	// To close the CleanSessions goroutine and database connection upon quiting program
	// Reference: https://gobyexample.com/signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<- sigs
		close(quit)
		session.CleanSessions(1*time.Hour, quit)
		helper.CloseDB()
		os.Exit(0)
	}()

	http.HandleFunc("/", handler.RootHandler)
	http.HandleFunc("/hello", handler.HelloHandler)
	http.HandleFunc("/search", handler.SearchHandler)
	http.HandleFunc("/createuser", handler.CreateUserHandler)
	http.HandleFunc("/login", handler.LoginHandler)
	http.HandleFunc("/passwordreset", handler.PasswordResetHandler)
	http.HandleFunc("/logout", handler.LogoutHandler)
	http.HandleFunc("/citylist.json", handler.CityHandler)
	
	log.Printf("Go to localhost:8081/")
	err := http.ListenAndServe(":8081", nil)

	checkErr("ListenAndServe error", err)
}

func checkErr(message string, err error) {
	if err != nil {
		log.Printf("%s: %s", message, err.Error())
	}
}