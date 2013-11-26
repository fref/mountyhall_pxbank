package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
)

type MenuDisplay func(*PxBank)

type OptionHandler func(*PxBank, string)

func (bank *PxBank) LoopMenu(display MenuDisplay, handle OptionHandler) {
	choice := ""
	for {
		display(bank)
		choice = readLine()
		if "0" == choice {
			return
		}
		fmt.Println("===")
		handle(bank, choice)
	}
}

func DisplayMainMenu(bank *PxBank) {
	PrintHeader(bank, "")
	fmt.Println("1. Trolls")
	fmt.Println("2. Groups")
	fmt.Println("3. Proposals")
	fmt.Println("4. Transfers")
	fmt.Println("")
	fmt.Println("0. Quit")
	fmt.Print("Choice: ")
}

func PrintHeader(bank *PxBank, section string) {
	fmt.Println("\n-----------------------------------------------------------------------------------")
	fmt.Print("PX BANK * ", section)
	fmt.Printf(" | %v Trolls known | %v Groups known | %v Proposals pending \n", len(bank.Trolls), len(bank.Groups), len(bank.Proposals))
	fmt.Println("-----------------------------------------------------------------------------------")
}

func HandleMainChoice(bank *PxBank, choice string) {
	switch choice {
	case "1":
		bank.LoopMenu(DisplayTrollMenu, HandleTrollAction)
		return
	case "2":
		bank.LoopMenu(DisplayGroupsMenu, HandleGroupAction)
		return
	case "3":
		bank.LoopMenu(DisplayProposalsMenu, HandleProposalAction)
		return
	case "4":
		bank.LoopMenu(DisplayTransfersMenu, HandleTransferAction)
		return
	}
	fmt.Println("Unknown choice")
}

type matcher struct {
	bestScore    int
	bestMatch    Named
	bestPosition int
	toFind       string
	index        int
}

func newMatcher(toFind string) matcher {
	return matcher{bestScore: math.MinInt32, toFind: normalize(toFind), bestPosition: -1}
}

func (m *matcher) check(possibleMatch Named) {
	matchName := normalize(possibleMatch.GetName())
	score := computeScore(m.toFind, matchName)
	if score > m.bestScore {
		m.bestMatch = possibleMatch
		m.bestScore = score
		m.bestPosition = m.index
	}
	m.index++
}

func computeScore(toFind, possibleMatch string) (score int) {
	if directMatch := strings.Index(possibleMatch, toFind); -1 < directMatch {
		if 0 == directMatch {
			score = 100
		} else {
			score = len(possibleMatch) - directMatch
		}
	}
	for i := 0; i < len(toFind)-1; i++ {
		looked := toFind[i : i+2]
		if strings.Contains(possibleMatch, looked) {
			score++
		} else {
		}
	}
	return
}

func normalize(toFind string) (normalized string) {
	normalized = strings.ToLower(toFind)
	normalized = strings.Replace(normalized, " ", "", 0)
	return
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	var line string
	var err error
	if line, err = reader.ReadString('\n'); err == nil {
		line = strings.TrimSpace(line)
	} else {
		fmt.Println(err)
	}

	return line
}
