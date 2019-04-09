package mongo

import (
	"context"
	"encoding/json"
	"github.com/btnguyen2k/consu/reddo"
	"github.com/btnguyen2k/godal"
	"github.com/btnguyen2k/prom"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"reflect"
)

/*
GenericRowMapperMongo is a generic implementation of godal.IRowMapper for MongoDB.

Available since v0.0.2.

Implementation rules:

	- ToRow: transform godal.IGenericBo "as-is" to map[string]interface{}.
	- ToBo: expect input is a JSON data (string or array/slice of bytes), transform it to godal.IGenericBo via JSON unmarshalling.
	- ColumnsList: return []string{"*"} (MongoDB is schema-free, hence column-list is not used).
*/
type GenericRowMapperMongo struct {
}

/*
ToRow implements godal.IRowMapper.ToRow.
This function transforms godal.IGenericBo to map[string]interface{}. Field names are kept intact.
*/
func (mapper *GenericRowMapperMongo) ToRow(storageId string, bo godal.IGenericBo) (interface{}, error) {
	if bo == nil {
		return nil, nil
	}
	result := make(map[string]interface{})
	err := bo.GboTransferViaJson(&result)
	return result, err
}

/*
ToBo implements godal.IRowMapper.ToBo.
This function expects input is a JSON data (string or array/slice of bytes), transforms it to godal.IGenericBo via JSON unmarshalling. Field names are kept intact.
*/
func (mapper *GenericRowMapperMongo) ToBo(storageId string, row interface{}) (godal.IGenericBo, error) {
	if row == nil {
		return nil, nil
	}
	v := reflect.ValueOf(row)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		bo := godal.NewGenericBo()
		err := bo.GboFromJson([]byte(v.Interface().(string)))
		return bo, err
	case reflect.Slice, reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// input is []byte
			zero := make([]byte, 0)
			arr, err := reddo.ToSlice(v.Interface(), reflect.TypeOf(zero))
			if err != nil {
				return nil, err
			}
			bo := godal.NewGenericBo()
			err = bo.GboFromJson(arr.([]byte))
			return bo, err
		}
	}
	return nil, errors.Errorf("cannot construct godal.IGenericBo from input %v", row)
}

/*
ColumnsList implements godal.IRowMapper.ColumnsList.
This function returns []string{"*"} since MongoDB is schema-free (hence column-list is not used).
*/
func (mapper *GenericRowMapperMongo) ColumnsList(storageId string) []string {
	return []string{"*"}
}

var (
	/*
		GenericRowMapperMongoInstance is a pre-created instance of GenericRowMapperMongo that is ready to use.
	*/
	GenericRowMapperMongoInstance godal.IRowMapper = &GenericRowMapperMongo{}
)

/*--------------------------------------------------------------------------------*/

/*
NewGenericDaoMongo constructs a new MongoDB implementation of 'godal.IGenericDao'
*/
func NewGenericDaoMongo(mongoConnect *prom.MongoConnect, agdao *godal.AbstractGenericDao) *GenericDaoMongo {
	dao := &GenericDaoMongo{AbstractGenericDao: agdao, mongoConnect: mongoConnect}
	if dao.GetRowMapper() == nil {
		dao.SetRowMapper(GenericRowMapperMongoInstance)
	}
	return dao
}

/*
GenericDaoMongo is MongoDB implementation of godal.IGenericDao.

Function implementations (n = No, y = Yes, i = inherited):

	(n) GdaoCreateFilter(storageId string, bo godal.IGenericBo) interface{}
	(i) GdaoDelete(storageId string, bo godal.IGenericBo) (int, error)
	(y) GdaoDeleteMany(storageId string, filter interface{}) (int, error)
	(y) GdaoFetchOne(storageId string, filter interface{}) (godal.IGenericBo, error)
	(y) GdaoFetchMany(storageId string, filter interface{}, sorting interface{}, startOffset, numItems int) ([]godal.IGenericBo, error)
	(y) GdaoCreate(storageId string, bo godal.IGenericBo) (int, error)
	(y) GdaoUpdate(storageId string, bo godal.IGenericBo) (int, error)
	(y) GdaoSave(storageId string, bo godal.IGenericBo) (int, error)
*/
type GenericDaoMongo struct {
	*godal.AbstractGenericDao
	mongoConnect *prom.MongoConnect
	txMode       bool
}

