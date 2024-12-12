package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
    valOut := reflect.ValueOf(out)
    if valOut.Kind() != reflect.Ptr || valOut.IsNil() {
        return fmt.Errorf("out must be a non-nil pointer")
    }
    valOut = valOut.Elem()
    if valOut.Kind() != reflect.Struct {
        return fmt.Errorf("out must be a pointer to a struct")
    }
    valData, ok := data.(map[string]interface{})
    if !ok {
        return fmt.Errorf("data must be a map[string]interface{}")
    }
    for i := 0; i < valOut.NumField(); i++ {
        field := valOut.Field(i)
        fieldName := valOut.Type().Field(i).Name
        if value, exists := valData[fieldName]; exists {
            if field.CanSet() {
                valueReflect := reflect.ValueOf(value)
                if valueReflect.Type().AssignableTo(field.Type()) {
                    field.Set(valueReflect)
                } else {
					switch field.Kind() {
						case reflect.Int:
							if valueReflect.Kind() == reflect.Float64 {
								field.SetInt(int64(valueReflect.Float()))
							}
						case reflect.Float64:
							if valueReflect.Kind() == reflect.Int {
								field.SetFloat(float64(valueReflect.Int()))
							}
						}
				}
            }
        }
    }

    return nil
}

func main() {
	fmt.Println("Let's go ")
}