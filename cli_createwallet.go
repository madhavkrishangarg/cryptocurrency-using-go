package main

import "fmt"

func (cli *CLI) createWallet(){
	wallets, _ := newWallets()
	address := wallets.createWallet()
	wallets.saveToFile()

	fmt.Printf("Your new address: %s\n", address)
}