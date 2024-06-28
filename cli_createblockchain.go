package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address string) {
	if !validateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := createBlockchain(address)
	bc.db.Close()

	fmt.Println("Done!")
}