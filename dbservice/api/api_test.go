package api

import (
	"context"
	"database-service/auth"
	"database-service/cache"
	"database-service/dbservice/api/controller"
	"database-service/dbservice/api/logger"
	"database-service/dbservice/api/services"
	"database-service/domain"
	"database-service/testcontainers"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	cachepkg "github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

type testUser struct {
	Id          int
	AccessLevel string
}

type testData struct {
	Id     int
	UserId int
	Data   string
}

const TestsKeyPrefix = "cache"
const TestCodePhrase = "code-phrase"

type Suite struct {
	suite.Suite
	// containers
	dbContainer    testcontainers.PostgresContainer
	redisContainer testcontainers.RedisContainer
	// real connections
	db    *sql.DB
	redis *redis.Client
	cache *cachepkg.Cache
	// stubs
	loggerStub         logger.Stub
	cacheStub          cache.Stub
	dbDataRepoStub     domain.DataRepositoryStub
	cachedDataRepoStub domain.DataRepositoryStub
	usersRepoStub      domain.UserRepositoryStub
	authServiceStub    auth.ServiceStub
	authService        *auth.JwtService
	dataServiceStub    services.FakeDataService
	// dynamic data
	token            string
	user             testUser
	data             testData
	afterExecDynamic func()
}

func (suite *Suite) BeforeTest(suiteName, testName string) {
	suite.reload()
}

func TestDataControllerSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (suite *Suite) SetupSuite() {
	os.Setenv("TZ", "UTC")

	suite.dbContainer = testcontainers.SetupDb()

	suite.db, _ = sql.Open("postgres", suite.dbContainer.DSN)

	suite.redisContainer = testcontainers.SetupRedis()

	suite.redis = redis.NewClient(&redis.Options{
		Addr:       suite.redisContainer.Conn(),
		DB:         0,
		MaxRetries: 3,
		PoolSize:   2,
	})

	suite.cache = cachepkg.New(&cachepkg.Options{
		Redis: suite.redis,
	})

	suite.authService = auth.NewJwtService(TestCodePhrase)
	suite.authService.TokenTTL = 10 * time.Minute
}

func (suite *Suite) reload() {

	suite.afterExecDynamic = func() {}

	suite.loggerStub = logger.Stub{
		InfoStub:  func(args ...interface{}) {},
		ErrorStub: func(args ...interface{}) {},
	}

	suite.cacheStub = cache.Stub{
		RealAdapter: cache.NewCacheAdapter(10*time.Second, suite.cache),
	}

	suite.dbDataRepoStub = domain.DataRepositoryStub{
		GetDataByUserStub:  nil,
		GetDataByAdminStub: nil,
		RealAdapter:        domain.NewDbDataRepository(suite.db),
	}

	suite.cachedDataRepoStub = domain.DataRepositoryStub{
		GetDataByUserStub:  nil,
		GetDataByAdminStub: nil,
	}

	suite.usersRepoStub = domain.UserRepositoryStub{
		IsAdminStub: nil,
		RealAdapter: domain.NewUserRepository(suite.db),
	}

	suite.authServiceStub = auth.ServiceStub{
		GetClaimsStub: nil,
		RealService:   suite.authService,
	}

	suite.dataServiceStub = services.FakeDataService{
		GetDataByAccessLevelStub: nil,
		Service:                  services.NewDataService(suite.cachedDataRepoStub, suite.usersRepoStub),
	}

	suite.user = testUser{
		Id:          1,
		AccessLevel: "user",
	}

	suite.data = testData{
		Id:     1,
		UserId: 1,
		Data:   "some_data",
	}

}

func (suite *Suite) getToken() string {
	return fmt.Sprintf(`Bearer %s`, suite.token)
}

func (suite *Suite) CreateUserWithData(userId int, data string) {
	_, err := suite.db.Exec(fmt.Sprintf("INSERT INTO users (id, access_level) values (%d, '%s')", userId, "user"))
	if err != nil {
		panic(err)
	}
	_, err = suite.db.Exec(fmt.Sprintf("INSERT INTO data (id, user_id, data) values (%d, %d, '%s')", userId, userId, data))
	if err != nil {
		panic(err)
	}
}

func (suite *Suite) execDynamic() {
	_, err := suite.db.Exec(fmt.Sprintf("INSERT INTO users (id, access_level) values (%d, '%s')", suite.user.Id, suite.user.AccessLevel))
	if err != nil {
		panic(err)
	}

	_, err = suite.db.Exec(fmt.Sprintf("INSERT INTO data (id, user_id, data) values (%d, %d, '%s')", suite.data.Id, suite.data.UserId, suite.data.Data))
	if err != nil {
		panic(err)
	}

	suite.afterExecDynamic()
}

func (suite *Suite) build() *controller.DataController {
	suite.execDynamic()

	suite.token, _ = suite.authService.GenerateToken(suite.user.Id)

	suite.cachedDataRepoStub.RealAdapter = domain.NewCachedDataRepository(suite.dbDataRepoStub, &suite.cacheStub, TestsKeyPrefix)

	suite.dataServiceStub.Service = services.NewDataService(suite.cachedDataRepoStub, suite.usersRepoStub)

	return controller.NewDataController(suite.authServiceStub, suite.dataServiceStub, suite.loggerStub)
}

func (suite *Suite) TearDownTest() {
	suite.truncateDb()
	suite.truncateRedis()
}

func (suite *Suite) truncateDb() {
	tx, err := suite.db.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	tx.Exec("DELETE FROM data")
	tx.Exec("DELETE FROM users")

	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}

func (suite *Suite) truncateRedis() {
	_ = suite.redis.FlushAll(context.Background())
}

func (suite *Suite) Test_GetData_Success() {
	response := suite.build().GetData(suite.getToken())

	body, _ := json.Marshal(response.Body)
	assert.Equal(suite.T(), http.StatusOK, response.Status)
	assert.Contains(suite.T(), string(body), `"id":1`)
	assert.Contains(suite.T(), string(body), `"user_id":1`)
	assert.Contains(suite.T(), string(body), `"data":`)
}

func (suite *Suite) Test_GetData_GetClaims_Error() {
	suite.authServiceStub.GetClaimsStub = func(token string) (claims *auth.TokenClaims, err error) {
		return nil, errors.New("cannot get claims")
	}

	response := suite.build().GetData(suite.getToken())

	body, _ := json.Marshal(response.Body)
	assert.Equal(suite.T(), http.StatusUnauthorized, response.Status)
	assert.Contains(suite.T(), string(body), `"cannot get claims"`)
}

func (suite *Suite) Test_GetData_DataService_Error() {
	suite.dataServiceStub.GetDataByAccessLevelStub = func(userId int) ([]domain.Data, error) {
		return nil, errors.New("cannot get data")
	}
	suite.loggerStub.ErrorStub = func(args ...interface{}) {
		assert.Contains(suite.T(), args[0].(error).Error(), "cannot get data")

	}
	response := suite.build().GetData(suite.getToken())

	assert.Equal(suite.T(), http.StatusInternalServerError, response.Status)
}

func (suite *Suite) Test_GetData_ByUser_FromCache() {
	var dataFromCache = []domain.CacheData{
		{
			Id:     1,
			UserId: 1,
			Data:   "message_1",
		},
		{
			Id:     2,
			UserId: 1,
			Data:   "message_2",
		},
	}
	marshal, err := suite.cache.Marshal(&dataFromCache)
	if err != nil {
		return
	}

	suite.redis.Set(context.Background(), "cache_1", marshal, 0)
	response := suite.build().GetData(suite.getToken())

	body, _ := json.Marshal(response.Body)
	assert.Equal(suite.T(), http.StatusOK, response.Status)

	assert.Contains(suite.T(), string(body), `"id":1`)
	assert.Contains(suite.T(), string(body), `"user_id":1`)
	assert.Contains(suite.T(), string(body), `"data":"message_1"`)
	assert.Contains(suite.T(), string(body), `"id":2`)
	assert.Contains(suite.T(), string(body), `"user_id":1`)
	assert.Contains(suite.T(), string(body), `"data":"message_2"`)
}

// Cache error suppressed, so we get data from db
func (suite *Suite) Test_GetData_ByUser_FromCache_Error() {
	suite.cacheStub.GetStub = func(ctx context.Context, key string, dest interface{}) error {
		return errors.New("cannot get data")
	}
	response := suite.build().GetData(suite.getToken())

	body, _ := json.Marshal(response.Body)
	assert.Equal(suite.T(), http.StatusOK, response.Status)
	assert.Contains(suite.T(), string(body), `"id":1`)
	assert.Contains(suite.T(), string(body), `"user_id":1`)
	assert.Contains(suite.T(), string(body), `"data":`)
}

func (suite *Suite) Test_GetData_ByUser_FromDb() {
	suite.cacheStub.GetStub = func(ctx context.Context, key string, dest interface{}) error {
		return errors.New("cannot get data")
	}
	response := suite.build().GetData(suite.getToken())

	body, _ := json.Marshal(response.Body)
	assert.Equal(suite.T(), http.StatusOK, response.Status)
	assert.Contains(suite.T(), string(body), `"id":1`)
	assert.Contains(suite.T(), string(body), `"user_id":1`)
	assert.Contains(suite.T(), string(body), `"data":`)
}

func (suite *Suite) Test_GetData_ByUser_FromDb_SetToCache() {

	response := suite.build().GetData(suite.getToken())

	body, _ := json.Marshal(response.Body)
	assert.Equal(suite.T(), http.StatusOK, response.Status)

	r := suite.redis.Get(context.Background(), "cache_1")
	b, _ := r.Bytes()
	var cachedData []domain.CacheData
	_ = suite.cache.Unmarshal(b, &cachedData)

	assert.Contains(suite.T(), string(body), fmt.Sprintf(`"id":%d`, cachedData[0].Id))
	assert.Contains(suite.T(), string(body), fmt.Sprintf(`"user_id":%d`, cachedData[0].UserId))
	assert.Contains(suite.T(), string(body), fmt.Sprintf(`"data":"%s"`, cachedData[0].Data))
}

func (suite *Suite) Test_GetData_ByAdmin_FromDb_SetToCache() {

	suite.usersRepoStub.IsAdminStub = func(userId int) bool {
		return true
	}

	suite.afterExecDynamic = func() {
		suite.CreateUserWithData(2, "message_2")
		suite.CreateUserWithData(3, "message_3")
	}

	response := suite.build().GetData(suite.getToken())

	body, _ := json.Marshal(response.Body)
	assert.Equal(suite.T(), http.StatusOK, response.Status)

	for i := 1; i <= 3; i++ {
		r := suite.redis.Get(context.Background(), "cache_"+strconv.Itoa(i))
		b, _ := r.Bytes()
		var cachedData []domain.CacheData
		_ = suite.cache.Unmarshal(b, &cachedData)

		assert.Contains(suite.T(), string(body), fmt.Sprintf(`"id":%d`, cachedData[0].Id))
		assert.Contains(suite.T(), string(body), fmt.Sprintf(`"user_id":%d`, cachedData[0].UserId))
		assert.Contains(suite.T(), string(body), fmt.Sprintf(`"data":"%s"`, cachedData[0].Data))
	}
}
