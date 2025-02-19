package gocliselect

import (
	"fmt"
	"log"

	"github.com/buger/goterm"
	"github.com/pkg/term"
)

// Raw input keycodes
var up byte = 65
var down byte = 66
var escape byte = 27
var enter byte = 13
var keys = map[byte]bool{
	up:   true,
	down: true,
}

type Menu struct {
	Prompt    string
	CursorPos int
	MenuItems []*MenuItem
}

type MenuItem struct {
	Text       string
	ID         string
	Selectable bool
	SubMenu    *Menu
}

func NewMenu(prompt string) *Menu {
	return &Menu{
		Prompt:    prompt,
		MenuItems: make([]*MenuItem, 0),
		CursorPos: -1,
	}
}

// AddItem will add a new menu option to the menu list
func (m *Menu) AddItem(option string, id string) *Menu {
	menuItem := &MenuItem{
		Text:       option,
		ID:         id,
		Selectable: true,
	}

	m.MenuItems = append(m.MenuItems, menuItem)
	// shows correct pos for menus starting from hint
	if m.CursorPos == -1 {
		m.CursorPos = len(m.MenuItems) - 1
	}
	return m
}

// AddHint will add a new hint line to the menu list
func (m *Menu) AddHint(hint string) *Menu {
	menuItem := &MenuItem{
		Text:       hint,
		Selectable: false,
	}

	m.MenuItems = append(m.MenuItems, menuItem)
	return m
}

// renderMenuItems prints the menu item list.
// Setting redraw to true will re-render the options list with updated current selection.
func (m *Menu) renderMenuItems(redraw bool) {
	// need this only we have more than 1 item
	if redraw && len(m.MenuItems) > 1 {
		// Move the cursor up n lines where n is the number of options, setting the new
		// location to start printing from, effectively redrawing the option list
		//
		// This is done by sending a VT100 escape code to the terminal
		// @see http://www.climagic.org/mirrors/VT100_Escape_Codes.html
		fmt.Printf("\033[%dA", len(m.MenuItems)-1)
	}

	for index, menuItem := range m.MenuItems {
		var newline = "\n"
		if index == len(m.MenuItems)-1 {
			// Adding a new line on the last option will move the cursor position out of range
			// For out redrawing
			newline = ""
		}

		menuItemText := menuItem.Text
		// if-condition was enough here but...
		switch menuItem.Selectable {
		case true:
			cursor := "  "
			if index == m.CursorPos {
				cursor = goterm.Color("> ", goterm.YELLOW)
				menuItemText = goterm.Color(menuItemText, goterm.YELLOW)
			}
			fmt.Printf("\r%s %s%s", cursor, menuItemText, newline)
		case false:
			fmt.Printf("\r%s%s", goterm.Color(menuItemText, goterm.CYAN), newline)
		}
	}
}

// Display will display the current menu options and awaits user selection
// It returns the users selected choice
func (m *Menu) Display() string {
	defer func() {
		// Show cursor again.
		fmt.Printf("\033[?25h")
	}()

	// adding blank item is the easiest way to keep everyting working in empty menus
	if len(m.MenuItems) == 0 {
		m.AddHint("")
	}
	fmt.Printf("%s\n", goterm.Color(goterm.Bold(m.Prompt)+":", goterm.CYAN))

	m.renderMenuItems(false)

	// Turn the terminal cursor off
	fmt.Printf("\033[?25l")

	for {
		// sorry im switch fanboy
		switch getInput() {
		case escape:
			fmt.Println("\r")
			return ""
		case enter:
			menuItem := m.MenuItems[m.CursorPos]
			m.CursorPos = -1
			m.renderMenuItems(true)
			fmt.Println("\r")
			return menuItem.ID
		case up:
			m.CursorPos = (m.CursorPos + len(m.MenuItems) - 1) % len(m.MenuItems)
			// prevent looping over non-selectable menus if you're going to make them for some reason
			iter := 0
			for !m.MenuItems[m.CursorPos].Selectable && iter < len(m.MenuItems) {
				m.CursorPos = (m.CursorPos + len(m.MenuItems) - 1) % len(m.MenuItems)
				iter++
			}
			m.renderMenuItems(true)
		case down:
			m.CursorPos = (m.CursorPos + 1) % len(m.MenuItems)
			// prevent looping over non-selectable menus if you're going to make them for some reason
			iter := 0
			for !m.MenuItems[m.CursorPos].Selectable && iter < len(m.MenuItems) {
				m.CursorPos = (m.CursorPos + 1) % len(m.MenuItems)
				iter++
			}
			m.renderMenuItems(true)
		}
	}
}

// getInput will read raw input from the terminal
// It returns the raw ASCII value inputted
func getInput() byte {
	t, _ := term.Open("/dev/tty")

	err := term.RawMode(t)
	if err != nil {
		log.Fatal(err)
	}

	var read int
	readBytes := make([]byte, 3)
	read, err = t.Read(readBytes)
	if err != nil {
		log.Fatal(err)
	}

	t.Restore()
	t.Close()

	// Arrow keys are prefixed with the ANSI escape code which take up the first two bytes.
	// The third byte is the key specific value we are looking for.
	// For example the left arrow key is '<esc>[A' while the right is '<esc>[C'
	// See: https://en.wikipedia.org/wiki/ANSI_escape_code
	if read == 3 {
		if _, ok := keys[readBytes[2]]; ok {
			return readBytes[2]
		}
	} else {
		return readBytes[0]
	}

	return 0
}
