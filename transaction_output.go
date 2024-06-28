package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TXOutput struct {
	Value        int
	PubKeyHash   []byte
}

// Lock signs the output
func (out *TXOutput) lock(address []byte) {
	pubKeyHash := base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) isLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

// NewTXOutput create a new TXOutput
func newTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.lock([]byte(address))

	return txo
}

type TXOutputs struct {
	Outputs []TXOutput
}

func (outs TXOutputs) serialize() []byte {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}

func deserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}