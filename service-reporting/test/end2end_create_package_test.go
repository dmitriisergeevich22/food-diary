package test

import (
	"context"
	"fmt"
	"testing"
)

type end2endTest struct {
	testName string // Название теста.
	ctx      context.Context
	err      error // Ожидаемая ошибка.
}

func (test end2endTest) Run(t *testing.T) {
	t.Run(test.testName, func(t *testing.T) {
		// Логика тестирования.
		fmt.Printf("Успешное завершение теста %s\n", test.testName)
	})
}
