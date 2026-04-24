package gin

import (
	"reflect"
	"sync"

	"github.com/f2xme/gox/validator"
	"github.com/gin-gonic/gin/binding"
)

// goxValidatorAdapter 将 gox/validator 适配为 gin 的 binding.StructValidator 接口
type goxValidatorAdapter struct {
	validator *validator.Validator
	once      sync.Once
}

var _ binding.StructValidator = (*goxValidatorAdapter)(nil)

// ValidateStruct 实现 binding.StructValidator 接口
func (v *goxValidatorAdapter) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return v.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return v.validator.Validate(obj)
	default:
		return nil
	}
}

// Engine 实现 binding.StructValidator 接口，返回底层验证引擎
func (v *goxValidatorAdapter) Engine() any {
	return v.validator
}

// newGoxValidator 创建一个新的 gox validator 实例
func newGoxValidator() *validator.Validator {
	return validator.New()
}
