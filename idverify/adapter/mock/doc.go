// Package mock 提供 idverify.Verifier 的内存实现，用于本地开发与测试。
//
// 默认所有合法请求核验通过。可通过选项指定姓名触发不匹配 / 查无记录。
package mock
