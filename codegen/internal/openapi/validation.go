package openapi

import (
	"fmt"
	"reflect"
)

const (
	fieldTag         = "validate"
	fieldTagRequired = "required"
)

// Validator can be used to validate individual objects.
type Validator interface {
	// Validate takes the entire document in case it is needed to look up references.
	Validate(*Document) error
}

// Validate runs through the validation requirements and enforces them.
func Validate(obj any, doc *Document) error {
	if obj == nil {
		return nil
	}

	objectType := reflect.TypeOf(obj)
	objectValue := reflect.ValueOf(obj)

	if objectType.Kind() == reflect.Pointer {
		objectType = objectType.Elem()
		objectValue = objectValue.Elem()
	}

	if objectType.Kind() != reflect.Struct {
		return nil // non struct types do not get validated.
	}

	for i := 0; i < objectType.NumField(); i++ {
		field := objectType.Field(i)
		validationTag := field.Tag.Get(fieldTag)
		fieldValue := objectValue.Field(i)

		if validationTag == fieldTagRequired {
			if fieldValue.IsZero() && field.Type.Kind() != reflect.Struct {
				return fmt.Errorf("%q is required", field.Name)
			}
			if fieldValue.Kind() == reflect.Slice && fieldValue.Len() == 0 {
				return fmt.Errorf("%q is required", field.Name)
			}
			if fieldValue.Kind() == reflect.Map && fieldValue.Len() == 0 {
				return fmt.Errorf("%q is required", field.Name)
			}
		}

		if validator, ok := getValidator(fieldValue); ok {
			if err := validator.Validate(doc); err != nil {
				return fmt.Errorf("invalid %q: %w", field.Name, err)
			}
			continue
		}

		switch field.Type.Kind() {
		case reflect.Slice:
			if !underlyingTypeNeedsValidation(field.Type.Elem()) {
				continue
			}
			canBeNil := false
			switch field.Type.Elem().Kind() {
			case reflect.Slice, reflect.Pointer:
				canBeNil = true
			}
			for j := 0; j < fieldValue.Len(); j++ {
				item := fieldValue.Index(j)
				if canBeNil && item.IsNil() {
					return fmt.Errorf("empty value at index %d", j)
				}
				if err := Validate(item.Addr().Interface(), doc); err != nil {
					return fmt.Errorf("invalid %s at index %d: %w", field.Type.Name(), j, err)
				}
			}
		case reflect.Map:
			if !underlyingTypeNeedsValidation(field.Type.Elem()) {
				continue
			}
			iterator := fieldValue.MapRange()
			for iterator.Next() {
				if err := Validate(iterator.Value().Interface(), doc); err != nil {
					return fmt.Errorf("invalid %s value for key %q: %w", iterator.Value().Type().Name(), iterator.Key(), err)
				}
			}
		case reflect.Struct:
			if err := Validate(fieldValue.Addr().Interface(), doc); err != nil {
				return fmt.Errorf("invalid %q: %w", field.Type.Name(), err)
			}
		}
	}

	return nil
}

func underlyingTypeNeedsValidation(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map:
		t = t.Elem()
	}

	return t.Kind() == reflect.Struct
}

func getValidator(t reflect.Value) (Validator, bool) {
	validator, ok := t.Interface().(Validator)
	if ok {
		return validator, true
	}

	if t.CanAddr() {
		validator, ok = t.Addr().Interface().(Validator)
		if ok {
			return validator, true
		}
	}

	if t.Kind() == reflect.Pointer {
		validator, ok = t.Elem().Interface().(Validator)
		if ok {
			return validator, true
		}
	}

	return nil, false
}
