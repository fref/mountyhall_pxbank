package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

func DisplayTransfersMenu(bank *PxBank) {
	PrintHeader(bank, "TRANSFERS")
	fmt.Println("1. Register Kill")
	fmt.Println("2. Register Transfer")
	fmt.Println("3. Register Adjustment within group")
	fmt.Println("")
	fmt.Println("0. Quit")
	fmt.Print("Choice: ")
}

func HandleTransferAction(bank *PxBank, choice string) {
	switch choice {
	case "1":
		RegisterKill(bank)
		return
	case "2":
		RegisterShare(bank)
		return
	case "3":
		RegisterAdjustment(bank)
		return
	}
}

func RegisterKill(bank *PxBank) {
	DisplayGroups(bank)
	selectedGroup, _ := ChooseGroup(bank)

	fmt.Print("Troll name : ")
	trollName := readLine()
	troll := FindTrollInList(bank, selectedGroup.Trolls, trollName)

	fmt.Print("Kill info : ")
	details := readLine()
	var px PxValue
	fmt.Print("Px : ")
	fmt.Scanln(&px)

	fmt.Printf("About to record: %v / %v / %v / %v px earned. Confirm (y)? ", troll.Name, selectedGroup.Name, details, px)
	confirm := readLine()
	if "y" == confirm {
		troll.Shareable += px
		bank.Trolls[troll.Id] = troll
		t := Transaction{Date: time.Now(), Transfer: Transfer{Total: px, PerTroll: px, Message: "Kill: " + details, Destinations: []TrollId{troll.Id}}}
		t.Persist()

		perTroll := PxAccount(px) / PxAccount(len(selectedGroup.Trolls))
		for _, groupMember := range selectedGroup.Trolls {
			troll = bank.Trolls[groupMember]
			troll.Account += perTroll
			bank.Trolls[groupMember] = troll
		}
		bank.Persist()
	}
}

func RegisterShare(bank *PxBank) {
	fmt.Print("Troll name : ")
	trollName := readLine()
	source := FindTroll(bank, trollName)

	var px PxValue
	fmt.Print("Px per Troll: ")
	fmt.Scanln(&px)

	trolls := PickTrollList(bank)
	fmt.Printf("About to record: %v / share with %v other troll(s) / %v px per troll. Confirm (y)? ", source.Name, len(trolls), px)
	confirm := readLine()
	if "y" == confirm {
		destinations := make([]TrollId, 0, len(trolls))
		for _, troll := range trolls {
			destinations = append(destinations, troll.Id)
		}
		transfer := NewTransfer(source, px, destinations, "Px Share")
		t := Transaction{Date: time.Now(), Transfer: transfer}
		t.AdjustAccounts(bank)
		t.Persist()
		bank.Persist()
	}
}

func RegisterAdjustment(bank *PxBank) {
	DisplayGroups(bank)
	selectedGroup, _ := ChooseGroup(bank)

	fmt.Print("Troll name : ")
	trollName := readLine()
	troll := FindTrollInList(bank, selectedGroup.Trolls, trollName)

	fmt.Print("Adjustment info : ")
	details := readLine()
	var px PxValue
	fmt.Print("Px : ")
	fmt.Scanln(&px)

	fmt.Printf("About to record: %v / %v / %v / %v px will go from group to troll. Confirm (y)? ", troll.Name, selectedGroup.Name, details, px)
	confirm := readLine()
	if "y" == confirm {
		troll.Account += PxAccount(px)
		destinations := make([]TrollId, 0, len(selectedGroup.Trolls))
		for _, trollInGroup := range selectedGroup.Trolls {
			if trollInGroup == troll.Id {
				continue
			}
			destinations = append(destinations, trollInGroup)
		}

		perTroll := -PxAccount(px) / PxAccount(len(destinations))
		bank.Trolls[troll.Id] = troll
		t := Transaction{Date: time.Now(), Transfer: Transfer{Total: -px, Message: "Group Adjustment: " + details, Destinations: destinations}}
		t.Persist()

		for _, groupMember := range destinations {
			trollInGroup := bank.Trolls[groupMember]
			trollInGroup.Account += perTroll
			bank.Trolls[groupMember] = trollInGroup
		}
		bank.Persist()
	}
}

func (transaction Transaction) AdjustAccounts(bank *PxBank) {
	// Adjust source account
	source := bank.Trolls[transaction.Source]
	source.Shareable -= transaction.Total
	source.Account -= PxAccount(transaction.PerTroll)
	bank.Trolls[source.Id] = source

	// Adjust recipient accounts
	for _, recipientId := range transaction.Destinations {
		recipient := bank.Trolls[recipientId]
		recipient.Account -= PxAccount(transaction.PerTroll)
		bank.Trolls[recipient.Id] = recipient
	}
}

func ListTransactions(bank *PxBank) {
	file, err := os.Open("px_bank_transactions.storage")
	if nil != err {
		fmt.Println("Px Bank Transaction storage does not exist yet")
		return
	}
	defer file.Close()

	var t Transaction
	decoder := json.NewDecoder(file)
	for err = decoder.Decode(&t); nil == err; err = decoder.Decode(&t) {
		fmt.Println(t)
	}
	if io.EOF != err {
		fmt.Println(err)
	}
}
