package main

import "fmt"

func main() {
	a := 2
	b := &a
	*b = 3  // a = 3
	c := &a // новый указатель на переменную a

	fmt.Printf("a = %+v\n", a)
	fmt.Printf("b = %+v\n", b)
	fmt.Printf("c = %+v\n", c)
	// получение указателя на переменнут типа int
	// инициализировано значением по-умолчанию
	d := new(int)

	fmt.Printf("d = %+v\n", d)
	fmt.Printf("*d = %+v\n", *d)

	*d = 12
	fmt.Printf("*d = %+v\n", *d)

	*c = *d // c = 12 -> a = 12

	fmt.Printf("a = %+v\n", a)
	fmt.Printf("b = %+v\n", b)
	fmt.Printf("c = %+v\n", c)
	fmt.Printf("d = %+v\n", d)
	fmt.Printf("*d = %+v\n", *d)

	*d = 13 // c и a не изменились

	c = d   // теперь с указывает туда же, куда d
	*c = 14 // с = 14 -> d = 14, a = 12
}
