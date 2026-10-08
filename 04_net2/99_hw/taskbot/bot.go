package main

// сюда писать код

import (
	"context"
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

// /tasks
// /new XXX YYY ZZZ - создаёт новую задачу
// /assign_$ID - делает пользователя исполнителем задачи
// /unassign_$ID - снимает задачу с текущего исполнителя
// /resolve_$ID - выполняет задачу, удаляет её из списка
// /my - показывает задачи, которые назначены на меня
// /owner - показывает задачи, которые были созданы мной. Подробности форматирования смотрите в тестах.

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
	mu     sync.Mutex
	nextID int64
}

func NewBot() *Bot {
	return &Bot{
		Tasks:  make([]Task, 0),
		nextID: 1,
	}
}

// func (b *Bot) CreateTask(title string, owner *tgbotapi.User, assigneeID int64, assigneeNickname string) *Task {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
// 	task := Task{
// 		ID:               b.nextID,
// 		Title:            title,
// 		OwnerID:          owner.ID,
// 		OwnerNickname:    owner.UserName,
// 		AssigneeID:       assigneeID,
// 		AssigneeNickname: assigneeNickname,
// 	}
// 	b.Tasks = append(b.Tasks, task)
// 	b.nextID++
// 	tgbotapi.handleNew()
// 	return &task
// }

// func (b *Bot) handlerCreateTask(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	var input struct {
// 		Title            string `json:"title"`
// 		OwnerID          int64  `json:"owner_id"`
// 		OwnerNickname    string `json:"owner_nickname"`
// 		AssigneeID       int64  `json:"assignee_id"`
// 		AssigneeNickname string `json:"assignee_nickname"`
// 	}
// 	err := json.NewDecoder(r.Body).Decode(&input)
// 	if err != nil {
// 		// тут обработка ошибки и возврат http.StatusBadRequest или internalServerError
// 		return
// 	}
// 	task := b.CreateTask(input.Title, input.OwnerID, input.OwnerNickname, input.AssigneeID, input.AssigneeNickname)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	err = json.NewEncoder(w).Encode(&task)
// 	if err != nil {
// 		// тут обработка ошибки и возврат http.StatusBadRequest или internalServerError
// 		return
// 	}
// }

func startTaskBot(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI("")
	if err != nil {
	}

	updates := bot.ListenForWebhook("")

	for update := range updates {
	}
	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
