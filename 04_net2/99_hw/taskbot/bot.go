package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

var (
	BotToken   = "XXX"
	WebhookURL = "https://525f2cb5.ngrok.io"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrTaskNotOnYou = errors.New("task not on you")
)

type Task struct {
	ID               int64
	Title            string
	OwnerID          int64
	OwnerNickname    string
	AssigneeID       int64
	AssigneeNickname string
}

type Bot struct {
	mu     *sync.RWMutex
	tasks  []Task
	nextID int64
}

func NewBot() *Bot {
	return &Bot{
		mu:     &sync.RWMutex{},
		tasks:  make([]Task, 0),
		nextID: 1,
	}
}

func (b *Bot) createTask(title string, ownerID int64, ownerNickname string) Task {
	b.mu.Lock()
	defer b.mu.Unlock()

	task := Task{
		ID:            b.nextID,
		Title:         title,
		OwnerID:       ownerID,
		OwnerNickname: ownerNickname,
	}
	b.tasks = append(b.tasks, task)
	b.nextID++

	return task
}

func (b *Bot) listTasks() []Task {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]Task, len(b.tasks))
	copy(out, b.tasks)

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (b *Bot) assign(taskID, assigneeID int64, assigneeNickname string) (Task, Task, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.tasks {
		if b.tasks[i].ID == taskID {
			old := b.tasks[i]
			b.tasks[i].AssigneeID = assigneeID
			b.tasks[i].AssigneeNickname = assigneeNickname
			return old, b.tasks[i], nil
		}
	}
	return Task{}, Task{}, ErrTaskNotFound
}

func (b *Bot) unassign(taskID, myID int64) (Task, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.tasks {
		if b.tasks[i].ID == taskID {
			if b.tasks[i].AssigneeID != myID {
				return Task{}, ErrTaskNotOnYou
			}
			b.tasks[i].AssigneeID = 0
			b.tasks[i].AssigneeNickname = ""
			return b.tasks[i], nil
		}
	}
	return Task{}, ErrTaskNotFound
}

func (b *Bot) resolve(taskID int64) (Task, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.tasks {
		if b.tasks[i].ID == taskID {
			removed := b.tasks[i]
			b.tasks = append(b.tasks[:i], b.tasks[i+1:]...)
			return removed, nil
		}
	}
	return Task{}, ErrTaskNotFound
}

func (b *Bot) myTasks(myID int64) []Task {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]Task, 0)
	for _, t := range b.tasks {
		if t.AssigneeID == myID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (b *Bot) ownerTasks(myID int64) []Task {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]Task, 0)
	for _, t := range b.tasks {
		if t.OwnerID == myID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func formatTaskLine(t Task, viewerID int64, withAssigneeLabel bool) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d. %s by @%s", t.ID, t.Title, t.OwnerNickname)

	switch {
	case t.AssigneeID == 0:

		sb.WriteString("\n/assign_" + strconv.FormatInt(t.ID, 10))
	case withAssigneeLabel && t.AssigneeID == viewerID:
		sb.WriteString("\nassignee: я")
		sb.WriteString(fmt.Sprintf("\n/unassign_%d /resolve_%d", t.ID, t.ID))
	case withAssigneeLabel:
		sb.WriteString("\nassignee: @" + t.AssigneeNickname)
	default:
		sb.WriteString(fmt.Sprintf("\n/unassign_%d /resolve_%d", t.ID, t.ID))
	}
	return sb.String()
}

func formatTaskList(tasks []Task, viewerID int64) string {
	if len(tasks) == 0 {
		return "Нет задач"
	}
	parts := make([]string, 0, len(tasks))
	for _, t := range tasks {
		parts = append(parts, formatTaskLine(t, viewerID, true))
	}
	return strings.Join(parts, "\n\n")
}

func formatTaskListShort(tasks []Task, viewerID int64) string {
	if len(tasks) == 0 {
		return "Нет задач"
	}
	parts := make([]string, 0, len(tasks))
	for _, t := range tasks {
		parts = append(parts, formatTaskLine(t, viewerID, false))
	}
	return strings.Join(parts, "\n\n")
}

