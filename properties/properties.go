package properties

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type propertiesFormater struct {
	UseJsonFlag bool         //在将结构体序列化为properties的时候,字段的名称是否使用属性上的jsonTag
	IgnoreFlags []IgnoreFlag //对于某些敏感字段,如果需要
	PrintEmpty  bool         //默认情况下,如果属性值为nil,则属性不会输出到文件
}
type IgnoreFlag struct {
	Flag  string
	Value string
}

func DefaultPropertiesFormater() *propertiesFormater {
	return &propertiesFormater{}
}
func NewPropertiesFormater(useJsonFlag bool, ignoreFlags []IgnoreFlag, printEmpty bool) *propertiesFormater {
	return &propertiesFormater{
		IgnoreFlags: ignoreFlags, UseJsonFlag: useJsonFlag, PrintEmpty: printEmpty,
	}
}

func (p propertiesFormater) Translate(rv reflect.Value) (properties []string, err error) {
	properties, _, err = p.translate(rv)
	return
}

func (p propertiesFormater) translate(rv reflect.Value) (properties []string, hasNext bool, err error) {
	if rv.Kind() == reflect.Bool {
		properties = append(properties, fmt.Sprintf("=%s", strconv.FormatBool(rv.Bool())))
	} else if rv.Kind() == reflect.String {
		properties = append(properties, fmt.Sprintf("=%s", rv.String()))
	} else if reflect.Int <= rv.Kind() && rv.Kind() <= reflect.Int64 {
		properties = append(properties, fmt.Sprintf("=%s", strconv.FormatInt(rv.Int(), 10)))
	} else if reflect.Uint <= rv.Kind() && rv.Kind() <= reflect.Uint64 {
		properties = append(properties, fmt.Sprintf("=%s", strconv.FormatUint(rv.Uint(), 10)))
	} else if reflect.Complex64 <= rv.Kind() && rv.Kind() <= reflect.Complex128 {
		properties = append(properties, fmt.Sprintf("=%s", strconv.FormatComplex(rv.Complex(), 'f', 2, 64)))
	} else if reflect.Float32 <= rv.Kind() && rv.Kind() <= reflect.Float64 {
		properties = append(properties, fmt.Sprintf("=%s", strconv.FormatFloat(rv.Float(), 'E', -1, 64)))
	} else if rv.Kind() == reflect.Chan {
		//ignore?
	} else if rv.Kind() == reflect.Uintptr {
		//ignore?
	} else if rv.Kind() == reflect.UnsafePointer {
		//ignore?
	} else if rv.Kind() == reflect.Func {
		//ignore?
	} else if rv.Kind() == reflect.String {
		properties = append(properties, rv.String())
	} else if rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		nextPtrValue := rv.Elem()
		ptrKind := nextPtrValue.Kind()
		if nextPtrValue.IsValid() {
			for {
				if ptrKind == reflect.Pointer {
					nextPtrValue = nextPtrValue.Elem()
					ptrKind = nextPtrValue.Kind()
				} else {
					break
				}
			}
		}
		return p.translate(nextPtrValue)
	} else if rv.Kind() == reflect.Map {
		mapKeys := rv.MapKeys()
		for _, key := range mapKeys {
			itemValue := rv.MapIndex(key)
			items, it, er := p.translate(itemValue)
			if er != nil {
				err = er
				return
			}
			for index, item := range items {
				if it {
					properties = append(properties, fmt.Sprintf("%s[%d].%s", key, index, item))
				} else {
					properties = append(properties, fmt.Sprintf("%s.%s", key, item))
				}
			}
		}
	} else if rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			itemValue := rv.Index(i)
			items, it, er := p.translate(itemValue)
			if er != nil {
				err = er
				return
			}
			for index, nextItem := range items {
				if strings.HasPrefix(nextItem, "=") || strings.HasPrefix(nextItem, "[") {
					if it {
						properties = append(properties, fmt.Sprintf("[%d].%s", index, nextItem))
					} else {
						properties = append(properties, fmt.Sprintf("[%d]%s", index, nextItem))
					}
				} else {
					properties = append(properties, fmt.Sprintf("[%d].%s", i, nextItem))

				}
			}
		}
		hasNext = true
	} else if rv.Kind() == reflect.Struct {
		filedNumber := rv.NumField()
	nextFiled:
		for i := 0; i < filedNumber; i++ {
			itemData := rv.Field(i)
			items, _, er := p.translate(itemData)
			if er != nil {
				err = er
				return
			}
			tag := rv.Type().Field(i).Tag
			for _, ignoreFlag := range p.IgnoreFlags {
				itemJsonTags := tag.Get(ignoreFlag.Flag)
				if itemJsonTags != "" && itemJsonTags == ignoreFlag.Value {
					continue nextFiled
				}
			}
			for _, item := range items {
				itemName := rv.Type().Field(i).Name
				if p.UseJsonFlag {
					itemJsonTag := tag.Get("json")
					if itemJsonTag != "" {
						itemJsonTagName, itemJsonTagOpt, _ := strings.Cut(itemJsonTag, ",")
						for _, c := range itemJsonTagName {
							switch {
							case strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", c):
								// Backslash and quote chars are reserved, but
								// otherwise any punctuation chars are allowed
								// in a tag name.
							case !unicode.IsLetter(c) && !unicode.IsDigit(c):

								err = errors.New("Invalid json tag")
								return
							}
						}
						if strings.Contains(itemJsonTagOpt, "omitempty") {
							if item == "" && !p.PrintEmpty {
								continue
							}
						}
						itemName = itemJsonTagName
					}
				}

				if strings.HasPrefix(item, "=") || strings.HasPrefix(item, "[") {
					properties = append(properties, fmt.Sprintf("%s%s", itemName, item))
				} else {
					properties = append(properties, fmt.Sprintf("%s.%s", itemName, item))
				}
			}
		}
	}
	return
}
