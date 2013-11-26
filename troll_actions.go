package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func DisplayTrollMenu(bank *PxBank) {
	PrintHeader(bank, "TROLLS")
	fmt.Println("1. List Trolls")
	fmt.Println("2. Add Trolls")
	fmt.Println("3. Rename Troll")
	fmt.Println()
	fmt.Println("0. Quit")
	fmt.Print("Choice: ")
}

func HandleTrollAction(bank *PxBank, choice string) {
	switch choice {
	case "1":
		ListTrolls(bank)
		return
	case "2":
		AddTrolls(bank)
		return
	case "3":
		RenameTroll(bank)
		return
	}
}

func ListTrolls(bank *PxBank) {
	ids := make([]int,0, len(bank.Trolls))
	for trollId, _ := range bank.Trolls {
		ids = append(ids, int(trollId))
	}
	sort.Ints(ids)
	var shareable PxValue
	var totalForBank PxAccount
	for _, trollId := range ids {
		troll := bank.Trolls[TrollId(trollId)]
		fmt.Println(troll)
		shareable += troll.Shareable
		totalForBank += troll.Account
	}
	fmt.Printf("\nTotal: %v\tShareable: %v\n", totalForBank, shareable)
}

func AddTrolls(bank *PxBank) {
	fmt.Println("To add a troll, type his id then his name")
	var id TrollId
	choice := readLine()
	_, err := fmt.Sscan(choice, &id)
	if nil != err {
		fmt.Println("Unable to parse id : ", err)
		return
	}
	segments := strings.Split(choice, strconv.Itoa(int(id)))
	var name string
	if 1 == len(segments) {
		name = segments[0]
	} else {
		name = segments[1]
	}
	name = strings.TrimSpace(name)

	if _, ok := bank.Trolls[id]; ok {
		troll := bank.Trolls[id]
		troll.Name = name
		bank.Trolls[id] = troll
	} else {
		troll := Troll{Id: id, Name: name}
		bank.Trolls[id] = troll
	}
	err = bank.Persist()
	if nil != err {
		fmt.Println("Could not persist bank : ", err)
	}
}

func RenameTroll(bank *PxBank) {
	fmt.Print("Troll name : ")
	trollName := readLine()
	toRename := FindTroll(bank, trollName)

	fmt.Print("New name : ")
	newName := readLine()

	fmt.Printf("About to rename: %v into %v . Confirm (y)? ", toRename.Name, newName)
	confirm := readLine()
	if "y" == confirm {
		toRename.Name = newName
		bank.Trolls[toRename.Id] = toRename
		err := bank.Persist()
		if nil != err {
			fmt.Println("Could not persist bank : ", err)
		}
	}
}

func (troll Troll) String() string {
	return fmt.Sprintf("[%v %v] : \t\tPX: %v  \t| ToShare : %v", troll.Id, troll.Name, troll.Account, troll.Shareable)
}

func PickTrollList(bank *PxBank) (trolls []Troll) {
	fmt.Print("\nTroll List (comma separated)? ")
	choice := readLine()
	if "" == choice {
		return
	}
	choices := strings.Split(choice, ",")
	for _, choice = range choices {
		choice = strings.TrimSpace(choice)
		trolls = append(trolls, FindTroll(bank, choice))
	}
	fmt.Print("Selection: ")
	for _, troll := range trolls {
		fmt.Print(troll.GetName(), ", ")
	}
	fmt.Println()
	return
}

func FindTroll(bank *PxBank, choice string) (selected Troll) {
	selected = findTroll(bank.Trolls, choice)
	return
}

func findTroll(Trolls map[TrollId]Troll, choice string) (selected Troll) {
	matcher := newMatcher(choice)
	for _, possibleMatch := range Trolls {
		matcher.check(possibleMatch)
	}
	selected = matcher.bestMatch.(Troll)
	return
}

func FindTrollInList(bank *PxBank, TrollIds []TrollId, choice string) (selected Troll) {
	matcher := newMatcher(choice)
	for _, id := range TrollIds {
		if troll, ok := bank.Trolls[id]; ok {
			matcher.check(troll)
		}
	}
	selected = matcher.bestMatch.(Troll)
	return
}
