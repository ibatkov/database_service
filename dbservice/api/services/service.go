package services

import (
	"database-service/domain"
)

type Data interface {
	GetDataByAccessLevel(userId int) ([]domain.Data, error)
}

type DataService struct {
	dataRepo  domain.IDataRepository
	usersRepo domain.IUserRepository
}

func NewDataService(dataRepo domain.IDataRepository, usersRepo domain.IUserRepository) Data {
	return &DataService{
		dataRepo:  dataRepo,
		usersRepo: usersRepo,
	}
}

func (service *DataService) GetDataByAccessLevel(userId int) ([]domain.Data, error) {
	isAdmin := service.usersRepo.IsAdmin(userId)
	if isAdmin {
		return service.dataRepo.GetDataByAdmin()
	}
	return service.dataRepo.GetDataByUser(userId)
}

type FakeDataService struct {
	GetDataByAccessLevelStub func(userId int) ([]domain.Data, error)
	Service                  Data
}

func (service FakeDataService) GetDataByAccessLevel(userId int) ([]domain.Data, error) {
	if service.GetDataByAccessLevelStub != nil {
		return service.GetDataByAccessLevelStub(userId)
	}
	if service.Service != nil {
		return service.Service.GetDataByAccessLevel(userId)
	}
	panic("GetDataByAccessLevel: no function or service provided")
}
