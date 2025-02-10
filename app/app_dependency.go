package app

import (
	"errors"
	"reflect"
)

type fieldInfo struct {
	FieldName   string
	FieldKind   reflect.Kind
	FieldType   reflect.Type
	IsDependAll bool
	DependType  ComponentType
	DependIds   []string
}

func resolveDependencies(component Component) (typeName string, types []ComponentType, instances []string, fields map[string]fieldInfo, err error) {
	if component == nil {
		return
	}
	fields = map[string]fieldInfo{}
	t := reflect.TypeOf(component)
	v := reflect.ValueOf(component)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		err = errors.New("ekit only support pointer receiver for component: " + t.Name())
		return
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeName = t.Name()
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			ekitTag := field.Tag.Get(TagEkit)
			if ekitTag == "" {
				continue
			}
			tagStr := EkitTagStr(ekitTag)
			tags, err1 := tagStr.Parse()
			if err1 != nil {
				err = errors.New(typeName + " 组件依赖解析失败：" + err1.Error())
				return
			}
			if tag, exist := FindTag(tags, TagComponent); exist {
				fieldKind := field.Type.Kind()
				fieldType := field.Type
				valueCount := tag.ValueCount()
				fi := fieldInfo{
					FieldName: field.Name,
					FieldKind: fieldKind,
					FieldType: fieldType,
				}
				switch fieldKind {
				case reflect.Struct:
					err = errors.New("ekit only support pointer receiver for component field: " + typeName + "." + field.Name)
					return
				case reflect.Ptr:
					if valueCount > 1 {
						err = errors.New("component only support nomore than 1 candidates" + field.Name)
						return
					}
					fieldType = fieldType.Elem()
				case reflect.Slice:
					fieldType = fieldType.Elem()
					if fieldType.Kind() != reflect.Ptr {
						err = errors.New("ekit only support pointer receiver for component slice field: " + typeName + "." + field.Name)
						return
					}
					fieldType = fieldType.Elem()

				case reflect.Map:
					if fieldType.Key().Kind() != reflect.String {
						err = errors.New("ekit only support string-key map: " + typeName + "." + field.Name)
						return
					}
					fieldType = fieldType.Elem()
					if fieldType.Kind() != reflect.Ptr {
						err = errors.New("ekit only support pointer receiver value for component map field: " + typeName + "." + field.Name)
						return
					}
					fieldType = fieldType.Elem()
				default:
					err = errors.New("unsupported Field kind:" + fieldKind.String())
					return
				}
				if valueCount == 0 {
					types = append(types, ComponentType(fieldType.Name()))
					fi.IsDependAll = true
					fi.DependType = ComponentType(fieldType.Name())
				} else {
					fi.DependIds = []string{}
					for _, value := range tag.Values {
						id := getComponentID(ComponentType(fieldType.Name()), value)
						instances = append(instances, id)
						fi.DependIds = append(fi.DependIds, id)
					}
				}
				fields[field.Name] = fi
			}
		}
	}
	return
}
