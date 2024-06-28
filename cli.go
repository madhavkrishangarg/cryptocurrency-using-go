package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct{}

// func (cli *CLI) createBlockchain(address string) {
// 	bc := createBlockchain(address)
// 	bc.db.Close()
// 	fmt.Println("Done!")
// }

// func (cli *CLI) getBalance(address string) {
// 	bc := newBlockchain(address)
// 	defer bc.db.Close()

// 	balance := 0
// 	UTXOs := bc.findUTXO(HashPubKey(address))

// 	for _, out := range UTXOs {
// 		balance += out.Value
// 	}

// 	fmt.Printf("Balance of '%s': %d\n", address, balance)
// }

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// func (cli *CLI) addBlock(data string) {
// 	cli.bc.addBlock(data)
// 	fmt.Println("Success!")
// }

// func (cli *CLI) printChain() {
// 	bc := newBlockchain("")
// 	defer bc.db.Close()

// 	bci := bc.iterator()

// 	for {
// 		block := bci.next()

// 		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
// 		fmt.Printf("Hash: %x\n", block.Hash)
// 		pow := newPow(block)
// 		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.validate()))
// 		fmt.Println()

// 		if len(block.PrevBlockHash) == 0 {
// 			break
// 		}
// 	}
// }

// func (cli *CLI) send(from, to string, amount int) {
// 	bc := newBlockchain(from)
// 	defer bc.db.Close()

// 	tx := newUTXOTransaction(from, to, amount, bc)
// 	bc.mineBlock([]*Transaction{tx})
// 	fmt.Println("Success!")
// }

func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send the genesis block reward to")

	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")

	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

}