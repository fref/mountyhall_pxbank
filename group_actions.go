package main

import "fmt"

func DisplayGroupsMenu(bank *PxBank) {
	PrintHeader(bank, "GROUPS")
	fmt.Println("1. List Groups")
	fmt.Println("2. View Group")
	fmt.Println("3. Create Group")
	fmt.Println("4. Add Trolls to Group")
	fmt.Println("5. Remove Trolls from Group")
	fmt.Println("6. Rename Group")
	fmt.Println("7. Delete Group")
	fmt.Println("")
	fmt.Println("0. Quit")
	fmt.Print("Choice: ")
}

func HandleGroupAction(bank *PxBank, choice string) {
	switch choice {
	case "1":
		DisplayGroups(bank)
		return
	case "2":
		DetailGroup(bank)
		return
	case "3":
		CreateGroup(bank)
		return
	case "4":
		AddTrollsToGroup(bank)
		return
	case "5":
		RemoveTrollsFromGroup(bank)
		return
	case "6":
		RenameGroup(bank)
		return
	case "7":
		DeleteGroup(bank)
		return
	}
	fmt.Println("Unknown choice")
}

func AddTrollsToGroup(bank *PxBank) {
	DisplayGroups(bank)
	selected, selectedIndex := ChooseGroup(bank)
	selected.PrintMembers(bank.Trolls)

	trolls := PickTrollList(bank)
	for _, troll := range trolls {
		selected.Trolls = append(selected.Trolls, troll.Id)
	}

	bank.Groups[selectedIndex] = selected
	bank.Persist()
	selected.PrintMembers(bank.Trolls)
}

func DetailGroup(bank *PxBank) {
	DisplayGroups(bank)
	selected, _ := ChooseGroup(bank)
	selected.PrintMembers(bank.Trolls)
}

func RemoveTrollsFromGroup(bank *PxBank) {
	DisplayGroups(bank)
	selected, selectedIndex := ChooseGroup(bank)
	selected.PrintMembers(bank.Trolls)
	choices := PickTrollList(bank)
	trolls := selected.Trolls
	for _, troll := range choices {
		toRemove := troll.Id
		for i := 0; i < len(trolls); i++ {
			if toRemove == trolls[i] {
				for j := i + 1; j < len(trolls); j++ {
					trolls[j-1] = trolls[j]
				}
				trolls = trolls[0 : len(trolls)-1]
			}
		}
	}
	selected.Trolls = trolls

	bank.Groups[selectedIndex] = selected
	bank.Persist()
	selected.PrintMembers(bank.Trolls)
}

func RenameGroup(bank *PxBank) {
	DisplayGroups(bank)

	// Old group to modify
	selected, selectedIndex := ChooseGroup(bank)

	// Change with?
	fmt.Print("New name ? ")
	newName := readLine()

	fmt.Printf("Change [%v] into [%v] (y/n)? ", selected.Name, newName)
	var choice string
	fmt.Scanln(&choice)
	if "y" == choice {
		bank.Groups[selectedIndex].Name = newName
		if nil == bank.Persist() {
			fmt.Println("**** Change successful")
		}
	}
}

func DeleteGroup(bank *PxBank) {
	DisplayGroups(bank)

	// Old group to modify
	selected, selectedIndex := ChooseGroup(bank)

	fmt.Printf("Delete Group [%v] (with %v trolls) (y/n)? ", selected.Name, len(selected.Trolls))
	var choice string
	fmt.Scanln(&choice)
	if "y" == choice {
		for i := selectedIndex + 1; i < len(bank.Groups); i++ {
			bank.Groups[i-1] = bank.Groups[i]
		}
		bank.Groups = bank.Groups[0 : len(bank.Groups)-1]
		if nil == bank.Persist() {
			fmt.Println("**** Removal successful")
		}
	}
}

func ChooseGroup(bank *PxBank) (selected Group, selectedIndex int) {
	var choice string
	fmt.Print("\nGroup name ? ")
	fmt.Scanln(&choice)
	matcher := newMatcher(choice)
	for _, possibleMatch := range bank.Groups {
		matcher.check(possibleMatch)
	}
	selected = matcher.bestMatch.(Group)
	selectedIndex = matcher.bestPosition

	fmt.Println("Selected: ", selected.Name)
	return
}

func DisplayGroups(bank *PxBank) {
	fmt.Println()
	for _, group := range bank.Groups {
		fmt.Println(group.Name, " : ", len(group.Trolls))
	}
}

func (group Group) PrintMembers(reference map[TrollId]Troll) {
	fmt.Printf("Members (%v):\n", len(group.Trolls))
	var totalForGroup PxAccount
	var shareable PxValue
	for _, trollId := range group.Trolls {
		troll := reference[trollId]
		fmt.Println(troll)
		totalForGroup += troll.Account
		shareable += troll.Shareable
	}
	fmt.Printf("\nTotal: %v\tShareable: %v\n", totalForGroup, shareable)
}

func CreateGroup(bank *PxBank) {
	fmt.Print("Group name > ")
	name := readLine()
	if 0 == len(name) {
		return
	}

	var group Group
	group.Name = name
	group.Trolls = make([]TrollId, 0, 20)

	bank.Groups = append(bank.Groups, group)
	bank.Persist()
}
