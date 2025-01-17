// Copyright 2019 Bytedance Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tagexpr

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func BenchmarkTagExpr(b *testing.B) {
	b.StopTimer()
	type T struct {
		a int `bench:"$%3"`
	}
	vm := New("bench")
	err := vm.WarmUp(new(T))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.StartTimer()
	var t = &T{10}
	for i := 0; i < b.N; i++ {
		tagExpr, err := vm.Run(t)
		if err != nil {
			b.FailNow()
		}
		if tagExpr.EvalFloat("a") != 1 {
			b.FailNow()
		}
	}
}

func BenchmarkReflect(b *testing.B) {
	b.StopTimer()
	type T struct {
		a int `remainder:"3"`
	}
	b.ReportAllocs()
	b.StartTimer()
	var t = &T{1}
	for i := 0; i < b.N; i++ {
		v := reflect.ValueOf(t).Elem()
		ft, ok := v.Type().FieldByName("a")
		if !ok {
			b.FailNow()
		}
		x, err := strconv.ParseInt(ft.Tag.Get("remainder"), 10, 64)
		if err != nil {
			b.FailNow()
		}
		fv := v.FieldByName("a")
		if fv.Int()%x != 1 {
			b.FailNow()
		}
	}
}

func Test(t *testing.T) {
	g := &struct {
		_ int
		h string `tagexpr:"$"`
		s []string
		m map[string][]string
	}{
		h: "haha",
		s: []string{"1"},
		m: map[string][]string{"0": {"2"}},
	}
	d := "ddd"
	e := new(int)
	*e = 3
	type iface interface{}
	var cases = []struct {
		tagName   string
		structure interface{}
		tests     map[string]interface{}
	}{
		{
			tagName: "tagexpr",
			structure: &struct {
				A     int              `tagexpr:"$>0&&$<10&&!''&&!!!0&&!nil&&$"`
				A2    int              `tagexpr:"{@:$>0&&$<10}"`
				b     string           `tagexpr:"{is:$=='test'}{msg:sprintf('expect: test, but got: %s',$)}"`
				c     float32          `tagexpr:"(A)$+$"`
				d     *string          `tagexpr:"$"`
				e     **int            `tagexpr:"$"`
				f     *[3]int          `tagexpr:"{x:len($)}"`
				g     string           `tagexpr:"{x:!regexp('xxx',$)}{y:regexp('g\\d{3}$')}"`
				h     []string         `tagexpr:"{x:$[1]}{y:$[10]}"`
				i     map[string]int   `tagexpr:"{x:$['a']}{y:$[0]} {z:$==nil}"`
				i2    *map[string]int  `tagexpr:"{x:$['a']}{y:$[0]} {z:$}"`
				j, j2 iface            `tagexpr:"{@:$==1} {y:$}"`
				k     *iface           `tagexpr:"$==nil"`
				m     *struct{ i int } `tagexpr:"{@:$}{x:$['a']['x']}"`
			}{
				A:  5.0,
				A2: 5.0,
				b:  "x",
				c:  1,
				d:  &d,
				e:  &e,
				f:  new([3]int),
				g:  "g123",
				h:  []string{"", "hehe"},
				i:  map[string]int{"a": 7},
				j2: iface(1),
				m:  &struct{ i int }{1},
			},
			tests: map[string]interface{}{
				"A":     true,
				"A2":    true,
				"b@is":  false,
				"b@msg": "expect: test, but got: x",
				"c":     6.0,
				"d":     d,
				"e":     float64(*e),
				"f@x":   float64(3),
				"g@x":   true,
				"g@y":   true,
				"h@x":   "hehe",
				"h@y":   nil,
				"i@x":   7.0,
				"i@y":   nil,
				"i@z":   false,
				"i2@x":  nil,
				"i2@y":  nil,
				"i2@z":  nil,
				"j":     false,
				"j@y":   nil,
				"j2":    true,
				"j2@y":  1.0,
				"k":     true,
				"m":     &struct{ i int }{1},
				"m@x":   nil,
			},
		},
		{
			tagName: "tagexpr",
			structure: &struct {
				A int    `tagexpr:"$>0&&$<10"`
				b string `tagexpr:"{is:$=='test'}{msg:sprintf('expect: test, but got: %s',$)}"`
				c struct {
					_ int
					d bool `tagexpr:"$"`
				}
				e *struct {
					_ int
					f bool `tagexpr:"$"`
				}
				g **struct {
					_ int
					h string `tagexpr:"$"`
					s []string
					m map[string][]string
				} `tagexpr:"$['h']"`
				i string  `tagexpr:"(g.s)$[0]+(g.m)$['0'][0]==$"`
				j bool    `tagexpr:"!$"`
				k int     `tagexpr:"!$"`
				m *int    `tagexpr:"$==nil"`
				n *bool   `tagexpr:"$==nil"`
				p *string `tagexpr:"$"`
			}{
				A: 5,
				b: "x",
				c: struct {
					_ int
					d bool `tagexpr:"$"`
				}{d: true},
				e: &struct {
					_ int
					f bool `tagexpr:"$"`
				}{f: true},
				g: &g,
				i: "12",
			},
			tests: map[string]interface{}{
				"A":     true,
				"b@is":  false,
				"b@msg": "expect: test, but got: x",
				"c.d":   true,
				"e.f":   true,
				"g":     "haha",
				"g.h":   "haha",
				"i":     true,
				"j":     true,
				"k":     true,
				"m":     true,
				"n":     true,
				"p":     nil,
			},
		},
		{
			tagName: "p",
			structure: &struct {
				q *struct {
					x int
				} `p:"(q.x)$"`
			}{},
			tests: map[string]interface{}{
				"q": nil,
			},
		},
	}
	for i, c := range cases {
		vm := New(c.tagName)
		// vm.WarmUp(c.structure)
		tagExpr, err := vm.Run(c.structure)
		if err != nil {
			t.Fatal(err)
		}
		for selector, value := range c.tests {
			val := tagExpr.Eval(selector)
			if !reflect.DeepEqual(val, value) {
				t.Fatalf("Eval Serial: %d, selector: %q, got: %v, expect: %v", i, selector, val, value)
			}
		}
		tagExpr.Range(func(selector string, eval func() interface{}) bool {
			t.Logf("Range selector: %s", selector)
			value := c.tests[selector]
			val := eval()
			if !reflect.DeepEqual(val, value) {
				t.Fatalf("Range NO: %d, selector: %q, got: %v, expect: %v", i, selector, val, value)
			}
			return true
		})
	}
}

