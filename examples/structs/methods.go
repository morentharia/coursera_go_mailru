package main

import (
	"fmt"

	"github.com/k0kubun/pp"
)

type Person struct {
	Id   int
	Name string
}

// не изменит оригинальной структуры, для который вызван
func (p Person) UpdateName(name string) {
	p.Name = name
}

// изменяет оригинальную структуру
func (p *Person) SetName(name string) {
	p.Name = name
}

type Account struct {
	Id   int
	Name string
	Person
}

func (p *Account) SetName(name string) {
	p.Name = name
}

type MySlice []int

func (sl *MySlice) Add(val int) {
	*sl = append(*sl, val)
}

func (sl *MySlice) Count() int {
	return len(*sl)
}

func main() {
	// pers := &Person{1, "Vasily"}
	// pers := new(Person)
	pers := Person{1, "Vasily"}
	pers.SetName("Vasily Romanov zzzzzjkms")
	// (&pers).SetName("Vasily Romanov")
	fmt.Printf("updated person: %#v\n", pers)
	// return

	var acc Account = Account{
		Id:   1,
		Name: "rvasily",
		Person: Person{
			Id:   2,
			Name: "Vasily Romanov",
		},
	}

	acc.SetName("Account Name")
	pp.Println(acc)

	acc.Person.SetName("Person Name")

	// fmt.Printf("%#v \n", acc)
	pp.Println(acc)

	sl := MySlice([]int{1, 2})
	sl.Add(5)
	fmt.Println(sl.Count(), sl)
	// pp.Println(sl)
}
