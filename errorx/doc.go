// Package errorx 提供增强的错误处理功能，支持错误码、堆栈跟踪和错误分类。
//
// # 功能特性
//
//   - 错误码和元数据：为错误附加结构化的错误码和上下文信息
//   - 错误分类（Kind）：使用预定义的错误类型进行分类（验证、未找到、未授权等）
//   - 堆栈跟踪：自动捕获错误发生时的调用栈信息
//   - 错误码注册表：支持国际化的错误消息管理
//   - 链式调用：流畅的 API 设计，支持链式设置错误属性
//   - 线程安全：所有函数和类型都可以在多个 goroutine 中并发使用
//
// # 快速开始
//
// 创建基本错误：
//
//	err := errorx.New("用户不存在")
//
// 创建带错误码和分类的错误：
//
//	err := errorx.New("用户不存在").
//		WithCode("USER_NOT_FOUND").
//		WithKind(errorx.KindNotFound).
//		WithMetadata("user_id", 12345)
//
// 包装已有错误：
//
//	err := errorx.Wrap(dbErr, "查询用户失败")
//
// 检查错误类型：
//
//	if errorx.IsKind(err, errorx.KindValidation) {
//		// 处理验证错误
//	}
//
// 获取错误码：
//
//	code := errorx.GetCode(err)
//
// 格式化错误（包含堆栈）：
//
//	fmt.Println(errorx.FormatWithStack(err))
//
// # 错误分类
//
// 内置的 Kind 类型：
//
//   - KindValidation: 验证错误（输入不合法）
//   - KindNotFound: 资源未找到
//   - KindConflict: 资源冲突（如重复键）
//   - KindUnauthorized: 未认证
//   - KindForbidden: 未授权（已认证但无权限）
//   - KindInternal: 内部服务器错误
//   - KindTimeout: 超时错误
//   - KindRetryable: 可重试的错误
//
// # 错误码注册表
//
// 注册错误码及其国际化消息：
//
//	errorx.Register("USER_NOT_FOUND", "zh", "用户不存在")
//	errorx.Register("USER_NOT_FOUND", "en", "User not found")
//	errorx.SetDefaultLang("zh")
//
// 使用注册的错误码创建错误：
//
//	err := errorx.NewCodeWithLang("USER_NOT_FOUND", "zh")
//
// # 最佳实践
//
// 在函数边界创建 errorx：
//
//	func GetUser(id int) (*User, error) {
//		user, err := db.FindUser(id)
//		if err == sql.ErrNoRows {
//			return nil, errorx.New("用户不存在").
//				WithCode("USER_NOT_FOUND").
//				WithKind(errorx.KindNotFound).
//				WithMetadata("user_id", id)
//		}
//		if err != nil {
//			return nil, errorx.Wrap(err, "查询用户失败")
//		}
//		return user, nil
//	}
//
// 使用 Kind 进行错误分类而不是字符串匹配：
//
//	// 推荐
//	if errorx.IsKind(err, errorx.KindValidation) {
//		return http.StatusBadRequest
//	}
//
//	// 不推荐
//	if strings.Contains(err.Error(), "validation") {
//		return http.StatusBadRequest
//	}
//
// 在日志中记录堆栈信息：
//
//	if err != nil {
//		log.Error(errorx.FormatWithStack(err))
//	}
package errorx
