package main

// сюда писать код

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
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

// /tasks 		 	  || Пустой список задач - - - - - - - - -|| Нет задач
// /new	  		 	  || Создание задачи - - - - - - - - - - -|| Задача "написать бота" создана, id=1

// /assign_$ID 	 	  || Задача назначена на тебя- - - - - - -|| Задача "написать бота" назначена на вас
// /assign_$ID 	 	  || Задача назначена на другого (автору) || Задача "написать бота" назначена на @aalexandrov

// /unassign_$ID 	  || Снятие задачи - - - - - - - - - - - -|| Принято
// /unassign_$ID 	  || Автору при снятии - - - - - - - - - -|| Задача "написать бота" осталась без исполнителя
// /unassign_$ID 	  || Снятие чужой задачи - - - - - - - - -|| Задача не на вас

// /handleResolveTask || Выполнение -  - - - - - - - - - - - -|| Задача "написать бота" выполнена
// /handleResolveTask || Автору при выполнении - - - - - - - -|| Задача "написать бота" выполнена @ppetrov

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
func (b Bot) handleTaskList() []Task {
	b.mu.RLock()
	defer b.mu.RUnlock()
	tasks := make([]Task, len(b.Tasks))
	copy(tasks, b.Tasks)
	return tasks
}

// /new XXX YYY ZZZ - создаёт новую задачу
func (b *Bot) handleCreateTask(title string, owner *tgbotapi.User) *Task {
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
func (b *Bot) handleUpdateAssignee(taskID, assigneeID int64, assigneeNickname string) (*Task, *Task) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, task := range b.Tasks {
		if task.ID == taskID {
			old := task
			b.Tasks[i].AssigneeID = assigneeID
			b.Tasks[i].AssigneeNickname = assigneeNickname
			return &b.Tasks[i], &old
		}
	}

	return nil, nil
}

// /unassign_$ID - снимает задачу с текущего исполнителя
func (b *Bot) handleUnassign(taskID, myID int64) (*Task, error) {
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
func (b *Bot) handleResolveTask(taskID int64) *Task {
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
func (b Bot) handleMeAssignee(myID int64) []Task {
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
func (b Bot) handleMeOwner(myID int64) []Task {
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
		log.Println("сообщение из бота = nil")
		return
	}

	user := msg.From
	chatID := msg.Chat.ID
	text := msg.Text

	send := func(text string) {
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
			log.Fatalln("ошибка отправки")
		}
	}

	switch {
	case text == "/tasks":
		tasks := b.handleTaskList()
		send("") // форматированный вывод
		_ = tasks

	case strings.HasPrefix(text, "/new "):
		title := strings.TrimPrefix(text, "/new ")
		b.handleCreateTask(title, user)
		send("") // форматированный вывод

	case strings.HasPrefix(text, "/assign"):

	case strings.HasPrefix(text, "/unassign"):

	case strings.HasPrefix(text, "/resolve"):

	case strings.HasPrefix(text, "/my"):

	case strings.HasPrefix(text, "/owner"):
		// /assign_$ID
		// /unassign_$ID
		// /resolve_$ID
		// /my
		// /owner
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
		if err := http.ListenAndServe(":8081", nil); err != nil && err != http.ErrServerClosed {
			log.Println("http server error:", err)
		}
	}()

	b := NewBot()

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("всё работает корректно"))
	})

	for {
		select {
		case <-ctx.Done():
			log.Println("бот завершён из-за отмены контекста")
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
