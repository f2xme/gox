# Crypto 包示例

本目录包含 `crypto` 包的可运行示例代码。

## 示例列表

每个示例为独立子目录（可单独 `go run`），避免同一目录多个 `main` 无法构建。

### 密码哈希

- **hash_bcrypt/** - bcrypt 密码哈希示例
- **hash_argon2/** - Argon2id 密码哈希示例（推荐）

### 数字签名

- **sign_ed25519/** - Ed25519 数字签名示例（推荐）
- **sign_ecdsa/** - ECDSA 数字签名示例
- **sign_rsa/** - RSA 数字签名示例

## 运行示例

在仓库根目录执行：

```bash
# bcrypt 密码哈希
go run ./examples/crypto/hash_bcrypt

# Argon2id 密码哈希
go run ./examples/crypto/hash_argon2

# Ed25519 数字签名
go run ./examples/crypto/sign_ed25519

# ECDSA 数字签名
go run ./examples/crypto/sign_ecdsa

# RSA 数字签名
go run ./examples/crypto/sign_rsa
```

## 算法选择建议

### 密码哈希

- **新项目推荐**: Argon2id（抗 GPU/ASIC 攻击，可配置内存和 CPU 成本）
- **兼容性需求**: bcrypt（广泛支持，成熟稳定）

### 数字签名

- **新项目推荐**: Ed25519（速度最快，密钥和签名最小）
- **标准兼容**: ECDSA（TLS、JWT 等场景）
- **传统系统**: RSA（需要与旧系统兼容时使用）
