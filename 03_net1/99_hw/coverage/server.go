package main

import (
	"cmp"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

var FilePath = "dataset.xml"

type XMLUser struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

// SearchServer - своего рода внешняя система. Непосредственно занимается поиском данных в файле dataset.xml.
func SearchServer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")
	if token == "" {
		http.Error(w, "user is unauthorized", http.StatusUnauthorized)
		return
	}

	urlQuery := r.URL.Query()

	// OrderField string - работает по полям Id, Age, Name, если пустой - то возвращаем по Name, если что-то другое - SearchServer ругается ошибкой. Name - это first_name + last_name из xml.
	orderField := urlQuery.Get("order_field")

	// 	OrderBy int - задает направление сортировки (по полю переданному в order_field) или ее отсутствие (OrderByAsIs). 1 по возрастанию, 0 как встретилось, -1 по убыванию
	orderBy, err := strconv.Atoi(urlQuery.Get("order_by"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(SearchErrorResponse{Error: "invalid order_by"})
		if err != nil {
			// log.Println("JSON is incorrect")
			return
		}
		return
	}

	// offset и limit позволяют получать отсортированный список юзеров пачками с индекса offset не более limit штук.
	limit, err := strconv.Atoi(urlQuery.Get("limit"))
	if err != nil || limit < 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(SearchErrorResponse{Error: "invalid limit"})
		if err != nil {
			// log.Println("data for JSON is incorrect")
			return
		}
		return
	}

	// Можно учесть после сортировки
	offset, err := strconv.Atoi(urlQuery.Get("offset"))
	if err != nil || offset < 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(SearchErrorResponse{Error: "invalid offset"})
		if err != nil {
			// log.Println("data for JSON is incorrect")
			return
		}
		return
	}

	// Query string - подстрока в 1 из полей. Ищет по полям Name и About. Если query пустой, то делаем только сортировку, т.е. возвращаем все записи
	query := urlQuery.Get("query")

	compares := map[string]func(a, b User) int{
		"":     func(a, b User) int { return cmp.Compare(a.Name, b.Name) },
		"Name": func(a, b User) int { return cmp.Compare(a.Name, b.Name) },
		"Id":   func(a, b User) int { return cmp.Compare(a.ID, b.ID) },
		"Age":  func(a, b User) int { return cmp.Compare(a.Age, b.Age) },
	}

	cmpFunc, ok := compares[orderField]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(SearchErrorResponse{Error: ErrorBadOrderField})
		if err != nil {
			// log.Println("data for JSON is incorrect")
			return
		}
		return
	}

	users, err := takeUsersFromXML()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// -------------- QUERY --------------

	queryes := make([]User, 0, len(users)/2)
	if query != "" {
		users = queryesSwap(queryes, users, query)
	}

	// -------------- Сортировка --------------

	// нужно ли чтобы orderBy был в диапазоне от -1 до 1 ?
	switch orderBy {
	case OrderByDesc:
		inner := cmpFunc
		cmpFunc = func(a, b User) int { return -inner(a, b) }
		slices.SortStableFunc(users, cmpFunc)

	case OrderByAsc:
		slices.SortStableFunc(users, cmpFunc)

	case OrderByAsIs:

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(SearchErrorResponse{Error: "invalid order_by"})
		if err != nil {
			// log.Println("data for JSON is incorrect")
			return
		}
		return
	}

	// -------------- Отрезаем --------------

	end := offset + limit

	switch {
	case offset > len(users):
		users = make([]User, 0)

	case end <= len(users):
		users = users[offset:end]

	default:
		users = users[offset:]
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&users)
	if err != nil {
		// log.Println("data for JSON is incorrect")
		return
	}
}

func takeUsersFromXML() ([]User, error) {
	var root struct {
		XMLUsers []XMLUser `xml:"row"`
	}
	file, err := os.Open(FilePath)
	if err != nil {
		return nil, fmt.Errorf("err open file")
	}
	defer file.Close()

	err = xml.NewDecoder(file).Decode(&root)
	if err != nil {
		return nil, fmt.Errorf("err parse file")
	}

	users := make([]User, len(root.XMLUsers))

	for i, xmlUser := range root.XMLUsers {
		users[i] = User{
			ID:     xmlUser.ID,
			Name:   xmlUser.FirstName + " " + xmlUser.LastName,
			Age:    xmlUser.Age,
			About:  xmlUser.About,
			Gender: xmlUser.Gender,
		}
	}
	return users, nil
}

func queryesSwap(queryes, users []User, query string) []User {
	for _, user := range users {
		if strings.Contains(user.Name, query) || strings.Contains(user.About, query) {
			queryes = append(queryes, user)
		}
	}
	return queryes
}
