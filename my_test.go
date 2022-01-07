package main_test

import (
	"encoding/json"
	"testing"

	"github.com/dop251/goja"
	. "github.com/pepinns/go-hamcrest"
)

func Test_hello(t *testing.T) {
	vm := goja.New()
	v, err := vm.RunString("'hello '+'world'")
	Assert(t).That(err, IsNil())
	s := v.Export().(string)
	Assert(t).That(s, Equals("hello world"))
}

type Person struct {
	Name string
	Age  int
}
type Address struct {
	Desc string
}

type Recipient struct {
	Person  *Person
	Address *Address
}

func Test_pass_Go_struct_to_JS_func_and_modify_fields(t *testing.T) {
	js := goja.New()
	_, err := js.RunString(`
		function fillIn(person) {
			person.Name = "Dave";
			person.Age = 44;
		}
	`)
	Assert(t).That(err, IsNil())

	// fillIn, ok := goja.AssertFunction(vm.Get("fillIn"))
	// Assert(t).That(ok, IsTrue())

	// person := &Person{}
	// fillIn(goja.Undefined(), vm.ToValue(person))

	var fillIn func(person *Person)
	err = js.ExportTo(js.Get("fillIn"), &fillIn)
	Assert(t).That(err, IsNil())

	person := &Person{}
	fillIn(person)

	Assert(t).That(person.Name, Equals("Dave"))
	Assert(t).That(person.Age, Equals(44))
}

func StartName() string {
	return "Methusel"
}

func Test_set_Go_funcs_and_structs_into_context(t *testing.T) {
	js := goja.New()

	tweaker := func(p *Person) {
		p.Name = p.Name + "a"
		p.Age = p.Age + 1
	}
	person := &Person{}

	js.Set("person", &person)
	js.Set("StartName", StartName)
	js.Set("tweaker", tweaker)

	_, err := js.RunString(`
		person.Name = StartName()
		person.Age = 968
		tweaker(person)
	`)
	Assert(t).That(err, IsNil())

	Assert(t).That(person.Name, Equals("Methusela"))
	Assert(t).That(person.Age, Equals(969))

}

func Test_fabricate_struct_in_js(t *testing.T) {
	js := goja.New()

	v, err := js.RunString(`
		person = {}
		person.Name = "Lemmy"
		person.Age = 70
		person
	`)
	Assert(t).That(err, IsNil())

	var person Person
	err = js.ExportTo(v, &person)
	Assert(t).That(err, IsNil())

	Assert(t).That(person, Not(IsNil()))
	Assert(t).That(person.Name, Equals("Lemmy"))
	Assert(t).That(person.Age, Equals(70))
}

func Test_fabricate_compound_struct_in_js(t *testing.T) {
	js := goja.New()

	v, err := js.RunString(`
		person = {}
		person.Name = "Lemmy"
		person.Age = 70
		addr = {
			Desc: "The Rainbow"
		}
		// person
		recip = {
			Address: addr,
			Person: person,
		}
	`)
	Assert(t).That(err, IsNil())

	var recip Recipient
	err = js.ExportTo(v, &recip)
	Assert(t).That(err, IsNil())

	Assert(t).That(recip.Address, Not(IsNil()))
	Assert(t).That(recip.Address.Desc, Equals("The Rainbow"))
	Assert(t).That(recip.Person.Age, Equals(70))
	Assert(t).That(recip.Person, Not(IsNil()))
	Assert(t).That(recip.Person.Name, Equals("Lemmy"))
	Assert(t).That(recip.Person.Age, Equals(70))
}

