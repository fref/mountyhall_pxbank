package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

func DisplayProposalsMenu(bank *PxBank) {
	PrintHeader(bank, "PROPOSALS")
	fmt.Println("1. Compute Proposal")
	fmt.Println("2. List Proposals")
	fmt.Println("3. Accept Proposal")
	fmt.Println("4. Reject Proposal")
	fmt.Println("0. Quit")
	fmt.Print("Choice: ")
}

func HandleProposalAction(bank *PxBank, choice string) {
	switch choice {
	case "1":
		ComputeProposal(bank)
		return
	case "2":
		ListProposals(bank)
		return
	case "3":
		AcceptProposal(bank)
		return
	case "4":
		RejectProposal(bank)
		return
	}
	fmt.Println("Unknown choice")

}

func ComputeProposal(bank *PxBank) {
	DisplayGroups(bank)
	selected, _ := ChooseGroup(bank)
	selected.PrintMembers(bank.Trolls)

	sharers := make(map[TrollId]*Troll)
	groupMembers := make(TrollSelection, 0, len(selected.Trolls))
	membersMap := make(map[TrollId]Troll)
	var totalToShare PxValue
	for _, trollId := range selected.Trolls {
		troll := bank.Trolls[trollId]
		groupMembers = append(groupMembers, &troll)
		membersMap[troll.Id] = troll
		if 0 < troll.Shareable {
			sharers[troll.Id] = &troll
			totalToShare += troll.Shareable
		}
	}

	// Integrate existing proposals to avoid computing weird values
	if 0 < len(bank.Proposals) {
		fmt.Println("Integrating existing proposals")
		for _, proposal := range bank.Proposals {
			if sharer, ok := sharers[proposal.Source]; ok {
				fmt.Printf("%v already supposed to share %v, removing from pool (proposal: %v)\n", bank.Trolls[sharer.Id].Name, proposal.Total, proposal.Id)
				sharer.Shareable -= proposal.Total
				totalToShare -= proposal.Total
				sharer.Account -= PxAccount(proposal.PerTroll)
				for _, recipient := range proposal.Destinations {
					if member, ok := membersMap[recipient]; ok {
						member.Account -= PxAccount(proposal.PerTroll)
					}
				}
			}
		}

	}

	if 0 >= totalToShare {
		fmt.Println("Nothing to share")
		return
	}

	for 0 < totalToShare {
		var biggestSharer *Troll
		biggestToShare := PxValue(math.MinInt32)
		for _, sharer := range sharers {
			if biggestToShare < sharer.Shareable {
				biggestSharer = sharer
				biggestToShare = sharer.Shareable
			}
		}
		fmt.Printf("Biggest sharer : %v \n", biggestSharer)
		bestPerTroll, bestCount, bestTotal := computeBestValues(biggestSharer, groupMembers)
		proposalRecipients, proposalRecipientsId := extractRecipients(biggestSharer, bestCount, bestPerTroll, groupMembers)
		//		fmt.Printf("bestPerTroll %v, bestCount %v, bestTotal %v\n", bestPerTroll, bestCount, bestTotal)
		if 0 == bestCount { // auto-share
			bestPerTroll = biggestToShare
		}
		if len(proposalRecipients) != bestCount {
			fmt.Printf("Inconsistent recipient count: %v whereas computed value was %v\n", len(proposalRecipients), bestCount)
			for _, member := range groupMembers {
				membersMap[member.Id] = *member
			}
			selected.PrintMembers(membersMap)
			return
		}
		transfer := NewTransfer(*biggestSharer, bestPerTroll, proposalRecipientsId, fmt.Sprintf("Transfer of %v px towards %v other trolls", bestTotal, bestCount))
		proposal := Proposal{Id: id(), Transfer: transfer}

		// Apply proposal
		totalToShare -= proposal.Total
		biggestSharer.Shareable -= proposal.Total
		biggestSharer.Account -= PxAccount(proposal.PerTroll)
		for _, recipient := range proposalRecipients {
			recipient.Account -= PxAccount(proposal.PerTroll)
		}
		updateMembersMap(membersMap, groupMembers)
		bank.Proposals = append(bank.Proposals, proposal)

		fmt.Println(proposal.ToString(bank))
		fmt.Println("Remaining to share: ", totalToShare)
	}

	selected.PrintMembers(membersMap)
	bank.Persist()
}

func updateMembersMap(membersMap map[TrollId]Troll, groupMembers TrollSelection) {
	for _, member := range groupMembers {
		membersMap[member.Id] = *member
	}

}

