// Package idverify 提供身份证姓名与证件号二要素核验的统一抽象。
//
// 业务只依赖 [Verifier] 接口；具体厂商实现位于 adapter 子包：
//
//   - mock：内存实现，便于本地与单测
//   - baidu：百度人脸 person/idmatch
//   - aliyun：阿里云 Cloudauth Id2MetaVerify（独立 module，含 SDK 依赖）
//
// # 语义约定
//
//   - 业务不一致（姓名不匹配、证件无效等）：返回 Result{Matched:false} 且 error 为 nil
//   - 配置缺失、网络错误、上游系统错误：返回 error（可用 errors.Is 判断 [ErrNotConfigured] 等）
//
// # 快速开始
//
//	import (
//		"context"
//
//		"github.com/f2xme/gox/idverify"
//		"github.com/f2xme/gox/idverify/adapter/mock"
//	)
//
//	v, _ := mock.New()
//	res, err := v.Verify(context.Background(), idverify.Request{
//		Name: "张三", IDNumber: "110101199001011234",
//	})
//	if err != nil {
//		// 系统错误
//	}
//	if !res.Matched {
//		// 业务不匹配，查看 res.ErrorCode
//	}
package idverify