func Test_return_object_structure_from_js(t *testing.T) {
	js := goja.New()
	v, err := js.RunString(`
		const obj = {
			number: 42,
			animals: [
				"zerba",
				"lion",
			],
			meta: {
				thinger: "hullo"
			}
		};
		obj
	`)
	Assert(t).That(err, IsNil())

	Assert(t).That(v, Not(IsNil()))
	m := v.ToObject(js)
	Assert(t).That(m.Get("meta").ToObject(js).Get("thinger").ToString(), Equals("hullo"))

	arr := m.Get("animals").ToObject(js)
	Assert(t).That(arr, Not(IsNil()))

	// TODO: better way of accessing Arrays?
	// arr2, ok := m.Get("animals").Export().(*goja.Array)
	// Assert(t).That(ok, IsTrue())
	// Assert(t).That(arr2, Not(IsNil()))

	Assert(t).That(m, Not(IsNil()))
	ks := m.Keys()
	Assert(t).That(ks, Not(IsNil()))
}

func Test_return_array_of_structs(t *testing.T) {
	js := goja.New()

	v, err := js.RunString(`
		x = [
			{
				Name: "Lemmy",
				Age: 70,
			},
			{
				Name: "Ozzy",
				Age: 73,
			},
		]
		x
	`)
	Assert(t).That(err, IsNil())

	var gods = make([]Person, 0)
	err = js.ExportTo(v, &gods)
	Assert(t).That(err, IsNil())
	Assert(t).That(gods[0].Name, Equals("Lemmy"))
	Assert(t).That(gods[1].Name, Equals("Ozzy"))
}

func Test_consecutive_execution_shared_vars_in_global_context(t *testing.T) {
	js := goja.New()

	_, err := js.RunString(`
		function gods() {
			return [
				{
					Name: "Lemmy",
					Age: 70,
				},
				{
					Name: "Ozzy",
					Age: 73,
				},
			]
		}
		null
	`)
	Assert(t).That(err, IsNil())
	v2, err := js.RunString(`
		gods()
	`)
	Assert(t).That(err, IsNil())

	var gods = make([]Person, 0)
	err = js.ExportTo(v2, &gods)
	Assert(t).That(err, IsNil())
	Assert(t).That(gods[0].Name, Equals("Lemmy"))
	Assert(t).That(gods[1].Name, Equals("Ozzy"))
}

type Module struct {
	Exports map[string]interface{}
}

func Test_share_js_func_across_runtimes(t *testing.T) {
	js1 := goja.New()
	// mod := &Module{}
	// js1.Set("module",mod)
	godsVal, err := js1.RunString(`
		function gods() {
			return [
				{
					Name: "Lemmy",
					Age: 70,
				},
				{
					Name: "Ozzy",
					Age: 73,
				},
			]
		}
		// module.exports.gods = gods
		gods
	`)
	Assert(t).That(err, IsNil())

	godsFunc, ok := goja.AssertFunction(godsVal)
	Assert(t).That(ok, IsTrue())

	js2 := goja.New()
	js2.Set("gods", godsFunc)
	v2, err := js2.RunString("gods()")
	Assert(t).That(err, IsNil())
	Assert(t).That(v2, Not(IsNil()))

}

func Test_func_to_define_func(t *testing.T) {
	js := goja.New()

	loadCode := func() {
		js.RunString(`
			function getAnswer() {
				return 42;
			}
		`)
	}
	js.Set("loadCode", loadCode)

	v, err := js.RunString(`
		loadCode()
		getAnswer()
	`)
	Assert(t).That(err, IsNil())
	Assert(t).That(v.ToString(), Equals("42"))

}

func Test_json(t *testing.T) {
	js := goja.New()

	gods := []Person{
		{
			Name: "Lemmy",
			Age:  70,
		},
		{
			Name: "Ozzy",
			Age:  73,
		},
	}
	jsonText, err := json.Marshal(gods)
	Assert(t).That(err, IsNil())

	js.Set("jsonText", string(jsonText))

	v, err := js.RunString(`
		const gods = JSON.parse(jsonText);
		gods[1].Name
	`)
	Assert(t).That(err, IsNil())
	Assert(t).That(v.ToString(), Equals("Ozzy"))

}
