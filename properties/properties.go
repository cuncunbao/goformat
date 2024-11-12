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
	UseJsonFlag     bool            //在将结构体序列化为properties的时候,字段的名称是否使用属性上的jsonTag
	IgnoreFlags     []IgnoreFlag    //对于某些敏感字段,如果需要
	PrintEmpty      bool            //默认情况下,如果属性值为nil,则属性不会输出到文件
	Conjunction     Conjunction     //key之间的连接词,默认是.,
	ArrayParticiple ArrayParticiple //默认是[index],也可以修改为__index__
}

type Conjunction string

const (
	Dot       Conjunction = "Dot"       //点区分上下级,默认值
	Underline Conjunction = "Underline" //下划线
)

type ArrayParticiple string

const (
	SquareBracket   ArrayParticiple = "SquareBracket"   //方括号,默认值
	DoubleUnderline ArrayParticiple = "DoubleUnderline" //双下划线
)

type IgnoreFlag struct {
	Flag  string
	Value string
}

func (p propertiesFormater) KeyConjunction() string {
	if p.Conjunction == Dot {
		return "."
	} else if p.Conjunction == Underline {
		return "_"
	}
	return "."
}

func (p propertiesFormater) LeftArrayParticiple() string {
	if p.ArrayParticiple == SquareBracket {
		return "["
	} else if p.ArrayParticiple == DoubleUnderline {
		return "__"
	}
	return "["
}
func (p propertiesFormater) RightArrayParticiple() string {
	if p.ArrayParticiple == SquareBracket {
		return "]"
	} else if p.ArrayParticiple == DoubleUnderline {
		return "__"
	}
	return "]"
}

func DefaultPropertiesFormater() *propertiesFormater {
	return &propertiesFormater{
		UseJsonFlag:     true,
		Conjunction:     Dot,
		ArrayParticiple: SquareBracket,
	}
}
func NewPropertiesFormater(useJsonFlag bool, ignoreFlags []IgnoreFlag, printEmpty bool, conjunction Conjunction, arrayParticiple ArrayParticiple) *propertiesFormater {
	return &propertiesFormater{
		IgnoreFlags: ignoreFlags, UseJsonFlag: useJsonFlag, PrintEmpty: printEmpty, Conjunction: conjunction, ArrayParticiple: arrayParticiple,
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
		properties = append(properties, fmt.Sprintf("=\"%s\"", rv.String()))
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
			return p.translate(nextPtrValue)
		}
		return
	} else if rv.Kind() == reflect.Map {
		mapKeys := rv.MapKeys()
		for _, key := range mapKeys {
			itemValue := rv.MapIndex(key)
			if !itemValue.IsValid() {
				continue
			}
			items, it, er := p.translate(itemValue)
			if er != nil {
				err = er
				return
			}
			itemKeyName := p.stripMapKey(key)
			for _, nextItem := range items {
				if strings.HasPrefix(nextItem, "=") || strings.HasPrefix(nextItem, p.LeftArrayParticiple()) {
					if it {
						properties = append(properties, fmt.Sprintf("%s%s", itemKeyName, nextItem))
					} else {
						properties = append(properties, fmt.Sprintf("%s%s%s", itemKeyName, p.KeyConjunction(), nextItem))
					}
				} else {
					properties = append(properties, fmt.Sprintf("%s%s%s", itemKeyName, p.KeyConjunction(), nextItem))
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
				if strings.HasPrefix(nextItem, "=") || strings.HasPrefix(nextItem, p.LeftArrayParticiple()) {
					if it {
						properties = append(properties, fmt.Sprintf("%s%d%s%s%s", p.LeftArrayParticiple(), index, p.RightArrayParticiple(), p.KeyConjunction(), nextItem))
					} else {
						properties = append(properties, fmt.Sprintf("%s%d%s%s", p.LeftArrayParticiple(), i, p.RightArrayParticiple(), nextItem))
					}
				} else {
					properties = append(properties, fmt.Sprintf("%s%d%s%s%s", p.LeftArrayParticiple(), i, p.RightArrayParticiple(), p.KeyConjunction(), nextItem))
				}
			}
		}
		hasNext = true
	} else if rv.Kind() == reflect.Struct {
		filedNumber := rv.NumField()
	nextFiled:
		for i := 0; i < filedNumber; i++ {
			itemData := rv.Field(i)
			if !itemData.IsValid() && !p.PrintEmpty {
				continue
			}
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
					itemJsonTag := tag.Get("property")
					if itemJsonTag == "" {
						itemJsonTag = tag.Get("json")
					}
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
						if itemJsonTagName == "-" {
							continue //说明是需要被忽略的值
						}
						itemName = itemJsonTagName
					}
				}

				if strings.HasPrefix(item, "=") || strings.HasPrefix(item, p.LeftArrayParticiple()) {
					properties = append(properties, fmt.Sprintf("%s%s", itemName, item))
				} else {
					properties = append(properties, fmt.Sprintf("%s%s%s", itemName, p.KeyConjunction(), item))
				}
			}
		}
	}
	return
}

func (p propertiesFormater) stripMapKey(key reflect.Value) string {
	var itemKeyName string
	switch key.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer, reflect.Array, reflect.Struct:
		itemKeyName = strconv.FormatUint(uint64(uintptr(key.UnsafePointer())), 16)
	case reflect.Bool:
		itemKeyName = strconv.FormatBool(key.Bool())
	case reflect.Float32:
		itemKeyName = strconv.FormatFloat(key.Float(), 'E', -1, 32)
	case reflect.Float64:
		itemKeyName = strconv.FormatFloat(key.Float(), 'E', -1, 64)
	case reflect.Complex64:
		itemKeyName = strconv.FormatComplex(key.Complex(), 'E', -1, 64)
	case reflect.Complex128:
		itemKeyName = strconv.FormatComplex(key.Complex(), 'E', -1, 128)
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		itemKeyName = strconv.FormatInt(key.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		itemKeyName = strconv.FormatUint(key.Uint(), 10)
	case reflect.String:
		itemKeyName = key.String()
	case reflect.Interface:
		implValue := key.Elem()
		if implValue.IsValid() {
			return p.stripMapKey(implValue)
		}
	default:
		return itemKeyName
	}
	return itemKeyName
}
