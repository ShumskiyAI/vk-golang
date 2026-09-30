package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var world = make(map[string]*Room)
var player *Player

type Player struct {
	location *Room
	items    []Item
}

type Room struct {
	title   string
	items   map[string][]Item
	exits   map[string]func() string
	look    func() string
	actions map[string]func(string) string
	locked  bool
}

type Item struct {
	title    string
	wearable bool
}

func NewPlayer() *Player {
	return &Player{
		items: make([]Item, 0, 4),
	}
}

func NewRoom(title string) *Room {
	return &Room{
		title:   title,
		items:   make(map[string][]Item, 2),
		exits:   make(map[string]func() string, 1),
		actions: make(map[string]func(string) string),
		locked:  false,
	}
}

func NewItem(title string, wearable bool) Item {
	return Item{
		title:    title,
		wearable: wearable,
	}
}

func addRoomsInWorld(rooms ...*Room) {
	for _, room := range rooms {
		world[room.title] = room
	}
}

func makeItemsTitle(items []Item) string {
	titles := make([]string, len(items))

	for i := range items {
		titles[i] = items[i].title
	}

	return strings.Join(titles, ", ")
}

// эта функция инициализирует игровой мир - все комнаты
// если что-то было - оно корректно перезатирается
func initGame() {
	world = make(map[string]*Room)
	p := NewPlayer()

	kitchen := NewRoom("кухня")
	corridor := NewRoom("коридор")
	bedroom := NewRoom("комната")
	street := NewRoom("улица")

	corridor.locked = true
	corridorDescription := "ничего интересного. можно пройти - кухня, комната, улица"

	addRoomsInWorld(kitchen, corridor, bedroom, street)

	p.location = kitchen
	player = p

	// -------------- КУХНЯ --------------
	kitchen.items["на столе"] = []Item{
		NewItem("чай", false),
	}

	kitchen.exits["коридор"] = func() string {
		p.location = world["коридор"]
		return corridorDescription
	}

	kitchen.look = func() string {
		titles := makeItemsTitle(kitchen.items["на столе"])

		if !p.hasItem("рюкзак") {
			return fmt.Sprintf("ты находишься на кухне, на столе: %s, надо собрать рюкзак и идти в универ. можно пройти - коридор", titles)
		}

		return fmt.Sprintf("ты находишься на кухне, на столе: %s, надо идти в универ. можно пройти - коридор", titles)
	}

	// kitchen.actions[""] = func(nothing string) string { return "" } // можно придумать самому

	// -------------- КОРИДОР --------------

	// corridor.items[""] = []Item{} // можно придумать самому
	corridor.exits["кухня"] = func() string {
		p.location = world["кухня"]
		return "кухня, ничего интересного. можно пройти - коридор"
	}

	corridor.exits["комната"] = func() string {
		p.location = world["комната"]
		return "ты в своей комнате. можно пройти - коридор"
	}

	corridor.exits["улица"] = func() string {
		if corridor.locked {
			return "дверь закрыта"
		}

		p.location = world["улица"]
		return "на улице весна. можно пройти - домой"
	}

	// corridor.look = func() string { titles := makeItemsTitle(kitchen.items[""]) return fmt.Sprintf("") } // можно придумать самому

	corridor.actions["ключи"] = func(object string) string {
		if object == "дверь" {
			corridor.locked = false
			return "дверь открыта"
		}

		return "не к чему применить"
	}

	// -------------- КОМНАТА --------------

	bedroom.items["на столе"] = []Item{
		NewItem("ключи", false),
		NewItem("конспекты", false),
	}
	bedroom.items["на стуле"] = []Item{
		NewItem("рюкзак", true),
	}

	bedroom.exits["коридор"] = func() string {
		p.location = world["коридор"]
		return corridorDescription
	}

	bedroom.look = func() string {
		table := makeItemsTitle(bedroom.items["на столе"])
		chair := makeItemsTitle(bedroom.items["на стуле"])

		switch {
		case table == "" && chair == "":
			return "пустая комната. можно пройти - коридор"

		case table == "" && chair != "":
			return fmt.Sprintf("на стуле: %s. можно пройти - коридор", chair)

		case table != "" && chair == "":
			return fmt.Sprintf("на столе: %s. можно пройти - коридор", table)
		}

		return fmt.Sprintf("на столе: %s, на стуле: %s. можно пройти - коридор", table, chair)
	}

	// bedroom.actions[""] = func(object string) string {
	// 	return fmt.Sprintf("%s", object) // в тестах этого нет, добавил от себя
	// }

	// -------------- УЛИЦА --------------

	street.exits["домой"] = func() string {
		p.location = world["коридор"]
		return corridorDescription
	}
}

// идти
func (p *Player) goTo(move string) string {
	exit, exists := p.location.exits[move]
	if !exists {
		return fmt.Sprintf("нет пути в %s", move)
	}

	return exit()
}

func (p Player) hasItem(object string) bool {
	for _, item := range p.items {
		if item.title == object {
			return true
		}
	}
	return false
}

// надеть
func (p *Player) putOn(object string) string {
	for place, items := range p.location.items { // вынести в отдельную функцию
		for i, item := range items {
			if item.title == object {
				if item.wearable {
					items = append(items[:i], items[i+1:]...)
					p.location.items[place] = items

					p.items = append(p.items, item)

					return fmt.Sprintf("вы надели: %s", item.title)
				}
				return fmt.Sprintf("%s нельзя надеть", item.title)
			}
		}
	}

	return "нет такого"
}

// взять
func (p *Player) take(object string) string {
	if !p.hasItem("рюкзак") {
		return "некуда класть"
	}

	for place, items := range p.location.items {
		for i, item := range items {
			if item.title == object {
				items = append(items[:i], items[i+1:]...)
				p.location.items[place] = items

				p.items = append(p.items, item)

				return fmt.Sprintf("предмет добавлен в инвентарь: %s", item.title)
			}
		}
	}
	return "нет такого"
}

// применить
func (p *Player) use(args []string) string {
	item := args[0]
	object := args[1]

	if !p.hasItem(args[0]) {
		return fmt.Sprintf("нет предмета в инвентаре - %s", item)
	}

	action, exists := p.location.actions[item]
	if !exists {
		return "не к чему применить"
	}

	return action(object)
}

// данная функция принимает команду от "пользователя"
// и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
func handleCommand(command string) string {
	strs := strings.Fields(command)
	if len(strs) == 0 {
		return "команда не указана"
	}

	action := strs[0]

	hasParam := func(strs []string) bool {
		return len(strs) > 1
	}

	switch action {
	case "идти":
		if !hasParam(strs) {
			return "недостаточно параметров, куда идти?"
		}
		return player.goTo(strs[1])

	case "осмотреться":
		return player.location.look()

	case "надеть":
		if !hasParam(strs) {
			return "недостаточно параметров, что надевать?"
		}
		return player.putOn(strs[1])

	case "взять":
		if !hasParam(strs) {
			return "недостаточно параметров, что брать?"
		}
		return player.take(strs[1])

	case "применить":
		if !hasParam(strs) {
			return "недостаточно параметров, что и к чему применить?"
		}
		return player.use(strs[1:])

	}

	return "неизвестная команда"
}

// в этой функции можно ничего не писать,
// но тогда у вас не будет работать через go run main.go
// очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
func main() {
	initGame()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		command := scanner.Text()

		fmt.Println(handleCommand(command))

		fmt.Print("> ")
		if scanner.Err() != nil {
			if scanner.Err() != io.EOF {
				fmt.Printf("произошла ошибка ввода: %v", scanner.Err())
				return
			}
		}
	}
}
