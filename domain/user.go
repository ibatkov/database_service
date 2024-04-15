package domain

import (
	"database/sql"
)

type IUserRepository interface {
	IsAdmin(userId int) bool
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo UserRepository) IsAdmin(userId int) bool {
	row := repo.db.QueryRow(`SELECT access_level FROM users WHERE id = ?`, userId)
	var isAdmin bool
	_ = row.Scan(&isAdmin)
	return isAdmin
}
