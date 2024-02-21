package main

import (
	"fmt"
	"log"
	"time"

	"github.com/NHAS/autoetcdtls/manager"
)

func main() {

	firstMember, err := manager.New("certs", "https://localhost:4433")
	if err != nil {
		log.Fatal(err)
	}

	err = firstMember.StartListening()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Blocking, press ctrl+c to continue...")

	token, err := firstMember.CreateToken("https://localhost:4444")
	if err != nil {
		log.Fatal("making token: ", err)
	}

	token.SetAdditional("fronk", "bonk")

	log.Println("join token: ", token)

	_, err = manager.Join(token.Token, "mock", map[string]func(name string, data string){
		"fronk": func(name, data string) {
			log.Println("got additional data: ", name, data)
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

	time.Sleep(35 * time.Second)

}
