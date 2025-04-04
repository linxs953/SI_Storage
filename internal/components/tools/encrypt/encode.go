package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

const (
	iterations = 310000 // PBKDF2迭代次数，符合NIST标准
	keyLength  = 32     // AES-256密钥长度
	saltSize   = 16     // 盐值长度
	nonceSize  = 12     // GCM模式Nonce长度
)

// EncryptString 加密字符串：输入明文和密码，返回Base64加密字符串
func EncryptString(plaintext, password string) (string, error) {
	// 生成随机盐值
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 通过PBKDF2派生密钥
	key := pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha256.New)

	// 创建AES-GCM加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机Nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// 加密并拼接完整密文：salt + nonce + ciphertext
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	fullCiphertext := append(append(salt, nonce...), ciphertext...)

	// 转为Base64字符串
	return base64.StdEncoding.EncodeToString(fullCiphertext), nil
}

func DecryptString(encodedCiphertext, password string) (string, error) {
	// 解码Base64
	fullCiphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}

	// 验证密文长度有效性
	minLen := saltSize + nonceSize
	if len(fullCiphertext) < minLen {
		return "", errors.New("invalid ciphertext")
	}

	// 拆分盐值、Nonce和密文
	salt := fullCiphertext[:saltSize]
	nonce := fullCiphertext[saltSize : saltSize+nonceSize]
	ciphertext := fullCiphertext[saltSize+nonceSize:]

	// 派生密钥
	key := pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha256.New)

	// 创建AES-GCM解密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed: invalid password or corrupted data")
	}

	return string(plaintext), nil
}
