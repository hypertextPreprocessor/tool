package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"flag"
	"path"
	"os"
)
// PKCS7UnPadding 还原逻辑
func PKCS7UnPadding(origData []byte) ([]byte, error) {
    length := len(origData)
    if length == 0 {
        return nil, errors.New("data is empty")
    }
    // 读取最后一个字节的值
    padding := int(origData[length-1])
    // 简单校验填充是否合法
    if padding > length || padding == 0 {
        return nil, errors.New("invalid padding")
    }
    return origData[:(length - padding)], nil
}

// PKCS7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// EncryptAESCBC 加密
func EncryptAESCBC(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	padded := pkcs7Padding(plaintext, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(padded))
	iv := ciphertext[:aes.BlockSize]

	// 填充随机 IV
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
// DecryptAESCBC 解密
func DecryptAESCBC(cryptoText string, key []byte) ([]byte, error) {
	// 1. Base64 解码
	ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 2. 校验长度：至少要包含一个 IV 的长度 (16字节)
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// 3. 提取 IV 和 真正的密文
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// 4. CBC 模式解密
	// CBC 模式要求长度必须是 BlockSize 的倍数
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// 5. 去除 PKCS7 填充
	return PKCS7UnPadding(ciphertext)
}

// 修改后的加密函数，返回二进制 []byte 而不是 Base64 字符串
func EncryptAESCBCBinary(plaintext []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil { return nil, err }
    padded := pkcs7Padding(plaintext, aes.BlockSize)
    ciphertext := make([]byte, aes.BlockSize+len(padded))
    iv := ciphertext[:aes.BlockSize]
    io.ReadFull(rand.Reader, iv)
    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(ciphertext[aes.BlockSize:], padded)
    return ciphertext, nil // 直接返回二进制
}
// DecryptAESCBCBinary 解密二进制密文 (与 EncryptAESCBCBinary 对应)
func DecryptAESCBCBinary(ciphertext []byte, key []byte) ([]byte, error) {
	// 1. 检查长度：密文至少要包含一个 IV 的长度 (16字节)
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 2. 提取 IV (前16字节) 和 剩余的加密数据
	iv := ciphertext[:aes.BlockSize]
	encryptedData := ciphertext[aes.BlockSize:]

	// 3. 校验长度：CBC 模式要求数据长度必须是 BlockSize 的倍数
	if len(encryptedData)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	// 4. 解密
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encryptedData, encryptedData)

	// 5. 去除 PKCS7 填充并返回结果
	return PKCS7UnPadding(encryptedData)
}
// AES-CTR 模式流加密示例 (无需填充，完美流式)
func StreamEncryptCTR(src io.Reader, dst io.Writer, key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	io.ReadFull(rand.Reader, iv)
	dst.Write(iv) // 存储 IV

	stream := cipher.NewCTR(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: dst}

	// 核心：直接拷贝，流式加密
	_, err = io.Copy(writer, src)
	return err
}
// StreamDecryptCTR 使用 CTR 模式解密
// src: 输入流 (例如 os.File，包含 IV + 密文)
// dst: 输出流 (解密后的数据写入的目标)
// key: 16, 24 或 32 字节的密钥
func StreamDecryptCTR(src io.Reader, dst io.Writer, key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// 1. 读取头部 16 字节的 IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(src, iv); err != nil {
		return err
	}

	// 2. 创建 CTR 流解密器 (与加密逻辑一致)
	stream := cipher.NewCTR(block, iv)
	reader := &cipher.StreamReader{S: stream, R: src}

	// 3. 直接拷贝，流式解密
	// io.Copy 会自动从 reader 读取密文并进行异或运算，将明文写入 dst
	_, err = io.Copy(dst, reader)
	return err
}
func main() {
	var sFlag = flag.String("type","text","请输入加密类型(text | file | bulkFile)")
	var sVal = flag.String("val","","输入文件路径或明文信息")
	flag.Parse()
	key := []byte("QqH3+847'39(8#37djOvhfjlsi%kf@=]") // 32字节密钥 (AES-256)

	if *sFlag == "text" {
		
		if(*sVal == ""){
			fmt.Println("[val]参数不能为空")
			return
		}else{
			text := []byte(*sVal)
			encrypted, _ := EncryptAESCBC(text, key)
			fmt.Println("Base64 密文:", encrypted)
		}

	}else if *sFlag == "file" {
		bs,_  := os.ReadFile(*sVal)
		if(*sVal == ""){
			fmt.Println("[val]参数不能为空")
			return
		}else{
			encrypted,err := EncryptAESCBCBinary(bs, key)
			if err != nil { return }
			dstPath,_ := os.Getwd()
			if os.WriteFile(path.Join(dstPath,"file"), encrypted, 0644) !=nil {
				fmt.Println("error")
				return
			}
			
			//fmt.Println("Base64 密文:", encrypted)
		}
	}else if *sFlag == "bulkFile" {
		if *sVal == "" {
			fmt.Println("[val]参数不能为空")
		}else{
			inputFile,_ := os.Open(*sVal)
			defer inputFile.Close()
			outputFile,_ := os.Create("Kaiyuan")
			defer outputFile.Close()
			err := StreamEncryptCTR(inputFile,outputFile,key)
			//err := StreamDecryptCTR(inputFile,outputFile,key)
			if err != nil {
				fmt.Println("加密失败了",err)
			}
		}
		
	}else{
		fmt.Printf("[%s]参数不正确 使用: type | text",*sFlag)
	}



	// // 解密
	// decrypted, err := DecryptAESCBC(encrypted, key)
	// if err != nil {
	// 	fmt.Println("解密失败:", err)
	// 	return
	// }
	// fmt.Println("解密后:", string(decrypted))
}