module github.com/f2xme/gox/oss/adapter/aliyun

go 1.25.7

require (
	github.com/aliyun/aliyun-oss-go-sdk v3.0.2+incompatible
	github.com/f2xme/gox v0.15.5
)

require (
	golang.org/x/time v0.15.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/f2xme/gox => ../../..
