package main

import "fmt"

type AccessType int

const (
	ReadOnly AccessType = iota
	WriteOnly
	ReadWrite

)

func (a AccessType) String() string {
	return [...]string{"ReadOnly", "WriteOnly", "ReadWrite"}[a]
}


type OwnershipDetails struct {
	copySet map[string]bool
	owner *Process

}

type CentralManager struct {

}


type Process struct {
	id string
	localMemory map[string] AccessTypePage

}

//func (p *Process) request

type AccessTypePage struct {
	accessType AccessType
	page *Page

}


type Page struct {
	id string
	content string
}

func main(){
	fmt.Println(ReadOnly)
}