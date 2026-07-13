package mock

import (
	"fmt"
	"reflect"
	"time"

	"github.com/f2xme/gox/payment"
)

var timeType = reflect.TypeOf(time.Time{})

type cloneVisit struct {
	typeOf   reflect.Type
	kind     reflect.Kind
	pointer  uintptr
	length   int
	capacity int
}

func clonePaymentRecord(record PaymentRecord) PaymentRecord {
	return PaymentRecord{
		Order:  cloneOrder(record.Order),
		Result: clonePaymentResult(record.Result),
		Status: record.Status,
		PaidAt: cloneTime(record.PaidAt),
	}
}

func cloneRefundRecord(record RefundRecord) RefundRecord {
	return RefundRecord{
		Request: record.Request,
		Result:  cloneRefundResult(record.Result),
	}
}

func cloneOrder(order payment.Order) payment.Order {
	order.ExpireAt = cloneTime(order.ExpireAt)
	order.Extra = cloneMap(order.Extra)
	return order
}

func clonePaymentResult(result payment.PaymentResult) payment.PaymentResult {
	result.Extra = cloneMap(result.Extra)
	return result
}

func cloneRefundResult(result payment.RefundResult) payment.RefundResult {
	result.RefundAt = cloneTime(result.RefundAt)
	return result
}

func cloneMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		cloned[key] = cloneAny(item)
	}
	return cloned
}

func validateExtra(value map[string]any) error {
	if err := validateCloneValue(reflect.ValueOf(value), make(map[cloneVisit]struct{})); err != nil {
		return fmt.Errorf("%w: invalid mock extra: %v", payment.ErrInvalidRequest, err)
	}
	return nil
}

func validateCloneValue(value reflect.Value, visiting map[cloneVisit]struct{}) error {
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return validateCloneValue(value.Elem(), visiting)
	case reflect.Map, reflect.Slice, reflect.Pointer:
		if value.IsNil() {
			return nil
		}
		visit := cloneVisit{typeOf: value.Type(), kind: value.Kind(), pointer: value.Pointer()}
		if value.Kind() == reflect.Slice {
			visit.length = value.Len()
			visit.capacity = value.Cap()
		}
		if _, exists := visiting[visit]; exists {
			return fmt.Errorf("cyclic %s value", value.Kind())
		}
		visiting[visit] = struct{}{}
		defer delete(visiting, visit)
		switch value.Kind() {
		case reflect.Map:
			iterator := value.MapRange()
			for iterator.Next() {
				if err := validateCloneValue(iterator.Key(), visiting); err != nil {
					return err
				}
				if err := validateCloneValue(iterator.Value(), visiting); err != nil {
					return err
				}
			}
		case reflect.Slice:
			for i := range value.Len() {
				if err := validateCloneValue(value.Index(i), visiting); err != nil {
					return err
				}
			}
		case reflect.Pointer:
			return validateCloneValue(value.Elem(), visiting)
		}
		return nil
	case reflect.Array:
		for i := range value.Len() {
			if err := validateCloneValue(value.Index(i), visiting); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		if value.Type() == timeType {
			return nil
		}
		for i := range value.NumField() {
			field := value.Type().Field(i)
			if field.PkgPath != "" {
				if typeContainsReference(field.Type, make(map[reflect.Type]struct{})) {
					return fmt.Errorf("struct %s has unexported reference field %s", value.Type(), field.Name)
				}
				continue
			}
			if err := validateCloneValue(value.Field(i), visiting); err != nil {
				return err
			}
		}
		return nil
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return fmt.Errorf("unsupported %s value", value.Kind())
	default:
		return nil
	}
}

func typeContainsReference(value reflect.Type, visiting map[reflect.Type]struct{}) bool {
	if value == timeType {
		return false
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Map, reflect.Slice, reflect.Pointer,
		reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return true
	case reflect.Array:
		return typeContainsReference(value.Elem(), visiting)
	case reflect.Struct:
		if _, exists := visiting[value]; exists {
			return false
		}
		visiting[value] = struct{}{}
		defer delete(visiting, value)
		for i := range value.NumField() {
			if typeContainsReference(value.Field(i).Type, visiting) {
				return true
			}
		}
	}
	return false
}

func cloneAny(value any) any {
	if value == nil {
		return nil
	}
	return cloneReflectValue(reflect.ValueOf(value)).Interface()
}

func cloneReflectValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		cloned := reflect.New(value.Type()).Elem()
		cloned.Set(cloneReflectValue(value.Elem()))
		return cloned
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		cloned := reflect.MakeMapWithSize(value.Type(), value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			cloned.SetMapIndex(cloneReflectValue(iterator.Key()), cloneReflectValue(iterator.Value()))
		}
		return cloned
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		cloned := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := range value.Len() {
			cloned.Index(i).Set(cloneReflectValue(value.Index(i)))
		}
		return cloned
	case reflect.Array:
		cloned := reflect.New(value.Type()).Elem()
		for i := range value.Len() {
			cloned.Index(i).Set(cloneReflectValue(value.Index(i)))
		}
		return cloned
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		cloned := reflect.New(value.Type().Elem())
		cloned.Elem().Set(cloneReflectValue(value.Elem()))
		return cloned
	case reflect.Struct:
		if value.Type() == timeType {
			return value
		}
		cloned := reflect.New(value.Type()).Elem()
		cloned.Set(value)
		for i := range value.NumField() {
			if value.Type().Field(i).PkgPath == "" {
				cloned.Field(i).Set(cloneReflectValue(value.Field(i)))
			}
		}
		return cloned
	default:
		return value
	}
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
