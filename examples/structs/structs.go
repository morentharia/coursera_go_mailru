package main

import (
	"github.com/k0kubun/pp"
)

type Person struct {
	Id      int
	Name    string
	Address string
}

type Account struct {
	Id      int
	Name    string
	Cleaner func(string) string
	Owner   Person
	Person
}

func main() {
	// полное объявление структуры
	var acc Account = Account{
		Id:   1,
		Name: "rvasily",
		Person: Person{
			Name:    "Василий",
			Address: "Москва",
		},
	}
	acc.Owner = Person{
		2,
		"Romanov Vasily",
		"Moscow",
	}

	// fmt.Printf("%#v\n", acc)

	pp.Println(acc)

	// fmt.Println(acc.Name)
	// fmt.Println(acc.Person.Name)
	pp.Println(acc.Name)
}
