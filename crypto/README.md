# crypto

密码学工具包，提供密码哈希和数字签名功能。

## 功能

### 密码哈希

- **bcrypt**: 使用 bcrypt 算法进行密码哈希
- **Argon2id**: 使用 Argon2id 算法进行密码哈希（推荐用于新项目）

### 数字签名

- **RSA**: RSA-PSS 签名（2048+ 位）
- **ECDSA**: ECDSA 签名（P-256 曲线）
- **Ed25519**: Ed25519 签名（推荐用于新项目）

## 安装

```bash
go get github.com/f2xme/gox/crypto
```

## 使用示例

### 密码哈希 - bcrypt

```go
import "github.com/f2xme/gox/crypto"

// 哈希密码
hash, err := crypto.HashPassword("mypassword")
if err != nil {
    log.Fatal(err)
}

// 验证密码
if crypto.VerifyPassword("mypassword", hash) {
    fmt.Println("密码正确")
}
```

### 密码哈希 - Argon2id

```go
// 使用默认参数
hash, err := crypto.HashPasswordArgon2("mypassword", nil)
if err != nil {
    log.Fatal(err)
}

// 验证密码
valid, err := crypto.VerifyPasswordArgon2("mypassword", hash)
if err != nil {
    log.Fatal(err)
}
if valid {
    fmt.Println("密码正确")
}

// 自定义参数
params := &crypto.Argon2Params{
    Memory:      128 * 1024, // 128 MB
    Iterations:  4,
    Parallelism: 4,
    SaltLength:  16,
    KeyLength:   32,
}
hash, err = crypto.HashPasswordArgon2("mypassword", params)
```

### 数字签名 - RSA

```go
// 生成密钥对
privateKey, err := crypto.GenerateRSAKeyPair(2048)
if err != nil {
    log.Fatal(err)
}

data := []byte("message to sign")

// 签名
signature, err := crypto.SignRSA(privateKey, data)
if err != nil {
    log.Fatal(err)
}

// 验证
err = crypto.VerifyRSA(&privateKey.PublicKey, data, signature)
if err != nil {
    fmt.Println("签名无效")
} else {
    fmt.Println("签名有效")
}
```

### 数字签名 - ECDSA

```go
// 生成密钥对
privateKey, err := crypto.GenerateECDSAKeyPair()
if err != nil {
    log.Fatal(err)
}

data := []byte("message to sign")

// 签名
signature, err := crypto.SignECDSA(privateKey, data)
if err != nil {
    log.Fatal(err)
}

// 验证
if crypto.VerifyECDSA(&privateKey.PublicKey, data, signature) {
    fmt.Println("签名有效")
}
```

### 数字签名 - Ed25519

```go
// 生成密钥对
publicKey, privateKey, err := crypto.GenerateEd25519KeyPair()
if err != nil {
    log.Fatal(err)
}

data := []byte("message to sign")

// 签名
signature := crypto.SignEd25519(privateKey, data)

// 验证
if crypto.VerifyEd25519(publicKey, data, signature) {
    fmt.Println("签名有效")
}
```

## 安全建议

### 密码哈希

- 对于新项目，推荐使用 **Argon2id**（更安全，可配置内存和 CPU 成本）
- bcrypt 适用于需要兼容现有系统的场景
- 永远不要存储明文密码
- 使用足够的计算成本参数以抵御暴力破解

### 数字签名

- 对于新项目，推荐使用 **Ed25519**（更快，更安全，密钥更小）
- RSA 密钥长度至少 2048 位（推荐 3072 或 4096 位）
- ECDSA 使用 P-256 曲线（NIST 标准）
- 妥善保管私钥，永远不要泄露

## 依赖

- `golang.org/x/crypto` - Go 官方扩展加密库