/*
GetMongoConnect returns the '*prom.MongoConnect' instance attached to this DAO.
*/
func (dao *GenericDaoMongo) GetMongoConnect() *prom.MongoConnect {
	return dao.mongoConnect
}

/*
SetMongoConnect attaches a '*prom.MongoConnect' instance to this DAO.

Available since v0.0.2
*/
func (dao *GenericDaoMongo) SetMongoConnect(mc *prom.MongoConnect) *GenericDaoMongo {
	dao.mongoConnect = mc
	return dao
}

/*
GetTransactionMode returns 'true' if transaction mode is enabled, 'false' otherwise.
*/
func (dao *GenericDaoMongo) GetTransactionMode() bool {
	return dao.txMode
}

/*
SetTransactionMode enables/disables transaction mode.
*/
func (dao *GenericDaoMongo) SetTransactionMode(enabled bool) *GenericDaoMongo {
	dao.txMode = enabled
	return dao
}

/*
GetMongoCollection returns the MongoDB collection object specified by 'collectionName'.
*/
func (dao *GenericDaoMongo) GetMongoCollection(collectionName string, opts ...*options.CollectionOptions) *mongo.Collection {
	return dao.mongoConnect.GetCollection(collectionName, opts...)
}

/*
MongoDeleteMany performs a MongoDB's delete-many command on the specified collection.

	- 'filter': see MongoDB query selector (https://docs.mongodb.com/manual/reference/operator/query/#query-selectors)
*/
func (dao *GenericDaoMongo) MongoDeleteMany(ctx context.Context, collectionName string, filter map[string]interface{}) (*mongo.DeleteResult, error) {
	return dao.GetMongoCollection(collectionName).DeleteMany(ctx, filter)
}

/*
MongoFetchOne performs a MongoDB's find-one command on the specified collection.

	- 'filter': see MongoDB query selector (https://docs.mongodb.com/manual/reference/operator/query/#query-selectors)
*/
func (dao *GenericDaoMongo) MongoFetchOne(ctx context.Context, collectionName string, filter map[string]interface{}) *mongo.SingleResult {
	return dao.GetMongoCollection(collectionName).FindOne(ctx, filter)
}

/*
MongoFetchMany performs a MongoDB's find command on the specified collection.

	- 'filter': see MongoDB query selector (https://docs.mongodb.com/manual/reference/operator/query/#query-selectors)
	- 'sorting': see MongoDB ascending/descending sort (https://docs.mongodb.com/manual/reference/method/cursor.sort/index.html#sort-asc-desc)
*/
func (dao *GenericDaoMongo) MongoFetchMany(ctx context.Context, collectionName string, filter map[string]interface{}, sorting map[string]int, startOffset, numItems int) (*mongo.Cursor, error) {
	opt := &options.FindOptions{}
	if sorting != nil && len(sorting) > 0 {
		opt.SetSort(sorting)
	}
	if numItems > 0 {
		opt.SetLimit(int64(numItems))
	}
	if startOffset > 0 {
		opt.SetSkip(int64(startOffset))
	}
	return dao.GetMongoCollection(collectionName).Find(ctx, filter, opt)
}

/*
MongoInsertOne performs a MongoDB's insert-one command on the specified collection.
*/
func (dao *GenericDaoMongo) MongoInsertOne(ctx context.Context, collectionName string, doc interface{}) (*mongo.InsertOneResult, error) {
	return dao.GetMongoCollection(collectionName).InsertOne(ctx, doc)
}

/*
MongoUpdateOne performs a MongoDB's find-one-and-replace command with 'upsert=false' on the specified collection.

	- 'filter': see MongoDB query selector (https://docs.mongodb.com/manual/reference/operator/query/#query-selectors)
*/
func (dao *GenericDaoMongo) MongoUpdateOne(ctx context.Context, collectionName string, filter map[string]interface{}, doc interface{}) *mongo.SingleResult {
	upsert := false
	opt := options.FindOneAndReplaceOptions{Upsert: &upsert}
	return dao.GetMongoCollection(collectionName).FindOneAndReplace(ctx, filter, doc, &opt)
}

/*
MongoSaveOne performs a MongoDB's find-one-and-replace command with 'upsert=true' on the specified collection.

	- 'filter': see MongoDB query selector (https://docs.mongodb.com/manual/reference/operator/query/#query-selectors)
*/
func (dao *GenericDaoMongo) MongoSaveOne(ctx context.Context, collectionName string, filter map[string]interface{}, doc interface{}) *mongo.SingleResult {
	upsert := true
	opt := options.FindOneAndReplaceOptions{Upsert: &upsert}
	return dao.GetMongoCollection(collectionName).FindOneAndReplace(ctx, filter, doc, &opt)

}

