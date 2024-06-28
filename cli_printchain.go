package main

import (
	"fmt"
	"strconv"
)

func (cli *CLI) printChain() {
	bc := newBlockchain()
	defer bc.db.Close()

	bci := bc.iterator()

	for {
		block := bci.next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := newPow(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}