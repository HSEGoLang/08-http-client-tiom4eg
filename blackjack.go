package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	newDeckURL = "https://deckofcardsapi.com/api/deck/new/shuffle/?deck_count=1"
	drawURL    = "https://deckofcardsapi.com/api/deck/%s/draw/?count=1"
)

type Deck struct {
	Success bool   `json:"success"`
	DeckID  string `json:"deck_id"`
}

type Card struct {
	Code  string `json:"code"`
	Value string `json:"value"`
	Suit  string `json:"suit"`
}

type Draw struct {
	Success bool   `json:"success"`
	Cards   []Card `json:"cards"`
}

func formatCard(c Card) string {
	return fmt.Sprintf("%s (%s of %s)", c.Code, c.Value, c.Suit)
}

func printHand(hand []Card) {
	for _, c := range hand {
		fmt.Println(" ", formatCard(c))
	}
}

func handValue(hand []Card) int {
	sum := 0
	aces := 0
	for _, c := range hand {
		switch c.Value {
		case "ACE":
			sum += 11
			aces++
		case "KING", "QUEEN", "JACK":
			sum += 10
		default:
			var val int
			fmt.Sscanf(c.Value, "%d", &val)
			sum += val
		}
	}
	for sum > 21 && aces > 0 {
		sum -= 10
		aces--
	}
	return sum
}

func draw(deckID string) Card {
	url := fmt.Sprintf(drawURL, deckID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("draw error:", err)
		return Card{}
	}
	defer resp.Body.Close()
	var dr Draw
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		fmt.Println("draw error:", err)
		return Card{}
	}
	return dr.Cards[0]
}

func newDeck() (string, error) {
	resp, err := http.Get(newDeckURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var d Deck
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return "", err
	}
	if !d.Success {
		return "", fmt.Errorf("api failed")
	}
	return d.DeckID, nil
}

func main() {
	deckID, err := newDeck()
	if err != nil {
		fmt.Println("deck error:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	player := []Card{}
	dealer := []Card{}

	player = append(player, draw(deckID), draw(deckID))
	dealer = append(dealer, draw(deckID), draw(deckID))

	fmt.Println("Играем в блекджек!\nВзять ещё карту: hit/хит, остановиться: стоп/stop\n")
	for {
		fmt.Println("\nВаши карты:")
		printHand(player)
		fmt.Printf("\nСумма: %d\n", handValue(player))

		if handValue(player) > 21 {
			fmt.Println("Перебор, вы проиграли")
			return
		}

		fmt.Print("Ваш ход: ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)

		if cmd == "hit" || cmd == "хит" {
			player = append(player, draw(deckID))
		} else if cmd == "stand" || cmd == "стоп" {
			break
		} else {
			fmt.Println("Неверный ход")
		}
	}

	fmt.Println("Ход дилера\n")
	for {
		fmt.Println("\nКарты дилера:")
		printHand(dealer)
		dv := handValue(dealer)
		fmt.Printf("\nСумма дилера: %d\n", dv)
		if dv >= 17 {
			break
		}
		fmt.Println("Дилер берёт карту")
		dealer = append(dealer, draw(deckID))
	}

	pt, dt := handValue(player), handValue(dealer)
	fmt.Println("\nРезультат:\n")
	fmt.Printf("Ваши карты (%d):\n", pt)
	printHand(player)
	fmt.Printf("Карты дилера (%d):\n", dt)
	printHand(dealer)

	switch {
	case pt > 21:
		fmt.Println("У вас перебор, вы проиграли")
	case dt > 21:
		fmt.Println("У дилера перебор, вы выиграли")
	case pt > dt:
		fmt.Println("Вы ближе к 21, вы выиграли")
	case pt < dt:
		fmt.Println("Дилер ближе к 21, вы проиграли")
	default:
		fmt.Println("Ничья")
	}
}
