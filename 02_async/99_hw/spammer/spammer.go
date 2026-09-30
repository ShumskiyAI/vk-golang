package main

import (
	"cmp"
	"fmt"
	"slices"
	"sync"
)

// RunPipeline — связывает всё это каналами и правильно завершает pipeline.
// emails.txt | SelectUsers | SelectMessages | CheckSpam | CombineResults
func RunPipeline(cmds ...cmd) {
	channels := make([]chan interface{}, len(cmds)+1)

	for i := range channels {
		channels[i] = make(chan interface{})
	}

	var wg sync.WaitGroup

	for i := 0; i < len(cmds); i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			defer close(channels[i+1])
			if i == 0 {
				defer close(channels[i])
			}
			cmds[i](channels[i], channels[i+1])
		}(i)
	}

	wg.Wait()
}

// SelectUsers — параллельно вызывает GetUser, но убирает дубликаты;
// in - string
// out - User
func SelectUsers(in, out chan interface{}) {
	var wg sync.WaitGroup
	unique := sync.Map{}

	for data := range in {
		email, ok := data.(string)
		if !ok {
			continue
		}

		wg.Add(1)
		go func(email string) {
			defer wg.Done()

			user := GetUser(email)
			if _, loaded := unique.LoadOrStore(user.Email, struct{}{}); !loaded {
				out <- user
			}
		}(email)
	}

	wg.Wait()
}

// SelectMessages — параллельно работает, но группирует пользователей максимум по 2;
// in - User
// out - MsgID
func SelectMessages(in, out chan interface{}) {
	var wg sync.WaitGroup

	batch := make([]User, 0, GetMessagesMaxUsersBatch)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		users := make([]User, len(batch))
		copy(users, batch)
		batch = batch[:0]

		wg.Add(1)
		go func(users []User) {
			defer wg.Done()
			messagesID, err := GetMessages(users...)
			if err != nil {
				return
			}

			for _, msg := range messagesID {
				out <- msg
			}
		}(users)

	}

	for data := range in {
		user, ok := data.(User)
		if !ok {
			continue
		}

		batch = append(batch, user)
		if len(batch) == GetMessagesMaxUsersBatch {
			flush()
		}
	}

	if len(batch) == 1 {
		flush()
	}

	wg.Wait()
}

// HasSpamMaxAsyncRequests
// in - MsgID
// out - MsgData
// CheckSpam — параллельно, но максимум 5 одновременных HasSpam;

func CheckSpam(in, out chan interface{}) {
	sem := make(chan struct{}, HasSpamMaxAsyncRequests)

	var wg sync.WaitGroup

	for data := range in {
		msgID, ok := data.(MsgID)
		if !ok {
			continue
		}

		sem <- struct{}{}

		wg.Add(1)
		go func(msgID MsgID) {
			defer wg.Done()
			defer func() { <-sem }()

			hasSpam, err := HasSpam(msgID)
			if err != nil {
				return
			}

			out <- MsgData{
				ID:      msgID,
				HasSpam: hasSpam,
			}
		}(msgID)
	}
	wg.Wait()
}

// CombineResults — собирает всё, сортирует и выдаёт результат;
// in - MsgData
// out - string
func CombineResults(in, out chan interface{}) {
	messages := make([]MsgData, 0)

	for data := range in {
		if msg, ok := data.(MsgData); ok {
			messages = append(messages, msg)
		}
	}

	slices.SortFunc(messages, func(a, b MsgData) int {
		boolToInt := func(b bool) int {
			if b {
				return 1
			}
			return 0
		}

		return cmp.Or(
			cmp.Compare(boolToInt(b.HasSpam), boolToInt(a.HasSpam)),
			cmp.Compare(a.ID, b.ID),
		)
	})

	for _, msgData := range messages {
		str := fmt.Sprintf("%t %d", msgData.HasSpam, msgData.ID)
		out <- str
	}
}