func TestField(t *testing.T) {
	g := &struct {
		_ int
		h string
		s []string
		m map[string][]string
	}{
		h: "haha",
		s: []string{"1"},
		m: map[string][]string{"0": {"2"}},
	}
	structure := &struct {
		A int
		b string
		c struct {
			_ int
			d *bool
		}
		e *struct {
			_ int
			f bool
		}
		g **struct {
			_ int
			h string
			s []string
			m map[string][]string
		}
		i string
		j bool
		k int
		m *int
		n *bool
		p *string
	}{
		A: 5,
		b: "x",
		e: &struct {
			_ int
			f bool
		}{f: true},
		g: &g,
		i: "12",
	}
	vm := New("")
	e, err := vm.Run(structure)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		fieldSelector string
		value         interface{}
	}{
		{"A", structure.A},
		{"b", structure.b},
		{"c", structure.c},
		{"c.d", structure.c.d},
		{"e", structure.e},
		{"e.f", structure.e.f},
		{"g", structure.g},
		{"g.h", (*structure.g).h},
		{"g.s", (*structure.g).s},
		{"g.m", (*structure.g).m},
		{"i", structure.i},
		{"j", structure.j},
		{"k", structure.k},
		{"m", structure.m},
		{"n", structure.n},
		{"p", structure.p},
	}
	for _, c := range cases {
		val := e.Field(c.fieldSelector)
		if !reflect.DeepEqual(val, c.value) {
			t.Fatalf("%s: got: %v(%[2]T), expect: %v(%[3]T)", c.fieldSelector, val, c.value)
		}
	}
	var wall uint64 = 1024
	unix := time.Unix(1549186325, int64(wall))
	e, err = vm.Run(&unix)
	if err != nil {
		t.Fatal(err)
	}
	val := e.Field("wall")
	if !reflect.DeepEqual(val, wall) {
		t.Fatalf("Time.wall: got: %v(%[1]T), expect: %v(%[2]T)", val, wall)
	}
}

