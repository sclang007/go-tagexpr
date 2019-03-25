# Golang使用标签表达式校验结构体字段的有效性
一、背景

在服务的API接口层面，我们常常需要验证参数的有效性。 Golang中，大部分参数校验场景实际上是先将数据Bind到结构体，然后校验其字段值。

一般地，校验结构体字段值有如下两种实现方式。

Case-By-Case 针对每个需校验的结构体字段分别写校验代码
优点：自由灵活，适应所有场景
缺点：重复且琐碎的码农工作，易使人厌烦
规则匹配，在结构体标签中设置预先支持的验证规则，如email、max:100等形式
优点：使用简单，不需要写琐碎的代码
缺点：强依赖有限的规则，缺乏灵活性，无法满足复杂场景，如多字段关联验证等
思考：有没有一种方式，即简单易用（少写代码），又能满足各种复杂的校验场景？

答案是：有！结构体标签表达式 go-tagexpr 的出现，为我们提供了兼得鱼和熊掌的第三种选择。

二、认识 go-tagexpr

go-tagexpr 允许Gopher们在 struct tag 写表达式代码，并通过高性能的解释器计算其结果。

安装

   go get -u github.com/bytedance/go-tagexpr
下面使用一个小示例，演示含有枚举、比较、字段关联的较复杂场景。

示例代码

   import (
	 "fmt"

	 tagexpr "github.com/bytedance/go-tagexpr"
   )

   func ExampleTagexpr() {
	vm := tagexpr.New("te")
	type Meteorology struct {
		Season      string `te:"$=='spring'||$=='summer'||$=='autumn'||$=='winter'"`
		Weather     string `te:"$!='snowing' || (Season)$=='winter'"`
		Temperature int    `te:"{range:$>=-10 && $<38}{alarm:sprintf('Uncomfortable temperature: %v',$)}"`
	}
	m := &Meteorology{
		Season:      "summer",
		Weather:     "snowing",
		Temperature: 40,
	}
	r := vm.MustRun(m)
	fmt.Println(r.Eval("Season"))
	fmt.Println(r.Eval("Weather"))
	fmt.Println(r.Eval("Temperature@range"))
	fmt.Println(r.Eval("Temperature@alarm"))
	// Output:
	// true
	// false
	// false
	// Uncomfortable temperature: 40
 }
代码诠释：

新建一个标签名称为 te 的解释器

vm := tagexpr.New("te")
定义一个结构体，添加标签表达式，并实例化一个 m 对象。其中 $ 表示当前字段值，(Season)$ 表示 Season 字段的值

 type Meteorology struct {
    Season      string `te:"$=='spring'||$=='summer'||$=='autumn'||$=='winter'"`
    Weather     string `te:"$!='snowing' || (Season)$=='winter'"`
    Temperature int    `te:"{range:$>=-10 && $<38}{alarm:sprintf('Uncomfortable temperature: %v',$)}"`
 }
 m := &Meteorology{
    Season:      "summer",
    Weather:     "snowing",
    Temperature: 40,
 }
将对象实例 m 放入解释器中运行，返回表达式对象 r

r := vm.MustRun(m)
计算 Season 字段匿名表达式（$=='spring'||$=='summer'||$=='autumn'||$=='winter'）的值。因字段值 summer 在穷举列表中，故表达式结果为“true”

r.Eval("Season")
计算 Weather 字段匿名表达式 $!='snowing' || (Season)$=='winter' 的值。因字段值为 snowing 且 Season 为 summer，故表达式结果为“false”

r.Eval("Weather")
计算 Temperature 字段的 range 表达式 $>=-10 && $<38 的值。因字段值为 40，超出给出的范围，所以结果为“false”

r.Eval("Temperature@range")
计算 Temperature 字段的 alarm 表达式 sprintf('Uncomfortable temperature: %v',$) 的值。这是一个调用内部函数的表达式，它打印并返回字符串，结果为“Uncomfortable temperature: 40”

r.Eval("Temperature@alarm")
获取更多关于 go-expr 结构体标签表达式的语法知识 -> 查看这里

二、使用Validator校验

Validator 是有 go-expr 包提供的一个采用结构体标签表达式的参数校验组件。

主要特性

它要求在每个待校验字段上添加结果为布尔值的匿名表达式
当表达式结果为false时，表示验证不通过，此时组件将返回与该字段相关的错误信息
它支持使用名称为msg且结果为字符串的表达式作为错误信息
允许用户按需求自由修改错误信息的模板
支持各种常见的运算符
支持访问数组，切片，字典成员
支持访问当前结构体中的任何字段
支持访问嵌套字段，非导出字段等
支持注册自定义的验证函数表达式
内置len，sprintf，regexp，email，phone等函数表达式
安装

go get -u github.com/bytedance/go-tagexpr
我们基于前面示例稍作修改，来演示如何使用validator校验结构体字段的有效性。

示例代码

import (
	"fmt"

	"github.com/bytedance/go-tagexpr/validator"
)

func ExampleValidator() {
	vd := validator.New("vd")
	type Meteorology struct {
		Season      string `vd:"$=='spring'||$=='summer'||$=='autumn'||$=='winter'"`
		Weather     string `vd:"$!='snowing' || (Season)$=='winter'"`
		Temperature int    `vd:"{@:$>=-10 && $<38}{msg:sprintf('Uncomfortable temperature: %v',$)}"`
		Contact     string `vd:"email($)"`
	}
	m := &Meteorology{
		Season:      "summer",
		Weather:     "rain",
		Temperature: 40,
		Contact:     "henrylee2cn@gmail.com",
	}
	err := vd.Validate(m)
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// Uncomfortable temperature: 40
}
代码诠释：

新建一个标签名称为 vd 的校验器

vd := validator.New("vd")
定义一个结构体，在标签上添加校验表达式，并使用 m 实例进行测试。

type Meteorology struct {
    Season      string `vd:"$=='spring'||$=='summer'||$=='autumn'||$=='winter'"`
    Weather     string `vd:"$!='snowing' || (Season)$=='winter'"`
    Temperature int    `vd:"{@:$>=-10 && $<38}{msg:sprintf('Uncomfortable temperature: %v',$)}"`
    Contact     string `vd:"email($)"`
}
m := &Meteorology{
    Season:      "summer",
    Weather:     "rain",
    Temperature: 40,
    Contact:     "henrylee2cn@gmail.com",
}
校验实例 m 的各字段值是否有效，如果无效，则返回error信息

err := vd.Validate(m)
注册自己的校验函数

可能你已注意到 email($) 这个表达式，它是默认注册的一个函数表达式，用于验证邮箱的有效性。其实我们也可以定义自己通用的函数表达式，以便较少标签中的代码量，增加代码复用性。

下面以 email 函数的实现为例，演示如何注册自己的校验函数：

var pattern = "^([A-Za-z0-9_\\-\\.\u4e00-\u9fa5])+\\@([A-Za-z0-9_\\-\\.])+\\.([A-Za-z]{2,8})$"

emailRegexp := regexp.MustCompile(pattern)

validator.RegValidateFunc("email", func(args ...interface{}) bool {
	if len(args) != 1 {
		return false
	}
	s, ok := args[0].(string)
	if !ok {
		return false
	}
	return emailRegexp.MatchString(s)
}, true)
其中，validator.RegValidateFunc 的定义如下：

func RegValidateFunc(funcName string, fn func(args ...interface{}) bool, force ...bool) error
RegValidateFunc的force可选参数，表示是否强制覆盖已经注册了的同名函数。

**结论：**validator的使用方法非常简单、灵活且具有良好的扩展性，能够轻松满足各种复杂的验证场景。
