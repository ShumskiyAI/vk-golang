package main

// сюда писать код

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

var (
	// @BotFather в телеграме даст вам токен. Если захотите потыкать своего бота через телегу - используйте именно его
	BotToken = "XXX"

	// Урл, в который будет стучаться телега при получении сообщения от пользователя.
	// Может быть как айпишником личной виртуалки, так и просто выдан сервисом для деплоя
	WebhookURL = "https://525f2cb5.ngrok.io"
)

// Пустой список задач - - - - - - - - -|| Нет задач
// Создание задачи - - - - - - - - - - -|| Задача "написать бота" создана, id=1
// Задача назначена на тебя- - - - - - -|| Задача "написать бота" назначена на вас
// Задача назначена на другого (автору) || Задача "написать бота" назначена на @aalexandrov
// Снятие задачи - - - - - - - - - - - -|| Принято
// Автору при снятии - - - - - - - - - -|| Задача "написать бота" осталась без исполнителя
// Выполнение -  - - - - - - - - - - - -|| Задача "написать бота" выполнена
// Автору при выполнении - - - - - - - -|| Задача "написать бота" выполнена @ppetrov
// Снятие чужой задачи - - - - - - - - -|| Задача не на вас

type Task struct {
	ID               int64
	Title            string
	OwnerID          int64
	OwnerNickname    string
	AssigneeID       int64
	AssigneeNickname string
}

type Bot struct {
	Tasks  []Task
	mu     *sync.RWMutex
	nextID int64
}

func NewBot() *Bot {
	return &Bot{
		Tasks:  make([]Task, 0),
		mu:     &sync.RWMutex{},
		nextID: 1,
	}
}

// /tasks
func (b Bot) taskList() []Task {
	b.mu.RLock()
	defer b.mu.RUnlock()
	tasks := make([]Task, len(b.Tasks))
	copy(tasks, b.Tasks)
	return tasks
}

// /new XXX YYY ZZZ - создаёт новую задачу
func (b *Bot) createTask(title string, owner *tgbotapi.User) *Task {
	b.mu.Lock()
	defer b.mu.Unlock()
	task := &Task{
		ID:            b.nextID,
		Title:         title,
		OwnerID:       owner.ID,
		OwnerNickname: owner.UserName,
	}
	b.Tasks = append(b.Tasks, *task)
	b.nextID++

	return task
}

// /assign_$ID - делает пользователя исполнителем задачи
func (b *Bot) updateAssignee(taskID, assigneeID int64, assigneeNickname string) *Task {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, task := range b.Tasks {
		if task.ID == taskID {
			b.Tasks[i].AssigneeID = assigneeID
			b.Tasks[i].AssigneeNickname = assigneeNickname
			return &b.Tasks[i]
		}
	}

	return nil
}

// /unassign_$ID - снимает задачу с текущего исполнителя
func (b *Bot) unassign(taskID, myID int64) (*Task, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, task := range b.Tasks {
		if task.ID == taskID {
			if task.AssigneeID != myID {
				return nil, fmt.Errorf("задача не на вас")
			}
			b.Tasks[i].AssigneeID = 0
			b.Tasks[i].AssigneeNickname = ""
			return &b.Tasks[i], nil
		}
	}
	return nil, fmt.Errorf("задача не найдена")
}

// /resolve_$ID - выполняет задачу, удаляет её из списка
func (b *Bot) resolveTask(taskID int64) *Task {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, task := range b.Tasks {
		if task.ID == taskID {
			removed := task
			b.Tasks = append(b.Tasks[:i], b.Tasks[i+1:]...)

			return &removed
		}
	}
	return nil
}

// /my - показывает задачи, которые назначены на меня
func (b Bot) meAssignee(myID int64) []Task {
	b.mu.RLock()
	defer b.mu.RUnlock()
	TasksOnMe := make([]Task, 0)

	for _, task := range b.Tasks {
		if task.AssigneeID == myID {
			TasksOnMe = append(TasksOnMe, task)
		}
	}

	if len(TasksOnMe) != 0 {
		return TasksOnMe
	}

	return []Task{}
}

// /owner - показывает задачи, которые были созданы мной. Подробности форматирования смотрите в тестах.
func (b Bot) meOwner(myID int64) []Task {
	b.mu.RLock()
	defer b.mu.RUnlock()
	myTasks := make([]Task, 0)

	for _, task := range b.Tasks {
		if task.OwnerID == myID {
			myTasks = append(myTasks, task)
		}
	}

	if len(myTasks) != 0 {
		return myTasks
	}

	return []Task{}
}

func (b *Bot) handleMessage(bot *tgbotapi.BotAPI, upd tgbotapi.Update) {
	msg := upd.Message
	if msg == nil {
		log.Fatalf("сообщение из бота = nil")
	}

	user := msg.From
	chatID := msg.Chat.ID
	text := msg.Text

	switch {
	case text == "":
	}

}

func startTaskBot(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return fmt.Errorf("")
	}

	fmt.Printf("Авторизован аккаунт: %s\n", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		return fmt.Errorf("")
	}

	_, err = bot.Request(wh)
	if err != nil {
		return fmt.Errorf("")
	}

	updates := bot.ListenForWebhook("/")

	go func() {
		log.Fatalln("http err:", http.ListenAndServe(":8080", nil))
	}()

	b := NewBot()

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("всё работает корректно"))
	})

	for {
		select {
		case <-ctx.Done():
			log.Println("бот завершён из-за закрытия контекста")
			return nil
		case upd, ok := <-updates:
			if !ok {
				return nil
			}
			if upd.Message == nil {
				continue
			}
			b.handleMessage(bot, upd)
		}
	}
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
