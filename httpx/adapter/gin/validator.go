package gin

import (
	"reflect"
	"strings"

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
		normalized := normalizeStringFields(value)
		return v.validator.Validate(normalized.Interface())
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if err := v.validateValue(value.Index(i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func normalizeStringFields(value reflect.Value) reflect.Value {
	if !value.CanSet() {
		normalized := reflect.New(value.Type()).Elem()
		normalized.Set(value)
		value = normalized
	}

	trimStringFields(value, map[uintptr]struct{}{})
	return value
}

func trimStringFields(value reflect.Value, seen map[uintptr]struct{}) {
	if !value.IsValid() {
		return
	}

	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			return
		}
		ptr := value.Pointer()
		if ptr != 0 {
			if _, ok := seen[ptr]; ok {
				return
			}
			seen[ptr] = struct{}{}
		}
		trimStringFields(value.Elem(), seen)
	case reflect.Interface:
		if value.IsNil() {
			return
		}
		trimInterfaceStringFields(value, seen)
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			fieldType := value.Type().Field(i)
			if fieldType.PkgPath != "" {
				continue
			}
			trimStringFields(value.Field(i), seen)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			trimStringFields(value.Index(i), seen)
		}
	case reflect.Map:
		trimMapStringFields(value, seen)
	case reflect.String:
		if value.CanSet() {
			value.SetString(strings.TrimSpace(value.String()))
		}
	}
}

func trimInterfaceStringFields(value reflect.Value, seen map[uintptr]struct{}) {
	elem := value.Elem()
	if value.CanSet() && elem.Kind() != reflect.Ptr {
		normalized := reflect.New(elem.Type()).Elem()
		normalized.Set(elem)
		trimStringFields(normalized, seen)
		if normalized.Type().AssignableTo(value.Type()) {
			value.Set(normalized)
		}
		return
	}
	trimStringFields(elem, seen)
}

func trimMapStringFields(value reflect.Value, seen map[uintptr]struct{}) {
	if value.IsNil() {
		return
	}

	for _, key := range value.MapKeys() {
		elem := value.MapIndex(key)
		normalized := reflect.New(elem.Type()).Elem()
		normalized.Set(elem)
		trimStringFields(normalized, seen)
		value.SetMapIndex(key, normalized)
	}
}

// Engine 实现 binding.StructValidator 接口，返回底层验证引擎
func (v *goxValidatorAdapter) Engine() any {
	return v.validator
}

// newGoxValidator 返回默认 gox validator 实例
func newGoxValidator() *validator.Validator {
	return validator.Default()
}