func (b *Bot) handleMessage(bot *tgbotapi.BotAPI, upd tgbotapi.Update) {
	msg := upd.Message
	if msg == nil || msg.From == nil {
		return
	}

	user := msg.From
	chatID := msg.Chat.ID
	text := msg.Text

	send := func(chatID int64, text string) {
		if _, err := bot.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
			log.Println("send error:", err)
		}
	}

	switch {
	case text == "/tasks":
		send(chatID, formatTaskList(b.listTasks(), user.ID))

	case text == "/my":
		send(chatID, formatTaskListShort(b.myTasks(user.ID), user.ID))

	case text == "/owner":
		send(chatID, formatTaskListShort(b.ownerTasks(user.ID), user.ID))

	case strings.HasPrefix(text, "/new "):
		title := strings.TrimPrefix(text, "/new ")
		if title == "" {
			send(chatID, "Пустое название задачи")
			return
		}
		task := b.createTask(title, user.ID, user.UserName)
		send(chatID, fmt.Sprintf("Задача %q создана, id=%d", task.Title, task.ID))

	case strings.HasPrefix(text, "/assign_"):
		id, err := parseID(text, "/assign_")
		if err != nil {
			send(chatID, "Неверный id задачи")
			return
		}
		old, updated, err := b.assign(id, user.ID, user.UserName)
		if err != nil {
			send(chatID, "Задача не найдена")
			return
		}
		send(chatID, fmt.Sprintf("Задача %q назначена на вас", updated.Title))

		if old.AssigneeID != 0 && old.AssigneeID != user.ID {
			send(old.AssigneeID, fmt.Sprintf(
				"Задача %q назначена на @%s", updated.Title, user.UserName))
		} else if old.AssigneeID == 0 && updated.OwnerID != user.ID {
			send(updated.OwnerID, fmt.Sprintf(
				"Задача %q назначена на @%s", updated.Title, user.UserName))
		}

	case strings.HasPrefix(text, "/unassign_"):
		id, err := parseID(text, "/unassign_")
		if err != nil {
			send(chatID, "Неверный id задачи")
			return
		}
		task, err := b.unassign(id, user.ID)
		switch {
		case errors.Is(err, ErrTaskNotOnYou):
			send(chatID, "Задача не на вас")
		case err != nil:
			send(chatID, "Задача не найдена")
		default:
			send(chatID, "Принято")
			if task.OwnerID != user.ID {
				send(task.OwnerID, fmt.Sprintf(
					"Задача %q осталась без исполнителя", task.Title))
			}
		}

	case strings.HasPrefix(text, "/resolve_"):
		id, err := parseID(text, "/resolve_")
		if err != nil {
			send(chatID, "Неверный id задачи")
			return
		}
		task, err := b.resolve(id)
		if err != nil {
			send(chatID, "Задача не найдена")
			return
		}
		send(chatID, fmt.Sprintf("Задача %q выполнена", task.Title))
		if task.OwnerID != user.ID {
			send(task.OwnerID, fmt.Sprintf(
				"Задача %q выполнена @%s", task.Title, user.UserName))
		}
	}
}

func parseID(text, prefix string) (int64, error) {
	return strconv.ParseInt(strings.TrimPrefix(text, prefix), 10, 64)
}

func startTaskBot(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return fmt.Errorf("new bot api: %w", err)
	}
	log.Printf("authorized as @%s", bot.Self.UserName)

	hw, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		return fmt.Errorf("set webhook: %w", err)
	}

	if _, err := bot.Request(hw); err != nil {
		return fmt.Errorf("error request: %w", err)
	}

	u, err := url.Parse(WebhookURL)
	if err != nil {
		return fmt.Errorf("parse webhook url: %w", err)
	}
	addr := u.Host
	if u.Port() == "" {
		if u.Scheme == "https" {
			addr += ":443"
		} else {
			addr += ":80"
		}
	}

	updates := bot.ListenForWebhook("/")

	srv := &http.Server{Addr: addr}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("http server error:", err)
		}
	}()

	b := NewBot()

	for {
		select {
		case <-ctx.Done():
			_ = srv.Close()
			return nil
		case upd, ok := <-updates:
			if !ok {
				_ = srv.Close()
				return nil
			}
			b.handleMessage(bot, upd)
		}
	}
}

func main() {
	if err := startTaskBot(context.Background()); err != nil {
		log.Fatal(err)
	}
}
