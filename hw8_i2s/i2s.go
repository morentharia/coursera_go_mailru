package main

import (
	"errors"
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	v := reflect.ValueOf(data)

	outValue := reflect.ValueOf(out)

	if outValue.Kind() != reflect.Ptr {
		return errors.New("vOut.Kind() != reflect.Ptr")
	}

	outValue = outValue.Elem()

	if v.Kind() == reflect.Slice {
		if outValue.Kind() != reflect.Slice {
			return errors.New("vOut.Kind() != reflect.Slice")
		}
		valSlice := reflect.ValueOf(v.Interface())
		outSlice := reflect.MakeSlice(
			outValue.Type(),
			valSlice.Len(),
			valSlice.Len(),
		)
		for i := 0; i < v.Len(); i++ {
			err := i2s(
				valSlice.Index(i).Interface(),
				outSlice.Index(i).Addr().Interface(),
			)
			if err != nil {
				return err
			}
		}
		outValue.Set(outSlice)
	} else if v.Kind() == reflect.Map {

		if len(v.MapKeys()) == 0 {
			return fmt.Errorf("len(v.MapKeys()) == 0")
		}

		for _, key := range v.MapKeys() {
			outField := outValue.FieldByName(key.String())
			keyValue := v.MapIndex(key)

			switch reflect.TypeOf(keyValue.Interface()).Kind() {
			case reflect.String:
				val, ok := keyValue.Interface().(string)
				if !ok {
					return errors.New("fieldVal.Interface().(string) ")
				}
				if outField.Type().String() != "string" {
					return errors.New("!=string")
				}
				outField.SetString(val)

			case reflect.Float64:
				castVal, ok := keyValue.Interface().(float64)
				if !ok {
					return errors.New("fail fieldVal.Interface().(float64)")
				}
				if outField.Type().String() != "int" {
					return errors.New("!=int")
				}

				outField.SetInt(int64(castVal))

			case reflect.Bool:
				val, ok := keyValue.Interface().(bool)
				if !ok {
					return errors.New("fail fieldVal.Interface().(bool) ")
				}
				if outField.Type().String() != "bool" {
					return errors.New("!= bool")
				}
				outField.SetBool(val)

			case reflect.Map:
				outVal := reflect.New(outField.Type()).Elem()
				err := i2s(keyValue.Interface(), outVal.Addr().Interface())
				if err != nil {
					return err
				}
				outField.Set(outVal)

			case reflect.Slice:
				val := reflect.ValueOf(keyValue.Interface())
				outSlice := reflect.MakeSlice(outField.Type(), val.Len(), val.Len())
				for i := 0; i < val.Len(); i++ {
					err := i2s(val.Index(i).Interface(), outSlice.Index(i).Addr().Interface())
					if err != nil {
						return err
					}
				}
				outField.Set(outSlice)
			}
		}
	}
	return nil
}
