package gin

import (
	"reflect"

	"github.com/f2xme/gox/validator"
	"github.com/gin-gonic/gin/binding"
)

// goxValidatorAdapter 将 gox/validator 适配为 gin 的 binding.StructValidator 接口
type goxValidatorAdapter struct {
	validator *validator.Validator
}

var _ binding.StructValidator = (*goxValidatorAdapter)(nil)

// ValidateStruct 实现 binding.StructValidator 接口
func (v *goxValidatorAdapter) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}

	return v.validateValue(reflect.ValueOf(obj))
}

func (v *goxValidatorAdapter) validateValue(value reflect.Value) error {
	if !value.IsValid() {
		return nil
	}

	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return v.validateValue(value.Elem())
	case reflect.Struct:
		return v.validator.Validate(value.Interface())
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if err := v.validateValue(value.Index(i)); err != nil {
				return err
			}
		}
	}

	return nil
}

// Engine 实现 binding.StructValidator 接口，返回底层验证引擎
func (v *goxValidatorAdapter) Engine() any {
	return v.validator
}

// newGoxValidator 返回默认 gox validator 实例
func newGoxValidator() *validator.Validator {
	return validator.Default()
}
