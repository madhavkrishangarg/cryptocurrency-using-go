package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

const walletFile = "wallets.dat"

type Wallets struct {
    Wallets map[string]*Wallet
}

type SerializableWallet struct {
    PublicKey []byte
    D         []byte
}

func newWallets() (*Wallets, error) {
    wallets := Wallets{}
    wallets.Wallets = make(map[string]*Wallet)
    err := wallets.loadFile()
    return &wallets, err
}

func (ws *Wallets) createWallet() string {
    wallet := newWallet()
    address := string(wallet.getAddress())
    ws.Wallets[address] = wallet
    return address
}

func (ws *Wallets) getAddresses() []string {
    var addresses []string
    for address := range ws.Wallets {
        addresses = append(addresses, address)
    }
    return addresses
}

func (ws *Wallets) getWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) saveToFile() {
    var content bytes.Buffer
    serializedWallets := make(map[string]SerializableWallet)

    for address, wallet := range ws.Wallets {
        serializedWallets[address] = SerializableWallet{
            PublicKey: wallet.PublicKey,
            D:         wallet.PrivateKey.D.Bytes(),
        }
    }

    encoder := gob.NewEncoder(&content)
    err := encoder.Encode(serializedWallets)
    if err != nil {
        log.Panic(err)
    }

    err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
    if err != nil {
        log.Panic(err)
    }
}

func (ws *Wallets) loadFile() error {
    if _, err := os.Stat(walletFile); os.IsNotExist(err) {
        return err
    }

    fileContent, err := ioutil.ReadFile(walletFile)
    if err != nil {
        log.Panic(err)
    }

    var serializedWallets map[string]SerializableWallet
    decoder := gob.NewDecoder(bytes.NewReader(fileContent))
    err = decoder.Decode(&serializedWallets)
    if err != nil {
        log.Panic(err)
    }

    wallets := make(map[string]*Wallet)
    for address, sWallet := range serializedWallets {
        privKey := &ecdsa.PrivateKey{
            PublicKey: ecdsa.PublicKey{
                Curve: elliptic.P256(),
            },
            D: new(big.Int).SetBytes(sWallet.D),
        }

        // Reconstruct public key from private key
        privKey.PublicKey.X, privKey.PublicKey.Y = privKey.PublicKey.Curve.ScalarBaseMult(sWallet.D)

        wallets[address] = &Wallet{
            PublicKey:  sWallet.PublicKey,
            PrivateKey: privKey,
        }
    }

    ws.Wallets = wallets
    return nil
}