/*----------------------------------------------------------------------*/

func toMap(input interface{}) (map[string]interface{}, error) {
	if input == nil {
		return nil, nil
	}
	v := reflect.ValueOf(input)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		result := make(map[string]interface{})
		err := json.Unmarshal([]byte(v.Interface().(string)), &result)
		return result, err
	case reflect.Array, reflect.Slice:
		t, err := reddo.ToSlice(v.Interface(), reflect.TypeOf(byte(0)))
		if err != nil {
			return nil, err
		}
		result := make(map[string]interface{})
		err = json.Unmarshal(t.([]byte), &result)
		return result, err
	case reflect.Map:
		t := make(map[string]interface{})
		result, err := reddo.ToMap(v.Interface(), reflect.TypeOf(t))
		return result.(map[string]interface{}), err

	}
	return nil, errors.Errorf("cannot convert %v to map[string]interface{}", input)
}

func toSortingMap(input interface{}) (map[string]int, error) {
	if input == nil {
		return nil, nil
	}
	v := reflect.ValueOf(input)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		result := make(map[string]int)
		err := json.Unmarshal([]byte(v.Interface().(string)), &result)
		return result, err
	case reflect.Array, reflect.Slice:
		t, err := reddo.ToSlice(v.Interface(), reflect.TypeOf(byte(0)))
		if err != nil {
			return nil, err
		}
		result := make(map[string]int)
		err = json.Unmarshal(t.([]byte), &result)
		return result, err
	case reflect.Map:
		t := make(map[string]int)
		result, err := reddo.ToMap(v.Interface(), reflect.TypeOf(t))
		return result.(map[string]int), err

	}
	return nil, errors.Errorf("cannot convert %v to map[string]int", input)
}

/*
GdaoDeleteMany implements godal.IGenericDao.GdaoDeleteMany.

	- 'filter' should be a map[string]interface{}
	- 'filter' can be a string or []byte representing map[string]interface{} in JSON, then it is unmarshalled to map[string]interface{}
	- see MongoDB query selector (https://docs.mongodb.com/manual/reference/operator/query/#query-selectors)
*/
func (dao *GenericDaoMongo) GdaoDeleteMany(storageId string, filter interface{}) (int, error) {
	f, err := toMap(filter)
	if err != nil {
		return 0, err
	}
	ctx, _ := dao.mongoConnect.NewBackgroundContext()
	dbResult, err := dao.MongoDeleteMany(ctx, storageId, f)
	if err != nil {
		return 0, err
	}
	return int(dbResult.DeletedCount), nil
}

/*
GdaoFetchOne implements godal.IGenericDao.GdaoFetchOne.

	- 'filter' should be a map[string]interface{}
	- 'filter' can be a string or []byte representing map[string]interface{} in JSON, then it is unmarshalled to map[string]interface{}
	- see MongoDB query selector (https://docs.mongodb.com/manual/reference/operator/query/#query-selectors)
*/
func (dao *GenericDaoMongo) GdaoFetchOne(storageId string, filter interface{}) (godal.IGenericBo, error) {
	f, err := toMap(filter)
	if err != nil {
		return nil, err
	}
	ctx, _ := dao.mongoConnect.NewBackgroundContext()
	dbResult := dao.MongoFetchOne(ctx, storageId, f)
	jsData, err := dao.mongoConnect.DecodeSingleResultRaw(dbResult)
	if err != nil || jsData == nil {
		return nil, err
	}
	return dao.GetRowMapper().ToBo(storageId, jsData)
}

