package main

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type Named interface {
	GetName() string
}

type TrollId uint32

type PxValue int32

type PxAccount float32

type Troll struct {
	Id        TrollId
	Name      string
	Account   PxAccount
	Shareable PxValue
}

type TrollSelection []*Troll

// see sort.Interface
func (s TrollSelection) Len() int { return len(s) }

// see sort.Interface
func (s TrollSelection) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (troll Troll) GetName() string {
	return troll.Name
}

type Group struct {
	Name   string
	Trolls []TrollId
}

func (group Group) GetName() string {
	return group.Name
}

type Transfer struct {
	Source       TrollId
	Destinations []TrollId
	Total        PxValue
	PerTroll     PxValue
	Message      string
}

func NewTransfer(source Troll, pxPerTroll PxValue, trolls []TrollId, message string) Transfer {
	total := pxPerTroll * PxValue(1+len(trolls))
	t := Transfer{Source: source.Id, Total: total, PerTroll: pxPerTroll, Message: message}

	t.Destinations = make([]TrollId, 0, len(trolls))
	for _, recipient := range trolls {
		t.Destinations = append(t.Destinations, recipient)
	}
	return t
}

type Proposal struct {
	Transfer
	Id string
}

func (p Proposal) GetName() string {
	return p.Id
}

func (p Proposal) ToString(bank *PxBank) string {
	recipientNames := make([]string, 0, len(p.Destinations))
	for _, destination := range p.Destinations {
		value := fmt.Sprintf("%v (%v)", bank.Trolls[destination].Name, destination)
		recipientNames = append(recipientNames, value)
	}
	return fmt.Sprintf("%v : <strong>%v</strong> shares <strong>%v Px</strong> with <strong>%v</strong> (%v other trolls, <strong>%v per troll</strong>; total: %v)",
		p.Id, bank.Trolls[p.Source].Name, p.Total, recipientNames, len(recipientNames), p.PerTroll, p.Total)
}

func id() string {
	buf := make([]byte, 16)
	io.ReadFull(rand.Reader, buf)
	return fmt.Sprintf("%x", buf)
}

type Proposals []Proposal

type Transaction struct {
	Transfer
	Date time.Time
}

func (transaction *Transaction) Persist() error {
	file, err := os.OpenFile("px_bank_transactions.storage", os.O_APPEND|os.O_CREATE, 0x666)

	if nil != err {
		fmt.Println("---- !!!  Could not persist transaction : ", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(*transaction)
	err = encoder.Encode(transaction.Transfer)
	if nil != err {
		fmt.Println("#### !!!  Could not persist transaction : ", err)
	}
	return err
}

type PxBank struct {
	Trolls    map[TrollId]Troll
	Groups    []Group
	Proposals []Proposal
}

func (bank *PxBank) Persist() error {
	file, err := os.OpenFile("px_bank.storage", os.O_WRONLY|os.O_CREATE, 0x666)

	if nil != err {
		fmt.Println("#### !!!  Could not persist bank : ", err)
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(*bank)
	if nil != err {
		fmt.Println("#### !!!  Could not persist bank : ", err)
	} else {
		fmt.Println(".... ... Bank persisted")
	}
	return err
}

func (bank *PxBank) Reload() error {
	file, err := os.Open("px_bank.storage")
	if nil != err {
		fmt.Println("Px Bank storage does not exist yet, creating default storage")

		bank.Trolls = make(map[TrollId]Troll)
		bank.Groups = make([]Group, 0, 100)
		bank.Proposals = make([]Proposal, 0, 10)
		return nil
	}

	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(bank)
	return err
}
