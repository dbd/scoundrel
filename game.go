package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Suit int

const (
	Hearts Suit = iota
	Diamonds
	Clubs
	Spades
)

func (s Suit) String() string {
	return [...]string{"♥", "♦", "♣", "♠"}[s]
}

func (s Suit) ASCII() string {
	switch s {
	case Hearts:
		return "  ***   ***  \n ************\n ************\n  **********\n   ********\n    ******\n     ****\n      **\n            \n            \n            "
	case Diamonds:
		return "      **\n     ****\n    ******\n   ********\n  **********\n   ********\n    ******\n     ****\n      **\n            \n            "
	case Clubs:
		return "     ****\n    ******\n    ******\n     ****\n  ** **** **\n ************\n ************\n  **********\n      **\n     ****\n    ******"
	case Spades:
		return "      **\n     ****\n    ******\n   ********\n  **********\n ************\n ************\n  **********\n      **\n     ****\n    ******"
	default:
		return ""
	}
}

func (s Suit) Color() lipgloss.Color {
	if s == Hearts || s == Diamonds {
		return lipgloss.Color("#cc0000")
	}
	return lipgloss.Color("#000000")
}

func (s Suit) Background() lipgloss.Color {
	return lipgloss.Color("#ffffff")
}

type Card struct {
	Suit  Suit
	Value int
}

func (c Card) String() string {
	valueStr := ""
	switch c.Value {
	case 14:
		valueStr = "A"
	case 11:
		valueStr = "J"
	case 12:
		valueStr = "Q"
	case 13:
		valueStr = "K"
	default:
		valueStr = fmt.Sprintf("%d", c.Value)
	}
	return fmt.Sprintf("%s%s", valueStr, c.Suit)
}

type Weapon struct {
	Value      int
	MaxKillVal int
}

type model struct {
	deck          []Card
	currentRoom   []Card
	player        int
	weapon        *Weapon
	ranLastRoom   bool
	gameOver      bool
	won           bool
	message       string
	selectedCards map[int]bool
	roomsCleared  int
}

