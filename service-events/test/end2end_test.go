package test

import (
	"context"
	"testing"
)

func TestEnd2End(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация инфры для создания тестов (если требуется)

	// Тестирование.
	tests := []end2endTest{
		{
			testName: "Валидный тест.",
			ctx:      ctx,
			err:      nil,
		},
		{
			testName: "Невалидный тест.",
			ctx:      ctx,
			err:      nil,
		},
	}

	for _, test := range tests {
		test.Run(t)
	}
}