func computeBestValues(biggestSharer *Troll, groupMembers TrollSelection) (bestPerTroll PxValue, bestCount int, bestTotal PxValue) {
	for maxShareable := biggestSharer.Shareable / 2; maxShareable >= 1; maxShareable-- {
		perTroll, count := MaxValue(maxShareable, groupMembers, biggestSharer.Id)
		total := PxValue(1+count) * perTroll // must include the sharer
		if total > PxValue(biggestSharer.Shareable) {
			count = int(float32(biggestSharer.Shareable)/float32(perTroll)) - 1
		}
		total = PxValue(1+count) * perTroll // must include the sharer
		if total > bestTotal {
			bestPerTroll = perTroll
			bestCount = count
			bestTotal = total
		}
	}
	return
}

func extractRecipients(biggestSharer *Troll, bestCount int, bestPerTroll PxValue, groupMembers TrollSelection) (proposalRecipients TrollSelection, proposalRecipientsId []TrollId) {
	proposalRecipients = make(TrollSelection, 0, bestCount)
	proposalRecipientsId = make([]TrollId, 0, bestCount)

	// Make sure that those with something to share are paid last, and those with high accounts first
	sort.Sort(ByAccountShareable{groupMembers})

	for _, recipient := range groupMembers {
		if len(proposalRecipients) == bestCount {
			return
		}
		if recipient.Id == biggestSharer.Id || roundedAccount(recipient.Account) < bestPerTroll {
			continue
		}
		proposalRecipients = append(proposalRecipients, recipient)
		proposalRecipientsId = append(proposalRecipientsId, recipient.Id)
	}
	return
}

func MaxValue(smallerThan PxValue, recipients TrollSelection, source TrollId) (perTroll PxValue, count int) {
	for _, troll := range recipients {
		if troll.Id == source {
			continue
		}
		account := PxValue(troll.Account)
		if 1 <= account && account <= smallerThan && account >= perTroll {
			perTroll = account
		}
		rounded := roundedAccount(troll.Account)
		if 1 <= rounded && rounded <= smallerThan && rounded >= perTroll {
			perTroll = rounded
		}
	}
	for _, troll := range recipients {
		if troll.Id == source {
			continue
		}
		if roundedAccount(troll.Account) >= perTroll {
			count++
		}
	}
	return
}

func roundedAccount(account PxAccount) PxValue {
	value := PxValue(math.Ceil(float64(account)))
	return value
}

func ListProposals(bank *PxBank) {
	for _, proposal := range bank.Proposals {
		fmt.Println(proposal.ToString(bank))
	}

}

func AcceptProposal(bank *PxBank) {
	ListProposals(bank)
	selected, selectedIndex := chooseProposal(bank)

	fmt.Print("Approve (y)? ")
	confirm := readLine()
	if "y" == confirm {
		t := Transaction{Date: time.Now(), Transfer: selected.Transfer}
		t.AdjustAccounts(bank)
		t.Persist()
		removeProposal(bank, selectedIndex)
		bank.Persist()
	}
}

func RejectProposal(bank *PxBank) {
	ListProposals(bank)
	_, selectedIndex := chooseProposal(bank)

	fmt.Print("Reject (y)? ")
	confirm := readLine()
	if "y" == confirm {
		removeProposal(bank, selectedIndex)
		bank.Persist()
	}
}

func chooseProposal(bank *PxBank) (selected Proposal, selectedIndex int) {
	if 0 == len(bank.Proposals) {
		return
	}
	var choice string
	fmt.Print("\nProposal id ? ")
	fmt.Scanln(&choice)
	matcher := newMatcher(choice)
	for _, possibleMatch := range bank.Proposals {
		matcher.check(possibleMatch)
	}

	selected = matcher.bestMatch.(Proposal)
	selectedIndex = matcher.bestPosition

	fmt.Println("Selected: ", selected.ToString(bank))
	return
}

func removeProposal(bank *PxBank, position int) {
	for j := position + 1; j < len(bank.Proposals); j++ {
		bank.Proposals[j-1] = bank.Proposals[j]
	}
	bank.Proposals = bank.Proposals[0 : len(bank.Proposals)-1]
}

// Sorting helper
type ByAccountShareable struct {
	TrollSelection
}

// See sort.Interface
// Sorts by Shareable increasing then by Account decreasing
func (s ByAccountShareable) Less(i, j int) bool {
	left := s.TrollSelection[i]
	right := s.TrollSelection[j]
	if left.Shareable == right.Shareable {
		return left.Account > right.Account
	}
	return left.Shareable < right.Shareable
}
