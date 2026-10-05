package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// SearchServer (server.go) вы будете запускать через тестовый сервер (httptest.NewServer, пример использования в 3/4_http_testing/server_test.go)

// Покрыть тестами метод FindUsers, чтобы покрытие было 90% или более. Тесты писать в coverage_test.go.

// Тесты так же должны обеспечить 90%-е покрытие SearchServer. Там придётся подменять в некоторых случаях путь до файла (имя файла можно сделать глобальной переменной), чтобы ошибку получить, или же сам файл.

// Тесты писать полноценное, т.е. они реально должны проверять что другая сторона вернула корректный ответ, а не просто покрытие обеспечилось.
// Это значит, что вы должны реально искать по файлу, реально возвращать результаты, а в тесте смотреть что вернулось то, что вы забили в тест. В сравниваемых тестовых данных жестко указать записи не считается хардкодом.

func TestSearchServerAndSearchUser(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer testServer.Close()

	sc := SearchClient{
		URL:         testServer.URL,
		AccessToken: "x",
	}

	testCases := []struct {
		nameCase     string
		limit        int
		offset       int
		query        string
		orderField   string
		orderBy      int
		ExpectedErr  error
		expectedIDs  []int // ← какие ID ожидаем в resp.Users (в каком порядке)
		expectedNext bool  // ← каким должен быть resp.NextPage
	}{
		{
			"first user sort Name", // test Name
			5, 0, "", "", 1,        // SearchRequest
			nil,                      // FindUsers - Err
			[]int{15, 16, 19, 22, 5}, // resp.Users - ID
			true,                     // nextPage
		},
		{
			"first user sort Name", // test Name
			5, 0, "on", "Age", 1,   // SearchRequest
			nil,                    // FindUsers - Err
			[]int{1, 15, 0, 14, 2}, // resp.Users - ID
			true,                   // nextPage
		},
		// тут ещё будут тестовые кейсы
	}

	for _, tc := range testCases {
		t.Run(tc.nameCase, func(t *testing.T) {
			resp, err := sc.FindUsers(SearchRequest{
				Limit:      tc.limit,
				Offset:     tc.offset,
				Query:      tc.query,
				OrderField: tc.orderField,
				OrderBy:    tc.orderBy,
			})

			if tc.ExpectedErr != err {
				t.Errorf("ожидаемая ошибка: %v, полученная: %v", tc.ExpectedErr, err)
			}

			if len(resp.Users) != len(tc.expectedIDs) {
				t.Errorf("ожидаемое количество результатов: %d, полученное: %d", len(tc.expectedIDs), len(resp.Users))
			}
			if resp.NextPage != tc.expectedNext {
				t.Errorf("ожидаемое состояние NextPage: %v, полученное: %v", tc.expectedNext, resp.NextPage)
			}
			for i, user := range resp.Users {
				if user.ID != tc.expectedIDs[i] {
					t.Errorf("ожидаемый id: %d, полученный: %d", tc.expectedIDs[i], user.ID)
				}
			}

		})
	}

}

// -------------

// file, _ := os.Open(FilePath)

// -------------

// sClient := SearchClient{}

// resp, err := sClient.FindUsers(sReq)

// -------------
func TestFindIds(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=5&offset=0&order_field=Age&query=on&order_by=1", nil)
	w := httptest.NewRecorder()

	SearchServer(w, req)

	resp := w.Result()
	// body, _ := io.ReadAll(resp.Body)

	users := make([]User, 0, 1)

	_ = json.NewDecoder(resp.Body).Decode(&users)
	if len(users) != 5 {
		t.Errorf("users count incorrect")
	}
	for _, user := range users {
		fmt.Println(user.ID)
	}

}

// server := httptest.NewServer()

// ------------
//
//

// func nameTest(t *testing.T) {
// 	testCases := []struct {
// 		ID   int
// 		Name string
// 	}{
// 		{1, ""},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(tc.Name, func(t *testing.T) {
// 		})
// 	}
// }

// // 1. Сквозные тесты — покрывают и клиента, и сервера разом
// func TestFindUsers_Success(t *testing.T) { /* через httptest.NewServer(SearchServer) */ }
// func TestFindUsers_QueryMatch(t *testing.T) { ... }
// func TestFindUsers_SortDesc(t *testing.T) { ... }
// func TestFindUsers_Pagination(t *testing.T) { ... }
// func TestFindUsers_BadOrderField(t *testing.T) { ... }
// func TestFindUsers_FileNotFound(t *testing.T) { /* подменяем filePath */ }

// // 2. Прямые тесты сервера — для негативных параметров, которые клиент не пошлёт
// func TestSearchServer_InvalidLimit(t *testing.T) { /* httptest.NewRecorder */ }
// func TestSearchServer_InvalidOffset(t *testing.T) { ... }
// func TestSearchServer_InvalidOrderBy(t *testing.T) { ... }
// func TestSearchServer_OffsetOutOfRange(t *testing.T) { ... }

// // 3. Прямые тесты клиента — без сервера или с фейковым хендлером
// func TestFindUsers_NegativeLimit(t *testing.T) { ... }
// func TestFindUsers_NegativeOffset(t *testing.T) { ... }
// func TestFindUsers_BadJSON(t *testing.T) { /* свой http.HandlerFunc */ }
