package main

type Checker interface {
	//New(Parameter, Expect interface{}) Checker
	Check() (bool, uint, string)
	Test() string
}
