package domain

import (
	"database/sql"
)

const (
	sqlQueryDataByUser  = "SELECT * FROM data WHERE user_id = $1"
	sqlQueryDataByAdmin = "SELECT * FROM data"
)

type IDataRepository interface {
	GetDataByUser(userId int) ([]Data, error)
	GetDataByAdmin() ([]Data, error)
}

type DataRepository struct {
	db *sql.DB
}

func NewDataRepository(db *sql.DB) *DataRepository {
	return &DataRepository{db: db}
}

type Data interface {
	GetId() int
	GetUserId() int
	GetData() string
}

type DbData struct {
	id     int
	userId int
	data   string
}

func (d DbData) GetId() int {
	return d.id
}

func (d DbData) GetUserId() int {
	return d.userId
}

func (d DbData) GetData() string {
	return d.data
}

func (repo DataRepository) GetDataByUser(userId int) ([]Data, error) {
	rows, err := repo.db.Query(sqlQueryDataByUser, userId)
	if err != nil {
		return nil, err
	}
	return repo.dataFromRows(rows)
}

func (repo DataRepository) GetDataByAdmin() ([]Data, error) {
	rows, err := repo.db.Query(sqlQueryDataByAdmin)
	if err != nil {
		return nil, err
	}
	return repo.dataFromRows(rows)
}

func (repo DataRepository) dataFromRows(rows *sql.Rows) ([]Data, error) {
	defer rows.Close()

	data := make([]Data, 0)

	for rows.Next() {
		var d DbData
		err := rows.Scan(&d.id, &d.userId, &d.data)
		if err != nil {
			return nil, err
		}
		data = append(data, d)
	}

	return data, nil
}
