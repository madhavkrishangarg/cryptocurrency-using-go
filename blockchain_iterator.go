package main

import(
	"log"
	"github.com/boltdb/bolt"
)

type blockchainIterator struct{
	currentHash []byte
	db *bolt.DB
}

func (i *blockchainIterator) next() *block{
	var block *block

	err := i.db.View(func(tx *bolt.Tx) error{
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = deserialize(encodedBlock)

		return nil
	})
	if err != nil{
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}