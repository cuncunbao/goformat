package properties

import (
	"fmt"
	"reflect"
	"testing"
)

type GirlFriend struct {
	name string
	age  int
}

type Student struct {
	Name        map[string][]string `json:"names"`
	GirlFriends map[string][]GirlFriend
	Age         int `json:"Aage" `
	Addr        []Addres
	height      float64

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

func main() {

}

func TestProperties(t *testing.T) {
	// 将JSON数据解码到map结构
	//var data map[string]interface{}
	//fileJsonData, _ := os.ReadFile("/Users/wrong/itit/icode/go/golang/go_test_002/src/test_json/test/test.json")
	//if err := json.Unmarshal([]byte(fileJsonData), &data); err != nil {
	//	panic(err)
	//}
	type Person struct {
		Name string
		Age  int
	}

	alice := &Person{"Alice", 25}
	bob := &Person{"Bob", 30}

	mapPointer := map[*Person]string{
		alice: "Alice's pointer",
		bob:   "Bob's pointer",
	}
	fmt.Println(mapPointer)

	var tm map[string]*Student = make(map[string]*Student)
	tm["1"] = nil

	email := "admin@github.com"
	domain := "http://github.com"
	domianPtr := &Domain{
		url:      domain,
		describe: "my personal website",
	}
	s1 := Student{
		Name:   map[string][]string{"nickName": []string{"小王", "小风", "王小风"}, "normalName": []string{"wangrf", "wangrongfeng", "rongfeng"}},
		Age:    31,
		male:   true,
		height: 170.35,
		GirlFriends: map[string][]GirlFriend{
			"primary": []GirlFriend{{"g1", 18}, {"g2", 19}}, "middle": []GirlFriend{{"g3", 20}, {"g4", 21}},
		},
		email:  &email,
		domain: &domianPtr,
		Task: Homework{class: "学中文", details: []string{"学拼音", "背古诗"},
			time: []homeworktime{{[]int{1, 3, 5}}, {[]int{2, 4, 6}}}},
		grade: map[string]Score{
			"english": {"2023", 85},
			"math":    {"2024", 99},
		},
		Addr: []Addres{{"beijing", "100100"}, {"lijiang", "674800"}},
	}

	//file, err := properties.LoadFile("/Users/wrong/itit/icode/go/golang/go_test_002/src/test_json/test/test.json", properties.UTF8)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(file)
	// 创建properties格式的字符串

	drv := reflect.ValueOf(s1)
	//dname := drv.Type().Name()

	formater := NewPropertiesFormater(true, []IgnoreFlag{{"output", "hide"}}, false, Underline, DoubleUnderline)

	items, _ := formater.Translate(drv)

	for _, item := range items {
		fmt.Printf("%s\n", item)
	}

	formater1 := NewPropertiesFormater(true, []IgnoreFlag{{"output", "hide"}}, false, Dot, SquareBracket)

	items1, _ := formater1.Translate(drv)

	for _, item := range items1 {
		fmt.Printf("%s\n", item)
	}
}