/*
GdaoFetchMany implements godal.IGenericDao.GdaoFetchMany.

	- 'filter' should be a map[string]interface{}
	- 'filter' can be a string or []byte representing map[string]interface{} in JSON, then it is unmarshalled to map[string]interface{}
	- 'filter' is nil means "match all"
	- see MongoDB query selector (https://docs.mongodb.com/manual/reference/operator/query/#query-selectors)
	- 'sorting' should be a map[string]int
	- 'sorting' can be a string or []byte representing map[string]int in JSON, then it is unmarshalled to map[string]int
	- see MongoDB ascending/descending sort (https://docs.mongodb.com/manual/reference/method/cursor.sort/index.html#sort-asc-desc)
*/
func (dao *GenericDaoMongo) GdaoFetchMany(storageId string, filter interface{}, sorting interface{}, startOffset, numItems int) ([]godal.IGenericBo, error) {
	f, err := toMap(filter)
	if err != nil {
		return nil, err
	}
	s, err := toSortingMap(sorting)
	if err != nil {
		return nil, err
	}
	ctx, _ := dao.mongoConnect.NewBackgroundContext()
	cursor, err := dao.MongoFetchMany(ctx, storageId, f, s, startOffset, numItems)
	if cursor != nil {
		defer func() { _ = cursor.Close(ctx) }()
	}
	if err != nil {
		return nil, err
	}

	var resultBoList []godal.IGenericBo
	var resultError error = nil
	dao.mongoConnect.DecodeResultCallbackRaw(ctx, cursor, func(docNum int, doc []byte, err error) bool {
		if err != nil {
			resultError = err
			return false
		} else {
			bo, e := dao.GetRowMapper().ToBo(storageId, doc)
			if e != nil {
				resultError = e
				return false
			} else {
				resultBoList = append(resultBoList, bo)
			}
		}
		return true
	})
	return resultBoList, resultError
}

func (dao *GenericDaoMongo) insertIfNotExist(ctx context.Context, storageId string, bo godal.IGenericBo) (bool, error) {
	// first fetch existing document from storage
	filter, err := toMap(dao.GdaoCreateFilter(storageId, bo))
	if err != nil {
		return false, err
	}
	row := dao.MongoFetchOne(ctx, storageId, filter)
	jsData, err := dao.mongoConnect.DecodeSingleResultRaw(row)
	if err != nil || jsData != nil {
		// error or document already existed
		return false, err
	}

	// insert new document
	doc, err := dao.GetRowMapper().ToRow(storageId, bo)
	if err != nil {
		return false, err
	}
	_, err = dao.MongoInsertOne(ctx, storageId, doc)
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
GdaoCreate implements godal.IGenericDao.GdaoCreate.
*/
func (dao *GenericDaoMongo) GdaoCreate(storageId string, bo godal.IGenericBo) (int, error) {
	ctx, _ := dao.mongoConnect.NewBackgroundContext()
	if dao.txMode {
		numRows := 0
		err := dao.mongoConnect.GetMongoClient().UseSession(ctx, func(sctx mongo.SessionContext) error {
			err := sctx.StartTransaction(options.Transaction().
				SetReadConcern(readconcern.Snapshot()).
				SetWriteConcern(writeconcern.New(writeconcern.WMajority())))
			if err != nil {
				return err
			}
			result, err := dao.insertIfNotExist(sctx, storageId, bo)
			if err != nil {
				return err
			}
			if result {
				numRows = 1
			}
			return sctx.CommitTransaction(sctx)
		})
		return numRows, err
	} else {
		result, err := dao.insertIfNotExist(ctx, storageId, bo)
		if err != nil {
			return 0, err
		} else if result {
			return 1, nil
		}
		return 0, nil
	}
}

/*
GdaoUpdate implements godal.IGenericDao.GdaoUpdate.
*/
func (dao *GenericDaoMongo) GdaoUpdate(storageId string, bo godal.IGenericBo) (int, error) {
	ctx, _ := dao.mongoConnect.NewBackgroundContext()
	doc, err := dao.GetRowMapper().ToRow(storageId, bo)
	if err != nil {
		return 0, err
	}
	filter, err := toMap(dao.GdaoCreateFilter(storageId, bo))
	if err != nil {
		return 0, err
	}
	result := dao.MongoUpdateOne(ctx, storageId, filter, doc)
	err = result.Err()
	if err != nil {
		return 0, err
	}
	_, err = result.DecodeBytes()
	if err == mongo.ErrNoDocuments {
		return 0, nil
	}
	return 1, nil
}

/*
GdaoSave implements godal.IGenericDao.GdaoSave.
*/
func (dao *GenericDaoMongo) GdaoSave(storageId string, bo godal.IGenericBo) (int, error) {
	ctx, _ := dao.mongoConnect.NewBackgroundContext()
	doc, err := dao.GetRowMapper().ToRow(storageId, bo)
	if err != nil {
		return 0, err
	}
	filter, err := toMap(dao.GdaoCreateFilter(storageId, bo))
	if err != nil {
		return 0, err
	}
	result := dao.MongoSaveOne(ctx, storageId, filter, doc)
	return 1, result.Err()
}