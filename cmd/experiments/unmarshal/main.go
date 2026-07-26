package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Твоя структура
type errResp struct {
	// Здесь int, а не *int. Это значит, что поле НЕ МОЖЕТ быть nil.
	Code    string `json:"code,omitempty"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func main() {
	// 3 ключевых кейса для проверки
	cases := []struct {
		name string
		json string
	}{
		{
			name: "Кейс 1: code приходит как число (400)",
			json: `{"code": 400, "error": "bad_request", "message": "Неверные данные"}`,
		},
		{
			name: "Кейс 2: code приходит явно как null",
			json: `{"code": null, "error": "server_error", "message": "Сбой на сервере"}`,
		},
		{
			name: "Кейс 3: поле code вообще отсутствует в JSON",
			json: `{"error": "unauthorized", "message": "Токен не предоставлен"}`,
		},
	}

	for _, c := range cases {
		fmt.Printf("=== %s ===\n", c.name)
		fmt.Printf("Входной JSON: %s\n", c.json)

		var resp errResp
		err := json.Unmarshal([]byte(c.json), &resp)
		if err != nil {
			fmt.Printf("❌ Ошибка парсинга: %v\n", err)
		} else {
			// Поскольку Code имеет тип int, он никогда не будет nil.
			// Мы просто выводим его значение.
			fmt.Printf("✅ Code:    %d\n", resp.Code)
			fmt.Printf("   Error:   %q\n", resp.Error)
			fmt.Printf("   Message: %q\n", resp.Message)

			if resp.Code == 0 {
				fmt.Println("   ⚠️ Внимание: Code равен 0. Невозможно отличить, пришел ли 'null', поле отсутствовало, или действительно пришел код 0.")
			}
		}
		fmt.Println(strings.Repeat("-", 50))
	}
}
