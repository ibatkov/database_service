package domain

import (
	"database-service/dbservice/api/logger"
	"database/sql"
)

const AdminUserLevel = "admin"

type IUserRepository interface {
	IsAdmin(userId int) bool
}

type UserRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewUserRepository(db *sql.DB, logger logger.Logger) IUserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (repo UserRepository) IsAdmin(userId int) bool {
	row := repo.db.QueryRow(`SELECT access_level FROM users WHERE id = $1`, userId)
	var accessLevel string
	err := row.Scan(&accessLevel)
	if err != nil {
		repo.logger.Error(err)
	}
	return accessLevel == AdminUserLevel
}

type UserRepositoryStub struct {
	IsAdminStub func(userId int) bool
	RealAdapter IUserRepository
}

func (stub UserRepositoryStub) IsAdmin(userId int) bool {
	if stub.IsAdminStub != nil {
		return stub.IsAdminStub(userId)
	}
	if stub.RealAdapter != nil {
		return stub.RealAdapter.IsAdmin(userId)
	}
	panic("neither IsAdminStub nor RealAdapter are assigned")
}
