package main

import (
	"bytes"
	"math/big"
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz") // base58 alphabet

func base58Encode(input []byte) []byte { // encode a byte slice to base58
	x := new(big.Int).SetBytes(input)           // convert the byte slice to a big.Int type
	base := big.NewInt(int64(len(b58Alphabet))) // create a new big.Int type with the length of the base58 alphabet
	zero := big.NewInt(0)                       // create a new big.Int type with the value 0
	mod := &big.Int{}                           // create a new big.Int type to store the modulus
	result := []byte{}                          // create a byte slice to store the result

	for x.Cmp(zero) != 0 { // while x is not equal to 0
		x.DivMod(x, base, mod)                            // divide x by the base and store the remainder in mod
		result = append(result, b58Alphabet[mod.Int64()]) // append the corresponding character to the result
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 { // reverse the result
		result[i], result[j] = result[j], result[i]
	}

	for b := range input { // for each byte in the input
		if b == 0x00 { // if the byte is 0
			result = append([]byte{b58Alphabet[0]}, result...) // append the corresponding character to the result
		} else {
			break // break the loop
		}
	}

	return result // return the result
}

func base58Decode(input []byte) []byte { // decode a base58 encoded byte slice
	result := big.NewInt(0) // create a new big.Int type with the value 0
	zeroBytes := 0          // number of leading zero bytes

	for _, b := range input { // for each byte in the input
		if int(b) == int(b58Alphabet[0]) { // if the byte is the first character in the base58 alphabet
			zeroBytes++ // increment the number of leading zero bytes
		} else {
			break // break the loop
		}
	}

	payload := input[zeroBytes:] // get the payload

	for _, b := range payload { // for each byte in the payload
		charIndex := bytes.IndexByte(b58Alphabet, b)     // get the index of the byte in the base58 alphabet
		result.Mul(result, big.NewInt(58))               // multiply the result by 58
		result.Add(result, big.NewInt(int64(charIndex))) // add the index to the result
	}

	decoded := result.Bytes()                                           // convert the result to a byte slice
	decoded = append(bytes.Repeat([]byte{0x00}, zeroBytes), decoded...) // add the leading zero bytes

	return decoded // return the decoded byte slice
}
