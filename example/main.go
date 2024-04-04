package main

import (
	"fmt"

	"github.com/leByteBuster/gosimpleclimenu"
)

func main() {
	menu := gosimpleclimenu.NewMenu("Choose a colour")

	// TODO: allow to pass a construct which shows the structure of the menu and build a menu from it

	menu.AddItem("Red", "red")
	//menu.AddItem("Blue", "blue")
	//subMenu := gosimpleclimenu.NewMenu("Choose a Tone")
	//subMenu.AddItem("Dark", "dark")
	//subMenu.AddItem("Light", "light")
	//subMenu.AddItem("Light2", "light2")
	//subMenu.AddItem("Light3", "light3")
	//menu.AddSubmenuItem("Green", "green", subMenu)

	// menu.AddItem("Yellow", "yellow")
	// menu.AddItem("Cyan", "cyan")

	choice := menu.Display()

	fmt.Printf("Choice: %s\n", choice)
}
