package handlers

// toAppError — приватная конвертация error → *domain.AppError (презентация-слой).
// Публичная toAppError в cmd/server — для main.go; здесь локальная, чтобы
// пакет handlers не зависел от main. Логика идентичная: *domain.AppError
// прокидывается, иначе — domain.NewInternal().

import "mytrip/internal/domain"

func toAppError(err error) *domain.AppError {
	if ae, ok := err.(*domain.AppError); ok {
		return ae
	}
	return domain.NewInternal()
}
