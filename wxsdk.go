package wx

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	Block_size  = 32
	BASE_STRING = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

//生成随机字符，使用字母和数字
func getRandomString(length int) string {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	lenth := len(BASE_STRING)
	buffer := bytes.NewBufferString("")
	for i := 0; i < length; i++ {
		buffer.WriteString(string(BASE_STRING[r.Intn(lenth)]))
	}

	return buffer.String()
}

func PostToWx(theUrl string, data interface{}, result interface{}) error {
	content, _ := json.Marshal(data)
	resp, err := http.Post(theUrl, "application/json;charset=utf-8", strings.NewReader(string(content)))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	json.Unmarshal(body, result)
	return nil
}

//-------------网络字节序列----------------//

// 生成4个字节的网络字节序
func numberToBytesOrder(number int) []byte {
	var orderBytes []byte
	orderBytes = make([]byte, 4, 4)
	orderBytes[3] = byte(number & 0xFF)
	orderBytes[2] = byte(number >> 8 & 0xFF)
	orderBytes[1] = byte(number >> 16 & 0xFF)
	orderBytes[0] = byte(number >> 24 & 0xFF)

	return orderBytes
}

// 还原4个字节的网络字节序
func bytesOrderToNumber(orderBytes []byte) int {
	var number int = 0

	for i := 0; i < 4; i++ {
		number <<= 8
		number |= int(orderBytes[i] & 0xff)
	}
	return number
}

//-------------------AES----------------------//

type AES struct {
	key       []byte
	block     cipher.Block
	blockSize int
}

func (this *AES) Init(key []byte) {
	this.key = key
	block, err := aes.NewCipher(this.key)
	if err != nil {
		return
	}
	this.block = block
	this.blockSize = this.block.BlockSize()
}

func (this *AES) encrypt(sourceText []byte) []byte {
	sourceText = PKCS7Padding(sourceText, this.blockSize)

	blockModel := cipher.NewCBCEncrypter(this.block, this.key[:this.blockSize])

	ciphertext := make([]byte, len(sourceText))

	blockModel.CryptBlocks(ciphertext, sourceText)
	return ciphertext
}

func (this *AES) decrypt(encryptedText []byte) []byte {
	blockMode := cipher.NewCBCDecrypter(this.block, this.key[:this.blockSize])
	origData := make([]byte, len(encryptedText))
	blockMode.CryptBlocks(origData, encryptedText)
	origData = PKCS7UnPadding(origData, this.blockSize)
	return origData
}

//---------------------------PKCS7-----------------------------//
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(plantText []byte, blockSize int) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

//-------------------------------SHA1------------------------------//
func str2sha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func makeSignature(token string, timestamp, nonce string, msg string) string {
	sl := []string{token, timestamp, nonce, msg}
	sort.Strings(sl)
	return str2sha1(strings.Join(sl, ""))
}

func MakeSignatureForJs(token string, timestamp, nonce string) string {
	sl := []string{token, timestamp, nonce}
	sort.Strings(sl)
	return str2sha1(strings.Join(sl, ""))
}
