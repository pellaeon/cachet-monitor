package main

type Checker interface {
	//New(Parameter, Expect interface{}) Checker
	Check() (bool, int64, string)
}
