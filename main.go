package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NHAS/autoetcdtls/manager"
)

func main() {
	err := manager.StartListening(":4433", "localhost", "certs")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Blocking, press ctrl+c to continue...")

	token, err := manager.CreateToken("certs", "localhost")
	if err != nil {
		log.Fatal("making token: ", err)
	}

	log.Println("join token: ", token)

	err = manager.Connect("https://localhost:4433", token, "mock")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	<-done
}