func newDeck() []Card {
	var deck []Card
	for suit := Hearts; suit <= Spades; suit++ {
		for value := 2; value <= 13; value++ {
			if (suit == Hearts || suit == Diamonds) && value >= 11 {
				continue
			}
			deck = append(deck, Card{Suit: suit, Value: value})
		}
		// Add Ace (value 14) only for black suits
		if suit == Clubs || suit == Spades {
			deck = append(deck, Card{Suit: suit, Value: 14})
		}
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	return deck
}

func initialModel() model {
	deck := newDeck()
	room := deck[:4]
	deck = deck[4:]

	return model{
		deck:          deck,
		currentRoom:   room,
		player:        20,
		weapon:        nil,
		ranLastRoom:   false,
		gameOver:      false,
		won:           false,
		message:       "Select cards 1-4 in order (select 3 to face room or R to run)",
		selectedCards: make(map[int]bool),
		roomsCleared:  0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.gameOver {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			if msg.String() == "n" {
				return initialModel(), nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "1", "2", "3", "4":
			idx := int(msg.String()[0] - '1')
			if idx < len(m.currentRoom) && !m.selectedCards[idx] {
				card := m.currentRoom[idx]
				m.selectedCards[idx] = true
				m = m.handleCard(card)
				
				if m.gameOver {
					return m, nil
				}
				
				if len(m.selectedCards) == 3 {
					m.ranLastRoom = false
					m.roomsCleared++
					
					// Find the unselected card
					var unselectedCard Card
					for i := 0; i < 4; i++ {
						if !m.selectedCards[i] {
							unselectedCard = m.currentRoom[i]
							break
						}
					}
					
					m.selectedCards = make(map[int]bool)

					if len(m.deck) < 3 {
						m.gameOver = true
						m.won = true
						m.message = "You cleared the dungeon! You win!"
						return m, nil
					}

					// New room is the unselected card plus 3 new cards
					newRoom := []Card{unselectedCard}
					newRoom = append(newRoom, m.deck[:3]...)
					m.currentRoom = newRoom
					m.deck = m.deck[3:]
					m.message = "Room cleared! New room ahead."
				}
			}

		case "r":
			if m.ranLastRoom {
				m.message = "You can't run from two rooms in a row!"
				return m, nil
			}

			if len(m.selectedCards) > 0 {
				m.message = "You can't run after selecting cards!"
				return m, nil
			}

			if len(m.deck) < 4 {
				m.message = "Not enough cards to run!"
				return m, nil
			}

			m.deck = append(m.deck, m.currentRoom...)
			rand.Shuffle(len(m.deck), func(i, j int) {
				m.deck[i], m.deck[j] = m.deck[j], m.deck[i]
			})
			m.currentRoom = m.deck[:4]
			m.deck = m.deck[4:]
			m.ranLastRoom = true
			m.selectedCards = make(map[int]bool)
			m.message = "You ran away and shuffled into a new room!"
		}
	}

	return m, nil
}

func (m model) handleCard(card Card) model {
	switch card.Suit {
	case Hearts:
		m.player += card.Value
		m.message = fmt.Sprintf("Healed for %d HP!", card.Value)

	case Diamonds:
		m.weapon = &Weapon{Value: card.Value, MaxKillVal: 14}
		m.message = fmt.Sprintf("Picked up weapon with %d damage!", card.Value)

	case Clubs, Spades:
		canUseWeapon := m.weapon != nil && card.Value <= m.weapon.MaxKillVal
		
		var damage int
		if canUseWeapon {
			damage = card.Value - m.weapon.Value
			if damage < 0 {
				damage = 0
			}
			m.player -= damage
			m.weapon.MaxKillVal = card.Value - 1
			m.message = fmt.Sprintf("Used weapon! Took %d damage. Weapon dulled (can block ≤%d)", damage, m.weapon.MaxKillVal)
			if m.weapon.MaxKillVal < 0 {
				m.weapon = nil
				m.message += " - weapon broke!"
			}
		} else {
			damage = card.Value
			m.player -= damage
			if m.weapon != nil {
				m.message = fmt.Sprintf("Weapon too dull (can only block ≤%d)! Took %d damage!", m.weapon.MaxKillVal, damage)
			} else {
				m.message = fmt.Sprintf("No weapon! Took %d damage!", damage)
			}
		}

		if m.player <= 0 {
			m.gameOver = true
			m.won = false
			m.message = "You died! Game Over."
		}
	}

	return m
}

func (m model) View() string {
	if m.gameOver {
		var s strings.Builder
		s.WriteString("\n")
		if m.won {
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("10")).
				MarginBottom(1)
			s.WriteString(titleStyle.Render("🎉 VICTORY! 🎉"))
		} else {
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("9")).
				MarginBottom(1)
			s.WriteString(titleStyle.Render("💀 GAME OVER 💀"))
		}
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("Rooms Cleared: %d\n", m.roomsCleared))
		s.WriteString(fmt.Sprintf("Final HP: %d\n\n", m.player))
		s.WriteString("Press 'n' for new game, 'q' to quit\n")
		return s.String()
	}

	var s strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)

	s.WriteString(headerStyle.Render("SCOUNDREL"))
	s.WriteString("\n\n")

	hpColor := "10"
	if m.player <= 5 {
		hpColor = "9"
	} else if m.player <= 10 {
		hpColor = "11"
	}
	hpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hpColor))
	s.WriteString(fmt.Sprintf("HP: %s", hpStyle.Render(fmt.Sprintf("%d", m.player))))
	s.WriteString(" | ")

	if m.weapon != nil {
		s.WriteString(fmt.Sprintf("Weapon: %d (max kill: %d)", m.weapon.Value, m.weapon.MaxKillVal))
	} else {
		s.WriteString("Weapon: None (fists: 0)")
	}

	s.WriteString(fmt.Sprintf(" | Rooms: %d", m.roomsCleared))
	s.WriteString(fmt.Sprintf(" | Cards left: %d", len(m.deck)))
	s.WriteString("\n\n")

	s.WriteString("Current Room:\n\n")
	
	var toggles []string
	var cards []string
	
	for i, card := range m.currentRoom {
		valueStr := ""
		switch card.Value {
		case 14:
			valueStr = "A"
		case 11:
			valueStr = "J"
		case 12:
			valueStr = "Q"
		case 13:
			valueStr = "K"
		default:
			valueStr = fmt.Sprintf("%d", card.Value)
		}

		toggleStyle := lipgloss.NewStyle().
			Width(18).
			Align(lipgloss.Center).
			Bold(true)
		
		if m.selectedCards[i] {
			toggleStyle = toggleStyle.Foreground(lipgloss.Color("#00ff00"))
			toggles = append(toggles, toggleStyle.Render(fmt.Sprintf("[%d] SELECTED", i+1)))
		} else {
			toggleStyle = toggleStyle.Foreground(lipgloss.Color("#888888"))
			toggles = append(toggles, toggleStyle.Render(fmt.Sprintf("[%d]", i+1)))
		}

		artLines := strings.Split(card.Suit.ASCII(), "\n")
		
		maxArtWidth := 0
		for _, line := range artLines {
			if len(line) > maxArtWidth {
				maxArtWidth = len(line)
			}
		}
		
		cardWidth := maxArtWidth
		if cardWidth < 14 {
			cardWidth = 14
		}
		
		lines := []string{
			fmt.Sprintf("%-*s", cardWidth, valueStr),
			"",
		}
		
		lines = append(lines, artLines...)
		
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("%*s", cardWidth, valueStr))
		
		cardContent := strings.Join(lines, "\n")

		cardStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 1).
			Foreground(card.Suit.Color()).
			Background(card.Suit.Background()).
			Bold(true)

		if m.selectedCards[i] {
			cardStyle = cardStyle.
				BorderForeground(lipgloss.Color("#00ff00")).
				BorderStyle(lipgloss.ThickBorder())
		} else {
			cardStyle = cardStyle.BorderForeground(lipgloss.Color("#444444"))
		}

		cards = append(cards, cardStyle.Render(cardContent))
	}
	
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, toggles...))
	s.WriteString("\n")
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cards...))
	s.WriteString("\n\n")

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Italic(true)
	s.WriteString(messageStyle.Render(m.message))
	s.WriteString("\n\n")

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	s.WriteString(helpStyle.Render("Commands:"))
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("  1-4: Select card (applies immediately)"))
	s.WriteString("\n")

	if m.ranLastRoom {
		s.WriteString(helpStyle.Render("  R: Run (unavailable - ran last room)"))
	} else if len(m.selectedCards) > 0 {
		s.WriteString(helpStyle.Render("  R: Run (unavailable - cards already selected)"))
	} else {
		s.WriteString(helpStyle.Render("  R: Run (shuffle and get new room)"))
	}
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("  Q: Quit"))
	s.WriteString("\n")

	return s.String()
}


