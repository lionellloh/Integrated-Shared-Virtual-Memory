package main

import (
	"fmt"
	"strconv"
)


//sendMessage or sendCMMessage for both Process and CM are base level functions that requires the ptr to the Process to specify the receiver
//However, any other sending functions that invokes sendMessage will rely on parsing processID [string format]
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
	owner   string
}


type Node interface {
	sendMessage()
}

type CentralManager struct {
	id           string
	//map[pageOwnerID]
	//copySet {copyOwner1_ID, copyOwner2_ID}
	//trueOwner {trueOwner_ID}
	globalRecord map[string]OwnershipDetails
	channel      chan Message
	ptrMap      map[string]*Process
	pageIDCounter int
}

type Message struct {
	messageType  MessageType
	senderID     string
	receiverID   string
	targetPageID string //targetPageID is usually used by the CM when trying to invalidate to request readForward or writeForward
	page         Page
	accessType AccessType  //AccessType is for sendPage
	forwardDestID string
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

func (p *Process) sendMessage(msg Message, receiver *Process) {
	fmt.Printf("[Process %s] Sending Message of type <%s> to [Process %s] \n", p.id, msg.messageType, receiver.id)
	receiver.channel <- msg
}

func (p *Process) sendCMMessage(msg Message){
	fmt.Printf("[Process %s] Sending Message of type <%s> to CM \n", p.id, msg.messageType)
	p.cm.channel <- msg
}

func (cm *CentralManager) sendMessage(msg Message, p *Process){
	p.channel <- msg
}

func (p *Process) sendReadRequest(pageID string) {
	m := Message{
		messageType:  ReadRequest,
		senderID:     p.id,
		receiverID:   p.cm.id,
		targetPageID: pageID,
	}
	p.cm.channel <- m

}

func (p *Process) sendWriteRequest(pageID string) {
	m := Message{
		messageType:  WriteRequest,
		senderID:     p.id,
		receiverID:   p.cm.id,
		targetPageID: pageID,
	}
	p.cm.channel <- m

}

func (p *Process) sendPage(page Page, receiverID string, accessType AccessType) {
	msg := Message{
		messageType: SendPage,
		senderID:    p.id,
		receiverID:  receiverID,
		page:        page,
		accessType: accessType,
	}

	p.sendMessage(msg, p.ptrMap[receiverID])
}

//TODO: Look at refactoring sendPage using interface
func (cm *CentralManager) sendPage(page Page, receiverID string, accessType AccessType){
	msg := Message{
		messageType: SendPage,
		senderID: cm.id,
		receiverID: receiverID,
		page: page,
		accessType: accessType,
	}
	cm.sendMessage(msg, cm.ptrMap[receiverID])
}


/*
Process sending confirmation messages

*/
//Automatically assume that a read confirmation is sent to the CM
func (p *Process) sendReadConfirmation(pageID string) {
	msg := Message{
		messageType: ReadConfirmation,
		senderID:    p.id,
		receiverID:  p.cm.id,
		targetPageID: pageID,
	}

	p.sendCMMessage(msg)
}

func (p *Process) sendWriteConfirmation(pageID string) {
	msg := Message{
		messageType: WriteConfirmation,
		senderID:    p.id,
		receiverID:  p.cm.id,
		targetPageID: pageID,
	}

	p.sendCMMessage(msg)
}


func (p *Process) sendInvalidateConfirmation(pageID string) {
	msg := Message{
		messageType: InvalidateConfirmation,
		senderID:    p.id,
		receiverID:  p.cm.id,
		targetPageID: pageID,
	}

	p.sendCMMessage(msg)
}

//Receive handlers
func (p *Process) onReceivePage(msg Message) {
	p.localMemory[msg.page.id] = AccessTypePage{
		accessType: msg.accessType,
		page:       msg.page,
	}
}

func (p *Process) onReceiveInvalidation(msg Message) {
	delete(p.localMemory, msg.targetPageID)
	p.sendInvalidateConfirmation(msg.targetPageID)
}

func (p *Process) onReceiveReadForward(msg Message) {
	forwardDestID := msg.forwardDestID
	//TODO: Assert the correct access type?
	localPage := p.localMemory[msg.targetPageID].page
	p.sendPage(localPage, forwardDestID, ReadOnly)

}

func (p *Process) onReceiveWriteForward(msg Message) {
	forwardDestID := msg.forwardDestID
	//TODO: Assert the correct access type?
	localPage := p.localMemory[msg.targetPageID].page
	p.sendPage(localPage, forwardDestID, WriteOnly)

}


/*
CM Methods
*/


func (cm *CentralManager) sendReadForward(pageID string, receiverID string, forwardDestID string) {
	msg := Message{
		messageType:   ReadForward,
		senderID:      cm.id,
		receiverID:    receiverID,
		targetPageID:  pageID,
		page:          nil,
		accessType:    nil,
		forwardDestID: forwardDestID,
	}

	cm.sendMessage(msg, cm.ptrMap[receiverID])
}

func (cm *CentralManager) sendWriteForward(pageID string, receiverID string, forwardDestID string) {
	msg := Message{
		messageType:   WriteForward,
		senderID:      cm.id,
		receiverID:    receiverID,
		targetPageID:  pageID,
		page:          nil,
		accessType:    nil,
		forwardDestID: forwardDestID,
	}

	cm.sendMessage(msg, cm.ptrMap[receiverID])
}


func (cm *CentralManager) sendInvalidate(pageID string, receiverID string) {
	//pageID is the target pageID we want to invalidate
	msg := Message{
		messageType:   Invalidate,
		senderID:      cm.id,
		receiverID:    receiverID,
		targetPageID:  pageID,
		page:          nil,
		accessType:    nil,
		forwardDestID: "",
	}

	cm.sendMessage(msg, cm.ptrMap[receiverID])

}


func (cm *CentralManager) onReceiveReadRequest(msg Message) {
	var pageOwnerID string
	pageOwnershipDetails, exist := cm.globalRecord[msg.targetPageID]
	if exist {
		pageOwnerID = pageOwnershipDetails.owner
	} else {
		fmt.Printf("[CM] Page %s does not exist", msg.targetPageID)
		return
	}
	cm.sendReadForward(msg.targetPageID, pageOwnerID, msg.senderID)
}

func (cm *CentralManager) onReceiveWriteRequest(msg Message) {
	var pageOwnerID string
	pageOwnershipDetails, exist := cm.globalRecord[msg.targetPageID]
	if exist {
		pageOwnerID = pageOwnershipDetails.owner
	} else {
		fmt.Printf("[CM] Page %s does not exist", msg.targetPageID)

		//increment it so it's unique
		cm.pageIDCounter += 1

		//Give empty page to write?
		emptyPage := Page{
			id:      strconv.Itoa(cm.pageIDCounter),
			content: "",
		}
		cm.globalRecord[pageOwnerID] = OwnershipDetails{
			copySet: nil,
			owner:   msg.senderID,
		}

		//TODO: Is the access type ReadWrite?
		cm.sendPage(emptyPage, msg.senderID, ReadWrite)

	}
	cm.sendWriteForward(msg.targetPageID, pageOwnerID, msg.senderID)

}

func (cm *CentralManager) onReceiveReadConfirmation(msg Message) {
	//	TODO: not sure what to do here
	fmt.Printf("[CM] Received Read Confirmation \n")
}

func (cm *CentralManager) onInvalidateConfirmation(msg Message) {
	fmt.Printf("[CM] Received Read Confirmation \n")
}


type AccessTypePage struct {
	accessType AccessType
	page       Page
}

type Page struct {
	id      string
	content string
}

func main() {
	fmt.Println(ReadOnly)
	fmt.Println(ReadForward)
}
