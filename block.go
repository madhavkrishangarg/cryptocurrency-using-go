package main

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
)


type block struct {
	Timestamp int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash []byte
	Nonce int
}

func newBlock(transactions []*Transaction, prevBlockHash []byte) *block {
	block := &block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := newPow(block)
	nonce, hash := pow.run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// func (b *block) SetHash() {
// 	timestamp := []byte(strconv.FormatInt(b.timestamp, 10))
// 	headers := bytes.Join([][]byte{b.prevBlockHash, b.data, timestamp}, []byte{})
// 	hash := sha256.Sum256(headers)
// 	fmt.Printf("Hash: %x\n", hash)

// 	b.hash = hash[:]
// }

// func newBlockUnsafe(data string, prevBlockHash []byte) *block{
// 	block := &block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
// 	block.SetHash()
// 	block.nonce = 0

// 	return block

// }

func genesisBlock(coinbase *Transaction) *block {
	return newBlock([]*Transaction{coinbase}, []byte{})
}

func (b *block) serialize() []byte {
	var result bytes.Buffer		// buffer to store the serialized block
	encoder := gob.NewEncoder(&result)		// create a new encoder

	err := encoder.Encode(b)		// encode the block
	if err != nil {		// check for errors
		log.Panic(err)		
	}

	return result.Bytes()		// return the serialized block
}

func deserialize(data []byte) *block {
	var block block		// create a new block
	decoder := gob.NewDecoder(bytes.NewReader(data))		// create a new decoder

	err := decoder.Decode(&block)		// decode the data
	if err != nil {		// check for errors
		log.Panic(err)
	}

	return &block		// return the deserialized block
}

func (b *block) hashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
