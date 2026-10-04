package main

import (
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

	// -------------

	// file, _ := os.Open(FilePath)

	// testCases := []struct {
	// 	Method         string
	// 	URL            string
	// 	ExpectedStatus int
	// 	ExpectedErr    error
	// }{
	// 	{"GET", "/?limit=1&offset=0&order_field=Age&order_by=1", http.StatusOK, nil},
	// }

	// -------------

	// limit := ""
	// offset := ""
	// query := ""
	// orderField := ""
	// orderBy := ""
	// sReq := SearchRequest{
	// 	Limit:      limit,
	// 	Offset:     offset,
	// 	Query:      query,
	// 	OrderField: orderField,
	// 	OrderBy:    orderBy,
	// }

	// sClient := SearchClient{}

	// resp, err := sClient.FindUsers(sReq)

	// -------------

	// req := httptest.NewRequest("GET", "/?limit=1&offset=0&order_field=Age&order_by=1", nil)
	// w := httptest.NewRecorder()

	// SearchServer(w, req)

	// resp := w.Result()
	// // body, _ := io.ReadAll(resp.Body)

	// users := make([]User, 0, 1)

	// _ = json.NewDecoder(resp.Body).Decode(&users)
	// if len(users) != 1 {
	// 	t.Errorf("users count incorrect")
	// }

	// server := httptest.NewServer()
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
