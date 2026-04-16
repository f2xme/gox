# Crypto 包示例

本目录包含 `crypto` 包的可运行示例代码。

## 示例列表

### 密码哈希

- **hash_bcrypt.go** - bcrypt 密码哈希示例
- **hash_argon2.go** - Argon2id 密码哈希示例（推荐）

### 数字签名

- **sign_ed25519.go** - Ed25519 数字签名示例（推荐）
- **sign_ecdsa.go** - ECDSA 数字签名示例
- **sign_rsa.go** - RSA 数字签名示例

## 运行示例

在 `crypto/examples` 目录下运行：

```bash
# bcrypt 密码哈希
go run hash_bcrypt.go

# Argon2id 密码哈希
go run hash_argon2.go

# Ed25519 数字签名
go run sign_ed25519.go

# ECDSA 数字签名
go run sign_ecdsa.go

# RSA 数字签名
go run sign_rsa.go
```

## 算法选择建议

### 密码哈希

- **新项目推荐**: Argon2id（抗 GPU/ASIC 攻击，可配置内存和 CPU 成本）
- **兼容性需求**: bcrypt（广泛支持，成熟稳定）

### 数字签名

- **新项目推荐**: Ed25519（速度最快，密钥和签名最小）
- **标准兼容**: ECDSA（TLS、JWT 等场景）
- **传统系统**: RSA（需要与旧系统兼容时使用）
