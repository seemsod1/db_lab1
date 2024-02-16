package handlers

import "github.com/seemsod1/db_lab1/internal/config"

var Repo *Repository

type Repository struct {
	AppConfig *config.AppConfig
}

func NewRepo(app *config.AppConfig) *Repository {
	return &Repository{AppConfig: app}
}
func NewHandlers(r *Repository) {
	Repo = r
}
