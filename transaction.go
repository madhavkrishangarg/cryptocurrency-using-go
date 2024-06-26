package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const reward = 100 // reward for mining a block

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type TXInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

type TXOutput struct {
	Value        int
	ScriptPubKey string
}

func (tx Transaction) isCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (tx *Transaction) setTXID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

//check whether the address initiated the transaction
func (in *TXInput) usesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.ScriptSig)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

//check whether the address is the recipient of the transaction
func (out *TXOutput) isLockedWithKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(out.ScriptPubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}


func newCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{reward, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.setTXID()

	return &tx
}

func HashPubKey(pubKey string) []byte {
	pubKeyHash := sha256.Sum256([]byte(pubKey))
	return pubKeyHash[:]
}

func newUTXOTransaction(from, to string, amount int, bc *blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	pubKeyHash := HashPubKey(from)
	acc, validOutputs := bc.findSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.setTXID()

	return &tx
}