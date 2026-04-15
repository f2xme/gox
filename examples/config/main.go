package main

import (
	"fmt"
	"os"

	"github.com/f2xme/gox/config/adapter/viper"
)

func main() {
	fmt.Println("=== Config 使用示例 ===\n")

	// 1. 创建示例配置文件
	configContent := `# 应用配置
app:
  name: "gox-example"
  version: "1.0.0"
  debug: true
  port: 8080

database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "secret"
  database: "myapp"
  max_connections: 100

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

features:
  - "auth"
  - "logging"
  - "metrics"

timeout: 30s
`

	// 写入临时配置文件
	configFile := "config-example.yaml"
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		fmt.Printf("创建配置文件失败: %v\n", err)
		return
	}
	defer os.Remove(configFile)

	fmt.Printf("已创建配置文件: %s\n\n", configFile)

	// 2. 加载配置
	fmt.Println("1. 加载配置:")
	cfg, err := viper.New(configFile)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}
	fmt.Println("配置加载成功")

	// 3. 读取字符串配置
	fmt.Println("\n2. 读取字符串配置:")
	appName := cfg.GetString("app.name")
	appVersion := cfg.GetString("app.version")
	fmt.Printf("应用名称: %s\n", appName)
	fmt.Printf("应用版本: %s\n", appVersion)

	// 4. 读取整数配置
	fmt.Println("\n3. 读取整数配置:")
	appPort := cfg.GetInt("app.port")
	dbPort := cfg.GetInt("database.port")
	maxConn := cfg.GetInt("database.max_connections")
	fmt.Printf("应用端口: %d\n", appPort)
	fmt.Printf("数据库端口: %d\n", dbPort)
	fmt.Printf("最大连接数: %d\n", maxConn)

	// 5. 读取布尔配置
	fmt.Println("\n4. 读取布尔配置:")
	debug := cfg.GetBool("app.debug")
	fmt.Printf("调试模式: %v\n", debug)

	// 6. 读取时间配置
	fmt.Println("\n5. 读取时间配置:")
	timeout := cfg.GetDuration("timeout")
	fmt.Printf("超时时间: %v\n", timeout)

	// 7. 读取字符串切片
	fmt.Println("\n6. 读取字符串切片:")
	features := cfg.GetStringSlice("features")
	fmt.Printf("启用的功能: %v\n", features)

	// 8. 读取 Map 配置
	fmt.Println("\n7. 读取 Map 配置:")
	dbConfig := cfg.GetStringMap("database")
	fmt.Printf("数据库配置: %v\n", dbConfig)

	// 9. 使用 Get 读取任意类型
	fmt.Println("\n8. 读取任意类型:")
	redisHost := cfg.Get("redis.host")
	redisPort := cfg.Get("redis.port")
	fmt.Printf("Redis 主机: %v (类型: %T)\n", redisHost, redisHost)
	fmt.Printf("Redis 端口: %v (类型: %T)\n", redisPort, redisPort)

	// 10. 读取不存在的配置（返回零值）
	fmt.Println("\n9. 读取不存在的配置:")
	notExist := cfg.GetString("not.exist.key")
	fmt.Printf("不存在的键: '%s' (空字符串)\n", notExist)

	notExistInt := cfg.GetInt("not.exist.int")
	fmt.Printf("不存在的整数: %d (零值)\n", notExistInt)

	// 11. 构建数据库连接字符串示例
	fmt.Println("\n10. 实际应用示例 - 构建数据库连接:")
	dbHost := cfg.GetString("database.host")
	dbPortVal := cfg.GetInt("database.port")
	dbUser := cfg.GetString("database.username")
	dbName := cfg.GetString("database.database")
	dsn := fmt.Sprintf("%s:***@tcp(%s:%d)/%s", dbUser, dbHost, dbPortVal, dbName)
	fmt.Printf("数据库连接字符串: %s\n", dsn)

	fmt.Println("\n配置示例完成")
}
