package main

import (
	"fmt"
	"log"
)

func (cli *CLI) listAddresses() {
	wallets, err := newWallets()
	if err != nil {
		log.Panic(err)
	}

	addresses := wallets.getAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}