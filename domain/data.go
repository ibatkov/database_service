package domain

import (
	"context"
	"database-service/cache"
	"database/sql"
	"fmt"
)

const (
	sqlQueryDataByUser  = "SELECT * FROM data WHERE user_id = $1"
	sqlQueryDataByAdmin = "SELECT * FROM data"
)

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

type IDataRepository interface {
	GetDataByUser(userId int) ([]Data, error)
	GetDataByAdmin() ([]Data, error)
}

type DbDataRepository struct {
	db *sql.DB
}

func NewDbDataRepository(db *sql.DB) IDataRepository {
	return &DbDataRepository{db: db}
}

func (repo *DbDataRepository) GetDataByUser(userId int) ([]Data, error) {
	rows, err := repo.db.Query(sqlQueryDataByUser, userId)
	if err != nil {
		return nil, err
	}
	return repo.dataFromRows(rows)
}

func (repo *DbDataRepository) GetDataByAdmin() ([]Data, error) {
	rows, err := repo.db.Query(sqlQueryDataByAdmin)
	if err != nil {
		return nil, err
	}
	return repo.dataFromRows(rows)
}

func (repo *DbDataRepository) dataFromRows(rows *sql.Rows) ([]Data, error) {
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

type CachedDataRepository struct {
	dbRepo    IDataRepository
	cache     cache.Cache
	keyPrefix string
}

const keyFormat = `%s_%d`

func NewCachedDataRepository(dbRepo IDataRepository, cache cache.Cache, keyPrefix string) IDataRepository {
	return &CachedDataRepository{dbRepo: dbRepo, cache: cache, keyPrefix: keyPrefix}
}

func (repo *CachedDataRepository) GetDataByUser(userId int) ([]Data, error) {
	data, _ := repo.getFromCache(userId)
	if data != nil {
		return data, nil
	}
	data, err := repo.dbRepo.GetDataByUser(userId)
	_ = repo.setToCache(userId, data)
	return data, err
}

func (repo *CachedDataRepository) GetDataByAdmin() ([]Data, error) {
	data, err := repo.dbRepo.GetDataByAdmin()
	if err != nil {
		return nil, err
	}
	repo.updateUsersCache(data)
	return data, err
}

func (repo *CachedDataRepository) updateUsersCache(data []Data) {
	//TODO Можно сделать тут воркеры
	dataByUserId := make(map[int][]Data)
	for _, d := range data {
		if dataByUserId[d.GetUserId()] == nil {
			dataByUserId[d.GetUserId()] = make([]Data, 0)
		}
		dataByUserId[d.GetUserId()] = append(dataByUserId[d.GetUserId()], d)
	}
	userIds := make([]int, len(dataByUserId))
	for userId := range dataByUserId {
		userIds = append(userIds, userId)
	}
	for _, id := range userIds {
		_ = repo.setToCache(id, dataByUserId[id])
	}
}

type CacheData struct {
	Id     int
	UserId int
	Data   string
}

func (c CacheData) GetId() int {
	return c.Id
}

func (c CacheData) GetUserId() int {
	return c.UserId
}

func (c CacheData) GetData() string {
	return c.Data
}

func (repo *CachedDataRepository) getFromCache(userId int) ([]Data, error) {
	var cacheData []CacheData
	err := repo.cache.Get(context.Background(), repo.getKey(userId), &cacheData)
	if err != nil {
		return nil, err
	}
	if cacheData != nil {
		var data = make([]Data, 0)
		for _, cd := range cacheData {
			data = append(data, cd)
		}
		return data, err
	}
	return nil, nil
}

func (repo *CachedDataRepository) setToCache(userId int, data []Data) error {
	cacheData := make([]CacheData, len(data))
	for i, d := range data {
		cacheData[i] = CacheData{
			Id:     d.GetId(),
			UserId: d.GetUserId(),
			Data:   d.GetData(),
		}
	}
	return repo.cache.Set(context.Background(), repo.getKey(userId), cacheData)
}

func (repo *CachedDataRepository) getKey(userId int) string {
	return fmt.Sprintf(keyFormat, repo.keyPrefix, userId)
}
