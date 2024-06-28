package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const reward = 100 // reward for mining a block

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
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

func newCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		randomData := make([]byte, 20)		// create a random byte slice of size 20
		_, err := rand.Read(randomData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randomData)	// convert the byte slice to a string
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := newTXOutput(reward, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.setTXID()

	return &tx
}

func newUTXOTransaction(from, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	wallets, err := newWallets()
	if err != nil {
		log.Panic(err)
	}

	wallet := wallets.getWallet(from)
	pubKeyHash := hashPubKey(wallet.PublicKey)

	acc, validOutputs := UTXOSet.findSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
    	log.Panic("Error: Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *newTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *newTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.setTXID()
	UTXOSet.blockchain.signTransaction(&tx, *wallet.PrivateKey)

	return &tx
}


func (tx Transaction) serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

func (tx *Transaction) hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.serialize())
	return hash[:]
}

func (tx *Transaction) sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.isCoinbase() {
		return
	}

	txCopy := tx.trimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.hash()
		txCopy.Vin[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}

func (tx *Transaction) trimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

func (tx *Transaction) verify(prevTXs map[string]Transaction) bool {
	if tx.isCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("Error: Previous transaction is not correct")
		}
	}

	txCopy := tx.trimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.hash()
		txCopy.Vin[inID].PubKey = nil

		r, s := big.Int{}, big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes([]byte(vin.Signature[:(sigLen / 2)]))
		s.SetBytes([]byte(vin.Signature[(sigLen / 2):]))

		x, y := big.Int{}, big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes([]byte(vin.PubKey[:(keyLen / 2)]))
		y.SetBytes([]byte(vin.PubKey[(keyLen / 2):]))

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}


func (tx Transaction) toString() string{
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       PubKey: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}