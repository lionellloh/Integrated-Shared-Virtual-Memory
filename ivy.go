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

type MessageType int

const (
	ReadRequest MessageType = iota
	WriteRequest
	ReadForward
	WriteForward
	SendPage
	ReadConfirmation
	WriteConfirmation
	Invalidate
	InvalidateConfirmation
)

func (m MessageType) String() string {
	return [...]string{"ReadRequest", "WriteRequest", "ReadForward", "WriteForward", "SendPage",
		"ReadConfirmation", "WriteConfirmation", "Invalidate", "InvalidateConfirmation"}[m]
}

type OwnershipDetails struct {
	copySet map[string]bool
	owner   *Process
}

type CentralManager struct {
	id           string
	globalRecord map[string]OwnershipDetails
	channel      chan Message
}

type Message struct {
	messageType  MessageType
	senderID     string
	receiverID   string
	targetPageID string
	page         Page
}

type Process struct {
	id          string
	localMemory map[string]AccessTypePage
	channel     chan Message
	ptrMap      map[string]*Process
	cm          *CentralManager
}

func (p *Process) registerCM(cm *CentralManager) {
	p.cm = cm
}

func (sender interface{}) sendMessage(msg Message, receiver interface{}) {
	receiver.channel <- msg
}

func (p *Process) makeReadRequest(pageID string) {
	m := Message{
		messageType:  ReadRequest,
		senderID:     p.id,
		receiverID:   p.cm.id,
		targetPageID: pageID,
	}
	p.cm.channel <- m

}

func (p *Process) makeWriteRequest(pageID string) {
	m := Message{
		messageType:  WriteRequest,
		senderID:     p.id,
		receiverID:   p.cm.id,
		targetPageID: pageID,
	}
	p.cm.channel <- m

}

func (p *Process) sendPage(page Page, receiver *Process) {
	m := Message{
		messageType: SendPage,
		senderID:    p.id,
		receiverID:  receiver.id,
		page:        page,
	}

	receiver.channel <- page
}

type AccessTypePage struct {
	accessType AccessType
	page       *Page
}

type Page struct {
	id      string
	content string
}

func main() {
	fmt.Println(ReadOnly)
	fmt.Println(ReadForward)
}
