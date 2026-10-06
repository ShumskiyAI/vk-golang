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

func TestSearchUser(t *testing.T) {
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
			"Success",       // test Name
			5, 0, "", "", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,                      // FindUsers - Err
			[]int{15, 16, 19, 22, 5}, // resp.Users - ID
			true,                     // nextPage
		},
		{
			"Query_match",        // test Name
			5, 0, "on", "Age", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,                    // FindUsers - Err
			[]int{1, 15, 0, 14, 2}, // resp.Users - ID
			true,                   // nextPage
		},
		{
			"Sort_DESC",        // test Name
			6, 9, "", "Id", -1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,                           // FindUsers - Err
			[]int{25, 24, 23, 22, 21, 20}, // resp.Users - ID
			true,                          // nextPage
		},
		{
			"Next_page",        // test Name
			4, 34, "", "Id", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,       // FindUsers - Err
			[]int{34}, // resp.Users - ID
			false,     // nextPage
		},
		{
			"Bad_order_field",   // test Name
			5, 0, "", "test", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("OrderField test invalid"), // FindUsers - Err
			[]int{},                               // resp.Users - ID
			true,                                  // nextPage
		},
		{
			"Negative_limit", // test Name
			-1, 1, "", "", 0, // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("limit must be > 0"), // FindUsers - Err
			[]int{},                         // resp.Users - ID
			true,                            // nextPage
		},
		{
			"Negative_offset", // test Name
			1, -1, "", "", 0,  // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("offset must be > 0"), // FindUsers - Err
			[]int{},                          // resp.Users - ID
			true,                             // nextPage
		},
		{
			"File_not_found", // test Name
			1, 1, "", "", 0,  // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("SearchServer fatal error"), // FindUsers - Err
			[]int{},                                // resp.Users - ID
			true,                                   // nextPage
		},
	}

	for _, tc := range testCases {
		t.Run(tc.nameCase, func(t *testing.T) {
			if tc.nameCase == "File_not_found" {
				FilePath = "qwerty.xml" // проверка на то, что файл не будет найден
			}

			resp, err := sc.FindUsers(SearchRequest{
				Limit:      tc.limit,
				Offset:     tc.offset,
				Query:      tc.query,
				OrderField: tc.orderField,
				OrderBy:    tc.orderBy,
			})
			if tc.ExpectedErr != nil {
				if tc.ExpectedErr.Error() != err.Error() {
					t.Errorf("expected err: %v, got: %v", tc.ExpectedErr, err)
				}
			}
			if tc.ExpectedErr == nil && tc.ExpectedErr != err {
				t.Errorf("expected err: %v, got: %v", tc.ExpectedErr, err)
			}

			if resp == nil {
				t.Logf("resp = nil, err = %v", err)
				return
			}

			if resp.NextPage != tc.expectedNext {
				t.Errorf("expected condition NextPage: %v, got: %v", tc.expectedNext, resp.NextPage)
			}

			if len(resp.Users) != len(tc.expectedIDs) {
				t.Fatalf("expected count results: %d, got: %d", len(tc.expectedIDs), len(resp.Users))
			}

			for i, user := range resp.Users {
				if user.ID != tc.expectedIDs[i] {
					t.Errorf("expected id: %d, got: %d", tc.expectedIDs[i], user.ID)
				}
			}
		})
	}
}

// file, _ := os.Open(FilePath)

func TestFindIds(t *testing.T) {
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
			"Success",       // test Name
			5, 0, "", "", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,                      // FindUsers - Err
			[]int{15, 16, 19, 22, 5}, // resp.Users - ID
			true,                     // nextPage
		},
		{
			"Query_match",        // test Name
			5, 0, "on", "Age", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,                    // FindUsers - Err
			[]int{1, 15, 0, 14, 2}, // resp.Users - ID
			true,                   // nextPage
		},
		{
			"Sort_DESC",        // test Name
			6, 9, "", "Id", -1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,                           // FindUsers - Err
			[]int{25, 24, 23, 22, 21, 20}, // resp.Users - ID
			true,                          // nextPage
		},
		{
			"Next_page",        // test Name
			4, 34, "", "Id", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			nil,       // FindUsers - Err
			[]int{34}, // resp.Users - ID
			false,     // nextPage
		},
		{
			"Bad_order_field",   // test Name
			5, 0, "", "test", 1, // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("OrderField test invalid"), // FindUsers - Err
			[]int{},                               // resp.Users - ID
			true,                                  // nextPage
		},
		{
			"Negative_limit", // test Name
			-1, 1, "", "", 0, // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("limit must be > 0"), // FindUsers - Err
			[]int{},                         // resp.Users - ID
			true,                            // nextPage
		},
		{
			"Negative_offset", // test Name
			1, -1, "", "", 0,  // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("offset must be > 0"), // FindUsers - Err
			[]int{},                          // resp.Users - ID
			true,                             // nextPage
		},
		{
			"File_not_found", // test Name
			1, 1, "", "", 0,  // SearchRequest{limit, offset, query, orderField, orderBy}
			fmt.Errorf("SearchServer fatal error"), // FindUsers - Err
			[]int{},                                // resp.Users - ID
			true,                                   // nextPage
		},
	}
	req := httptest.NewRequest("GET", "/?limit=4&offset=0&order_field=qwe&query=&order_by=1", nil)
	w := httptest.NewRecorder()

	SearchServer(w, req)

	resp := w.Result()
	// body, _ := io.ReadAll(resp.Body)

	users := make([]User, 0, 1)

	_ = json.NewDecoder(resp.Body).Decode(&users)
	// if len(users) != 5 {
	// 	t.Errorf("users count incorrect")
	// }
	for _, user := range users {
		fmt.Println(user.ID)
	}

}

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

// // 2. Прямые тесты сервера — для негативных параметров, которые клиент не пошлёт
// func TestSearchServer_InvalidLimit(t *testing.T) { /* httptest.NewRecorder */ }
// func TestSearchServer_InvalidOffset(t *testing.T) { ... }
// func TestSearchServer_InvalidOrderBy(t *testing.T) { ... }
// func TestSearchServer_OffsetOutOfRange(t *testing.T) { ... }

// // 3. Прямые тесты клиента — без сервера или с фейковым хендлером
// func TestFindUsers_NegativeLimit(t *testing.T) { ... }
// func TestFindUsers_NegativeOffset(t *testing.T) { ... }
// func TestFindUsers_BadJSON(t *testing.T) { /* свой http.HandlerFunc */ }