func TestOperator(t *testing.T) {

	type Tmp1 struct {
		A string `tagexpr:$=="1"||$=="2"||$="3"`
		B []int  `tagexpr:len($)>=10&&$[0]<10`
		C interface{}
	}

	type Tmp2 struct {
		A *Tmp1
		B interface{}
	}

	type Target struct {
		A int             `tagexpr:"-$+$<=10"`
		B int             `tagexpr:"+$-$<=10"`
		C int             `tagexpr:"-$+(M)$*(N)$/$%(D.B)$[2]+$==1"`
		D *Tmp1           `tagexpr:"(D.A)$!=nil"`
		E string          `tagexpr:"((D.A)$=='1'&&len($)>1)||((D.A)$=='2'&&len($)>2)||((D.A)$=='3'&&len($)>3)"`
		F map[string]int  `tagexpr:"{x:len($)}{y:$['a']>10&&$['b']>1}"`
		G *map[string]int `tagexpr:"{x:$['a']+(F)$['a']>20}"`
		H []string        `tagexpr:"len($)>=1&&len($)<10&&$[0]=='123'&&$[1]!='456'"`
		I interface{}     `tagexpr:"$!=nil"`
		K *string         `tagexpr:"len((D.A)$)+len($)<10&&len((D.A)$+$)<10"`
		L **string        `tagexpr:"false"`
		M float64         `tagexpr:"$/2>10&&$%2==0"`
		N *float64        `tagexpr:"($+$*$-$/$+1)/$==$+1"`
		O *[3]float64     `tagexpr:"$[0]>10&&$[0]<20||$[0]>20&&$[0]<30"`
		P *Tmp2           `tagexpr:"{x:$!=nil}{y:len((P.A.A)$)<=1&&(P.A.B)$[0]==1}{z:$['A']['C']==nil}{w:$['A']['B'][0]==1}{r:$[0][1][2]==3}{s1:$[2]==nil}{s2:$[0][3]==nil}{s3:(ZZ)$}{s4:(P.B)$!=nil}"`
		Q *Tmp2           `tagexpr:"{s1:$['A']['B']!=nil}{s2:(Q.A)$['B']!=nil}{s3:$['A']['C']==nil}{s4:(Q.A)$['C']==nil}{s5:(Q.A)$['B'][0]==1}{s6:$['X']['Z']==nil}"`
	}

	k := "123456"
	n := float64(-12.5)
	o := [3]float64{15, 9, 9}
	var cases = []struct {
		tagName   string
		structure interface{}
		tests     map[string]interface{}
	}{
		{
			tagName: "tagexpr",
			structure: &Target{
				A: 5,
				B: 10,
				C: -10,
				D: &Tmp1{A: "3", B: []int{1, 2, 3}},
				E: "1234",
				F: map[string]int{"a": 11, "b": 9},
				G: &map[string]int{"a": 11},
				H: []string{"123", "45"},
				I: struct{}{},
				K: &k,
				L: nil,
				M: float64(30),
				N: &n,
				O: &o,
				P: &Tmp2{A: &Tmp1{A: "3", B: []int{1, 2, 3}}, B: struct{}{}},
				Q: &Tmp2{A: &Tmp1{A: "3", B: []int{1, 2, 3}}, B: struct{}{}},
			},
			tests: map[string]interface{}{
				"A":   true,
				"B":   true,
				"C":   true,
				"D":   true,
				"E":   true,
				"F@x": float64(2),
				"F@y": true,
				"G@x": true,
				"H":   true,
				"I":   true,
				"K":   true,
				"L":   false,
				"M":   true,
				"N":   true,
				"O":   true,

				"P@x":  true,
				"P@y":  true,
				"P@z":  true,
				"P@w":  true,
				"P@r":  true,
				"P@s1": true,
				"P@s2": true,
				"P@s3": nil,
				"P@s4": true,

				"Q@s1": true,
				"Q@s2": true,
				"Q@s3": true,
				"Q@s4": true,
				"Q@s5": true,
				"Q@s6": true,
			},
		},
	}

	for i, c := range cases {
		vm := New(c.tagName)
		// vm.WarmUp(c.structure)
		tagExpr, err := vm.Run(c.structure)
		if err != nil {
			t.Fatal(err)
		}
		for selector, value := range c.tests {
			val := tagExpr.Eval(selector)
			if !reflect.DeepEqual(val, value) {
				t.Fatalf("Eval NO: %d, selector: %q, got: %v, expect: %v", i, selector, val, value)
			}
		}
		tagExpr.Range(func(selector string, eval func() interface{}) bool {
			t.Logf("Range selector: %s", selector)
			value := c.tests[selector]
			val := eval()
			if !reflect.DeepEqual(val, value) {
				t.Fatalf("Range NO: %d, selector: %q, got: %v, expect: %v", i, selector, val, value)
			}
			return true
		})
	}

}
