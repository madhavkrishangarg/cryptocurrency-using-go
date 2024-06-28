package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from string, to string, amount int) {
	if !validateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}

	if !validateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := newBlockchain()
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	tx := newUTXOTransaction(from, to, amount, &UTXOSet)
	cbTx := newCoinbaseTX(from, "")		// cbTx is the coinbase transaction, which is the reward for the miner who mines the block
	txs := []*Transaction{cbTx, tx}		// txs is the list of transactions that will be included in the block

	newBlock := bc.mineBlock(txs)
	UTXOSet.update(newBlock)

	fmt.Println("Success!")
}