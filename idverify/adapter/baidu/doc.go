// Package baidu 实现基于百度智能云人脸身份证比对接口的二要素核验。
//
// 接口：POST /rest/2.0/face/v3/person/idmatch
//
// error_code 0 表示同一人；222351 姓名不匹配；222022 证件号无效。
package baidu
