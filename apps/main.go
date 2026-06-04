package main

import "fmt"

type Animal interface {
	Speak() string
}

type dog struct {
}

func (d *dog) Speak() string {
	return "bark"
}

type cat struct {
}

func (c *cat) Speak() string {
	return "meow"
}

type duck struct {
}

func (d *duck) Speak() string {
	return "quack"
}


type service struct {
	repo Animal
}

func NewService(repo Animal) *service {
	return &service{repo: repo}
}

func main() {
	d := &duck{}
	s := NewService(d)
	fmt.Println(s.repo.Speak())
}
