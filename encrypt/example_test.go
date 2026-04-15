// encrypt/example_test.go
package encrypt_test

import (
	"fmt"

	"github.com/f2xme/gox/encrypt"
)

func ExampleSHA256() {
	hash := encrypt.SHA256([]byte("hello"))
	fmt.Println(hash)
	// Output: 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
}

func ExampleMD5() {
	hash := encrypt.MD5([]byte("123456"))
	fmt.Println(hash)
	// Output: e10adc3949ba59abbe56e057f20f883e
}

func ExampleBlake3() {
	hash := encrypt.Blake3([]byte("123456"))
	fmt.Println(hash)
	// Output: 7adb787627ad5ee341fa0ba46a956e78fd85c39e195119bb260d5181b4f1e4ba
}

func ExampleAESEncrypt() {
	// 32 字节密钥 = AES-256
	key := []byte("12345678901234567890123456789012")
	plaintext := []byte("secret message")

	ciphertext, err := encrypt.AESEncrypt(plaintext, key)
	if err != nil {
		fmt.Println("encrypt error:", err)
		return
	}

	decrypted, err := encrypt.AESDecrypt(ciphertext, key)
	if err != nil {
		fmt.Println("decrypt error:", err)
		return
	}

	fmt.Println(string(decrypted))
	// Output: secret message
}

func ExampleRSAEncrypt() {
	pubKey, privKey, err := encrypt.GenerateRSAKeyPair(2048)
	if err != nil {
		fmt.Println("keygen error:", err)
		return
	}

	plaintext := []byte("hello RSA")
	ciphertext, err := encrypt.RSAEncrypt(plaintext, pubKey)
	if err != nil {
		fmt.Println("encrypt error:", err)
		return
	}

	decrypted, err := encrypt.RSADecrypt(ciphertext, privKey)
	if err != nil {
		fmt.Println("decrypt error:", err)
		return
	}

	fmt.Println(string(decrypted))
	// Output: hello RSA
}

func ExampleEncodeBase64() {
	// 加密后使用 Base64 编码便于传输
	key := []byte("12345678901234567890123456789012")
	ciphertext, _ := encrypt.AESEncrypt([]byte("secret"), key)

	encoded := encrypt.EncodeBase64(ciphertext)
	fmt.Println("base64 length > 0:", len(encoded) > 0)

	decoded, _ := encrypt.DecodeBase64(encoded)
	plaintext, _ := encrypt.AESDecrypt(decoded, key)
	fmt.Println(string(plaintext))
	// Output:
	// base64 length > 0: true
	// secret
}
