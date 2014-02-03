package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/howeyc/gopass"
	"io"
)

func getPassword() []byte {
	fmt.Println("Please enter your password:")
	pass := gopass.GetPasswd()
	return pass
}

func getConfirmPassword() []byte {
	fmt.Println("Please supply an encryption password:")
	pass := gopass.GetPasswd()

	fmt.Println("Please type password again for confirmation:")
	passConfirmation := gopass.GetPasswd()

	if bytes.Equal(pass, passConfirmation) {
		return pass
	} else {
		fmt.Println("Sorry, passwords don't match.")
		return getConfirmPassword()
	}
}

// Hashes a byte array to sha2
func createSha2(input []byte) []byte {
	hash := sha256.New()
	hash.Write(input)
	sha := string(hash.Sum(nil))
	return []byte(sha)
}

func encrypt(password []byte, toEncrypt []byte) []byte {
	key := createSha2(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(toEncrypt) < aes.BlockSize {
		panic("Somehow the ciphertext is too short")
	}

	ciphertext := make([]byte, aes.BlockSize+len(toEncrypt))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], toEncrypt)

	return ciphertext
}

// TODO: Check if this encryption is actually working properly
func decrypt(password []byte, toDecrypt []byte) []byte {
	key := createSha2(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(toDecrypt) < aes.BlockSize {
		panic("Somehow the ciphertext is too short")
	}
	iv := toDecrypt[:aes.BlockSize]
	ciphertext := toDecrypt[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)

	decoded := make([]byte, len(ciphertext))
	cfb.XORKeyStream(decoded, ciphertext)

	return decoded
}
