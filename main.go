package main

import (
	"fmt"
	"log"

	"github.com/NHAS/autoetcdtls/manager"
)

func main() {

	firstMember := manager.NewManager("certs")
	err := firstMember.StartListening(":4433", "localhost")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Blocking, press ctrl+c to continue...")

	token, err := firstMember.CreateToken("localhost")
	if err != nil {
		log.Fatal("making token: ", err)
	}

	log.Println("join token: ", token)

	secondMember := manager.NewManager("mock")

	err = secondMember.Join("https://localhost:4433", token)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

}
