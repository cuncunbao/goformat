package properties

import (
	"fmt"
	"reflect"
	"testing"
)

type Student struct {
	Name   string `json:"XXName"`
	Age    int    `json:"Aage" `
	Addr   []Addres
	height float64
	male   bool
	email  *string
	domain **Domain
	Task   Task
	grade  map[string]Score
}
type Task interface {
	Work() string
}

type Homework struct {
	class   string
	details []string
	time    []homeworktime
}
type homeworktime struct {
	times []int
}

func (Homework) Work() string {
	return "work hard"
}

type Score struct {
	year  string
	point int
}
type Addres struct {
	name string
	tel  string
}
type Domain struct {
	url      string
	describe string
}

func TestProperties(t *testing.T) {
	// 将JSON数据解码到map结构
	//var data map[string]interface{}
	//fileJsonData, _ := os.ReadFile("/Users/wrong/itit/icode/go/golang/go_test_002/src/test_json/test/test.json")
	//if err := json.Unmarshal([]byte(fileJsonData), &data); err != nil {
	//	panic(err)
	//}

	email := "admin@github.com"
	domain := "http://github.com"
	domianPtr := &Domain{
		url:      domain,
		describe: "my personal website",
	}
	s1 := Student{Name: "张三", Age: 31, male: true, height: 170.35,
		email: &email, domain: &domianPtr,
		Task: Homework{class: "学中文", details: []string{"学拼音", "背古诗"},
			time: []homeworktime{{[]int{1, 3, 5}}, {[]int{2, 4, 6}}}},
		grade: map[string]Score{
			"english": {"2023", 85},
			"math":    {"2024", 99},
		},
		Addr: []Addres{{"beijing", "100100"}, {"lijiang", "674800"}}}

	//file, err := properties.LoadFile("/Users/wrong/itit/icode/go/golang/go_test_002/src/test_json/test/test.json", properties.UTF8)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(file)
	// 创建properties格式的字符串

	drv := reflect.ValueOf(s1)
	//dname := drv.Type().Name()

	formater := NewPropertiesFormater(true, []IgnoreFlag{{"output", "hide"}}, true)

	items, _ := formater.Translate(drv)

	for _, item := range items {
		fmt.Printf("%s\n", item)
	}
}
