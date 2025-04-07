package configs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// equal 函数支持将驼峰式变量名转换为下划线分隔形式，并进行不区分大小写的比较
func equalFieldName(fieldName, key string) bool {
	// 将驼峰式变量名转换为下划线分隔形式
	normalizedFieldName := camelToSnake(fieldName)
	// 进行不区分大小写的比较
	return strings.EqualFold(normalizedFieldName, key)
}

// camelToSnake 将驼峰式变量名转换为下划线分隔形式
func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// SetConfigValue 设置配置值
func SetConfigValue(cfg interface{}, key, value string) error {
	keys := strings.Split(key, ".")
	if len(keys) == 0 {
		return fmt.Errorf("invalid configuration key format")
	}

	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByNameFunc(func(s string) bool {
		return equalFieldName(s, string(keys[0]))
	})

	if !field.IsValid() {
		return fmt.Errorf("unknown configuration section: %s", keys[0])
	}

	if len(keys) == 1 {
		if !field.CanSet() {
			return fmt.Errorf("cannot set configuration key: %s", keys[0])
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Int:
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v", keys[0], value)
			}
			field.SetInt(int64(intValue))
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v", keys[0], value)
			}
			field.SetBool(boolValue)
		case reflect.Float64:
			floatValue, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value for %s: %v", keys[0], value)
			}
			field.SetFloat(floatValue)
		case reflect.Array, reflect.Slice:
			// 处理数组或切片类型
			elements := strings.Split(value, ",")
			arrayLen := field.Len()
			if field.Kind() == reflect.Slice {
				// 如果是切片，先创建足够长度的切片
				field.Set(reflect.MakeSlice(field.Type(), len(elements), len(elements)))
				arrayLen = len(elements)
			} else if len(elements) != arrayLen {
				// 对于数组，检查长度是否匹配
				return fmt.Errorf("array length mismatch for %s: expected %d, got %d",
					keys[0], arrayLen, len(elements))
			}

			// 获取数组元素的类型
			elemType := field.Type().Elem()
			for i := 0; i < arrayLen && i < len(elements); i++ {
				elemValue := elements[i]
				elem := reflect.New(elemType).Elem()

				switch elem.Kind() {
				case reflect.String:
					elem.SetString(elemValue)
				case reflect.Int:
					intValue, err := strconv.Atoi(elemValue)
					if err != nil {
						return fmt.Errorf("invalid array element value for %s[%d]: %v", keys[0], i, elemValue)
					}
					elem.SetInt(int64(intValue))
				case reflect.Bool:
					boolValue, err := strconv.ParseBool(elemValue)
					if err != nil {
						return fmt.Errorf("invalid array element value for %s[%d]: %v", keys[0], i, elemValue)
					}
					elem.SetBool(boolValue)
				case reflect.Float64:
					floatValue, err := strconv.ParseFloat(elemValue, 64)
					if err != nil {
						return fmt.Errorf("invalid array element value for %s[%d]: %v", keys[0], i, elemValue)
					}
					elem.SetFloat(floatValue)
				default:
					return fmt.Errorf("unsupported array element type for %s: %v", keys[0], elemType)
				}

				field.Index(i).Set(elem)
			}
		default:
			return fmt.Errorf("unsupported configuration key type: %s", keys[0])
		}

		return nil
	}

	if field.Kind() != reflect.Struct {
		return fmt.Errorf("cannot set nested configuration key: %s", key)
	}

	return SetConfigValue(field.Addr().Interface(), strings.Join(keys[1:], "."), value)
}

// GetConfigValue 获取配置值
func GetConfigValue(cfg interface{}, key string) (*reflect.Value, error) {
	keys := strings.Split(key, ".")
	if len(keys) == 0 {
		return nil, fmt.Errorf("invalid configuration key format")
	}

	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByNameFunc(func(s string) bool {
		return equalFieldName(s, string(keys[0]))
	})

	if !field.IsValid() {
		return nil, fmt.Errorf("unknown configuration section: %s", keys[0])
	}

	if len(keys) == 1 {
		return &field, nil
	}
	// if field.Kind() != reflect.Struct {
	// 	return fmt.Errorf("cannot set nested configuration key: %s", key)
	// }

	return GetConfigValue(field.Addr().Interface(), strings.Join(keys[1:], "."))
}
