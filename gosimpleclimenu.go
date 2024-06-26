package gosimpleclimenu

import (
	"fmt"
	"log"

	"github.com/buger/goterm"
	"github.com/pkg/term"
)

// Raw input keycodes
var up byte = 65
var down byte = 66
var vimDown byte = 106
var vimUp byte = 107
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
	Text    string
	ID      string
	SubMenu *Menu
}

func NewMenu(prompt string) *Menu {
	return &Menu{
		Prompt:    prompt,
		MenuItems: make([]*MenuItem, 0),
	}
}

// AddItem will add a new menu option to the menu list
func (m *Menu) AddItem(option string, id string) {
	if id == "" {
		fmt.Println("ID must not be empty.")
		return
	}
	menuItem := &MenuItem{
		Text: option,
		ID:   id,
	}

	m.MenuItems = append(m.MenuItems, menuItem)
}

// func (m *Menu) AddSubmenuItem(option string, id string, submenu *Menu) {
func (m *Menu) AddSubmenuItem(option string, id string, submenu *Menu) {
	if id == "" {
		fmt.Println("ID must not be empty.")
		return
	}
	menuItem := &MenuItem{
		Text:    option,
		ID:      id,
		SubMenu: submenu,
	}

	m.MenuItems = append(m.MenuItems, menuItem)
}

func (m *Menu) drawHeading() {
	fmt.Printf("%s\n", goterm.Color(goterm.Bold(m.Prompt)+":", goterm.GREEN))
}

func (m *Menu) clearHeading() {
	if len(m.MenuItems) > 1 {
		fmt.Printf("\033[%dA", len(m.MenuItems)-1+1)
		fmt.Printf("\033[2K")
	}
	// else if len(m.MenuItems) == 1 {
	// 	fmt.Printf("\033[%dA", len(m.MenuItems)-1+1)
	// 	fmt.Printf("\033[2K")
	// }
}

// renderMenuItems prints the menu item list.
// Setting redraw to true will re-render the options list with updated current selection.
func (m *Menu) renderMenuItems(redraw bool) {
	if redraw {
		// Move the cursor up n lines (n is the number of menu options), setting the new
		// location to start printing from, effectively redrawing the option list.
		// This is necessary to redraw the menu at the same position instead of appending it
		// beneath the existing one.
		//
		// This is done by sending a VT100 escape code to the terminal
		// @see http://www.climagic.org/mirrors/VT100_Escape_Codes.html
		if len(m.MenuItems) > 1 {
			fmt.Printf("\033[%dA", len(m.MenuItems)-1)
		}
	}

	for index, menuItem := range m.MenuItems {
		var newline = "\n"
		if index == len(m.MenuItems)-1 {
			// Adding a new line on the lat option will move the cursor position out of range triggering redraw
			// so we dont do it.

			// TODO: this line was responsible for the bug if only one item to choose from exists. The menu was
			// drawn into negative (same element was drawn on top) each time one of the navigation keys (up
			// or down) were triggered
			// the bug is removed by adding the condition 'if (len(m.MenuItems) - 1) > 1' within the 'redraw'
			// block
			// why ??
			// somehow, if there is only a single element to choose from, and the up and down buttons are used to
			// try to navigate the menu is drawn into negative. Don't know why. If the sinlge element is not drawn
			// into the last ligne anymore this does not happen anymore. Thus we need the \n instead the "".
			newline = ""
		}

		menuItemText := menuItem.Text
		cursor := "  "
		if index == m.CursorPos {
			cursor = goterm.Color("> ", goterm.GREEN)
			menuItemText = goterm.Color(menuItemText, goterm.GREEN)
		}

		fmt.Printf("\r%s %s%s", cursor, menuItemText, newline)
	}
}

// Display will display the current menu options and awaits user selection
// It returns the users selected choice
func (m *Menu) Display() []string {

	if len(m.MenuItems) == 0 {
		fmt.Printf("No items added to menu. Add items.")
		return nil
	}

	defer func() {
		// Show cursor again.
		fmt.Printf("\033[?25h")
	}()

	m.drawHeading()

	m.renderMenuItems(false)

	// Turn the terminal cursor off
	fmt.Printf("\033[?25l")

	for {
		keyCode := getInput()

		if keyCode == escape {
			return []string{}
		} else if keyCode == enter {
			menuItem := m.MenuItems[m.CursorPos]
			fmt.Println("\r")
			sm := menuItem.SubMenu
			if sm != nil {

				// clear menu so it can be overwritten by submenu
				clearMenu(len(m.MenuItems))

				// show submenu
				elIdsSub := sm.Display()

				// if no elements are in array ESC was pressed in submenu. clear submenu. rerender menu
				if len(elIdsSub) == 0 {
					fmt.Println()
					clearMenu(len(sm.MenuItems))
					m.drawHeading()
					m.renderMenuItems(false)
				} else {
					return append([]string{menuItem.ID}, elIdsSub...)
				}
			} else {
				return []string{menuItem.ID}
			}
		} else if keyCode == up || keyCode == vimUp {
			m.CursorPos = (m.CursorPos + len(m.MenuItems) - 1) % len(m.MenuItems)
			m.renderMenuItems(true)
		} else if keyCode == down || keyCode == vimDown {
			m.CursorPos = (m.CursorPos + 1) % len(m.MenuItems)
			m.renderMenuItems(true)
		}
	}
}

// getInput will read raw input from the terminal
// It returns the raw ASCII value inputted
// NOTE:
// the way getInput works does not differ between 'A' and '<esc>[A' as input. Both would return
// an A as input which is treated as 'up'. Same accounts for '<esc>[C' and 'C' with 'down'.
// Proposal: return the full code and compare it to the actual full ascii (escape) codes
func getInput() byte {
	t, _ := term.Open("/dev/tty")

	err := term.RawMode(t)
	if err != nil {
		log.Fatal(err)
	}

	var read int
	readBytes := make([]byte, 3)
	read, err = t.Read(readBytes)

	t.Restore()
	t.Close()

	// Arrow keys are prefixed with the ANSI escape code which take up the first two bytes.
	// The third byte is the key specific value we are looking for.
	// For example the left arrow key is '<esc>[A' while the right is '<esc>[C'
	// See: https://en.wikipedia.org/wiki/ANSI_escape_code
	if read == 3 {
		if _, ok := keys[readBytes[2]]; ok {
			// we are only interested in the third byte in case its three bytes
			return readBytes[2]
		}
	} else {
		// in any other case we only return the first byte
		return readBytes[0]
	}
	return 0
}

// clears the menu from the cli (incl. heading)
func clearMenu(itemCount int) {
	// move cursor up n lines
	fmt.Printf("\033[%dA", itemCount+1)
	// move cursor to beginning of line
	fmt.Printf("\033[1G")
	// delte till end of screen
	fmt.Printf("\033[0J")
}
