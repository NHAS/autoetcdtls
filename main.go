package main

import (
	"fmt"
	"log"
	"time"

	"github.com/NHAS/autoetcdtls/manager"
)

func main() {

	firstMember := manager.New("certs")
	err := firstMember.StartListening(":4433", "localhost")
	if err != nil {
		log.Fatal(err)
	}

	firstMember.SetAdditional("fronk", "bonk")

	fmt.Println("Blocking, press ctrl+c to continue...")

	token, err := firstMember.CreateToken("localhost")
	if err != nil {
		log.Fatal("making token: ", err)
	}

	log.Println("join token: ", token)

	secondMember := manager.New("mock")

	secondMember.HandleAdditonal("fronk", func(name, data string) {
		log.Println("Got additional data: ", name, data)
	})

	err = secondMember.Join("https://localhost:4433", token)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

	time.Sleep(5 * time.Second)

}
