/*
MongoDB Dao example. Run with command:

$ go run examples_bo.go examples_dynamodb.go

AWS DynamoDB Dao implementation guideline:

	- Must implement method godal.IGenericDao.GdaoCreateFilter(storageId string, bo godal.IGenericBo) interface{}
	- If application uses its own BOs instead of godal.IGenericBo, it is recommended to implement a utility method
	  to transform godal.IGenericBo to application's BO and vice versa.
*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/btnguyen2k/consu/reddo"
	"github.com/btnguyen2k/godal"
	gdaodynamod "github.com/btnguyen2k/godal/dynamodb"
	"github.com/btnguyen2k/prom"
	"math/rand"
	"time"
)

type DaoAppDynamodb struct {
	*gdaodynamod.GenericDaoDynamodb
	tableName string
}

func NewDaoAppDynamodb(adc *prom.AwsDynamodbConnect, tableName string) IDaoApp {
	dao := &DaoAppDynamodb{tableName: tableName}
	dao.GenericDaoDynamodb = gdaodynamod.NewGenericDaoDynamodb(adc, godal.NewAbstractGenericDao(dao))
	dao.SetRowMapper(&gdaodynamod.GenericRowMapperDynamodb{ColumnsListMap: map[string][]string{tableName: {"id"}}})
	return dao
}

// toGenericBo transforms BoApp to godal.IGenericBo
func (dao *DaoAppDynamodb) toGenericBo(bo *BoApp) (godal.IGenericBo, error) {
	if bo == nil {
		return nil, nil
	}
	gbo := godal.NewGenericBo()
	js, err := json.Marshal(bo)
	if err != nil {
		return nil, err
	}
	err = gbo.GboFromJson(js)
	return gbo, err
}

// toBoApp transforms godal.IGenericBo to BoApp
func (dao *DaoAppDynamodb) toBoApp(gbo godal.IGenericBo) (*BoApp, error) {
	if gbo == nil {
		return nil, nil
	}
	bo := BoApp{}
	err := gbo.GboTransferViaJson(&bo)
	return &bo, err
}

/*----------------------------------------------------------------------*/
// EnableTxMode implements godal.IGenericDao.EnableTxMode.
func (dao *DaoAppDynamodb) EnableTxMode(txMode bool) {
	// NOTHING
}

// GdaoCreateFilter implements godal.IGenericDao.GdaoCreateFilter.
func (dao *DaoAppDynamodb) GdaoCreateFilter(storageId string, bo godal.IGenericBo) interface{} {
	return map[string]interface{}{"id": bo.GboGetAttrUnsafe("id", reddo.TypeString)}
}

// Delete implements IDaoApp.Delete
func (dao *DaoAppDynamodb) Delete(bo *BoApp) (bool, error) {
	gbo, err := dao.toGenericBo(bo)
	if err != nil {
		return false, err
	}
	numRows, err := dao.GdaoDelete(dao.tableName, gbo)
	return numRows > 0, err
}

// Create implements IDaoApp.Create
func (dao *DaoAppDynamodb) Create(bo *BoApp) (bool, error) {
	gbo, err := dao.toGenericBo(bo)
	if err != nil {
		return false, err
	}
	numRows, err := dao.GdaoCreate(dao.tableName, gbo)
	return numRows > 0, err
}

// Get implements IDaoApp.Get
func (dao *DaoAppDynamodb) Get(id string) (*BoApp, error) {
	filter := map[string]interface{}{"id": id}
	gbo, err := dao.GdaoFetchOne(dao.tableName, filter)
	if err != nil || gbo == nil {
		return nil, err
	}
	return dao.toBoApp(gbo)
}

// GetAll implements IDaoApp.GetAll
func (dao *DaoAppDynamodb) GetAll() ([]*BoApp, error) {
	rows, err := dao.GdaoFetchMany(dao.tableName, nil, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	var result []*BoApp
	for _, e := range rows {
		bo, err := dao.toBoApp(e)
		if err != nil {
			return nil, err
		}
		result = append(result, bo)
	}
	return result, nil
}

// Update implements IDaoApp.Update
func (dao *DaoAppDynamodb) Update(bo *BoApp) (bool, error) {
	gbo, err := dao.toGenericBo(bo)
	if err != nil {
		return false, err
	}
	numRows, err := dao.GdaoUpdate(dao.tableName, gbo)
	return numRows > 0, err
}

// Upsert implements IDaoApp.Upsert
func (dao *DaoAppDynamodb) Upsert(bo *BoApp) (bool, error) {
	gbo, err := dao.toGenericBo(bo)
	if err != nil {
		return false, err
	}
	numRows, err := dao.GdaoSave(dao.tableName, gbo)
	return numRows > 0, err
}

// GetN demonstrates fetching documents with paging
func (dao *DaoAppDynamodb) GetN(startOffset, numRows int) ([]*BoApp, error) {
	rows, err := dao.GdaoFetchMany(dao.tableName, nil, nil, startOffset, numRows)
	if err != nil {
		return nil, err
	}
	var result []*BoApp
	for _, e := range rows {
		bo, err := dao.toBoApp(e)
		if err != nil {
			return nil, err
		}
		result = append(result, bo)
	}
	return result, nil
}

/*----------------------------------------------------------------------*/
func createAwsDynamodbConnect() *prom.AwsDynamodbConnect {
	region := "ap-southeast-1"
	cfg := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewEnvCredentials(),
	}
	adc, err := prom.NewAwsDynamodbConnect(cfg, nil, nil, 10000)
	if err != nil {
		panic(err)
	}
	return adc
}

func initDataDynamodb(adc *prom.AwsDynamodbConnect, tableName string) {
	var schema, key []prom.AwsDynamodbNameAndType

	if ok, err := adc.HasTable(nil, tableName); err != nil {
		panic(err)
	} else if !ok {
		schema = []prom.AwsDynamodbNameAndType{
			{"id", prom.AwsAttrTypeString},
		}
		key = []prom.AwsDynamodbNameAndType{
			{"id", prom.AwsKeyTypePartition},
		}
		if err := adc.CreateTable(nil, tableName, 2, 2, schema, key); err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
		for status, err := adc.GetTableStatus(nil, tableName); status != "ACTIVE" && err == nil; {
			fmt.Printf("    Table [%s] status: %v - %e\n", tableName, status, err)
			time.Sleep(1 * time.Second)
			status, err = adc.GetTableStatus(nil, tableName)
		}
	}

	// delete all items
	pkAttrs := []string{"id"}
	adc.ScanItemsWithCallback(nil, tableName, nil, prom.AwsDynamodbNoIndex, nil, func(item prom.AwsDynamodbItem, lastEvaluatedKey map[string]*dynamodb.AttributeValue) (b bool, e error) {
		keyFilter := make(map[string]interface{})
		for _, v := range pkAttrs {
			keyFilter[v] = item[v]
		}
		_, err := adc.DeleteItem(nil, tableName, keyFilter, nil)
		if err != nil {
			fmt.Printf("    Delete item from table [%s] with key %s: %e\n", tableName, keyFilter, err)
		}
		// fmt.Printf("    Delete item from table [%s] with key %s: %e\n", table, keyFilter, err)
		return true, nil
	})
}

func demoDynamodbInsertItems(loc *time.Location, tableName string) {
	adc := createAwsDynamodbConnect()
	initDataDynamodb(adc, tableName)
	dao := NewDaoAppDynamodb(adc, tableName)

	fmt.Printf("-== Insert items to table ==-\n")

	// insert a document
	t := time.Unix(int64(rand.Int31()), rand.Int63()%1000000000).In(loc)
	bo := BoApp{
		Id:            "log",
		Description:   t.String(),
		ValBool:       rand.Int31()%2 == 0,
		ValInt:        rand.Int(),
		ValFloat:      rand.Float64(),
		ValString:     fmt.Sprintf("Logging application"),
		ValTime:       t,
		ValTimeZ:      t,
		ValDate:       t,
		ValDateZ:      t,
		ValDatetime:   t,
		ValDatetimeZ:  t,
		ValTimestamp:  t,
		ValTimestampZ: t,
		ValList:       []interface{}{true, 0, "1", 2.3, "system", "utility"},
		ValMap:        map[string]interface{}{"tags": []string{"system", "utility"}, "age": 103, "active": true},
	}
	fmt.Println("\tCreating bo:", string(bo.toJson()))
	result, err := dao.Create(&bo)
	if err != nil {
		fmt.Printf("\t\tError: %s\n", err)
	} else {
		fmt.Printf("\t\tResult: %v\n", result)
	}

	// insert another document
	t = time.Unix(int64(rand.Int31()), rand.Int63()%1000000000).In(loc)
	bo = BoApp{
		Id:            "login",
		Description:   t.String(),
		ValBool:       rand.Int31()%2 == 0,
		ValInt:        rand.Int(),
		ValFloat:      rand.Float64(),
		ValString:     fmt.Sprintf("Authentication application"),
		ValTime:       t,
		ValTimeZ:      t,
		ValDate:       t,
		ValDateZ:      t,
		ValDatetime:   t,
		ValDatetimeZ:  t,
		ValTimestamp:  t,
		ValTimestampZ: t,
		ValList:       []interface{}{false, 9.8, "7", 6, "system", "security"},
		ValMap:        map[string]interface{}{"tags": []string{"system", "security"}, "age": 81, "active": false},
	}
	fmt.Println("\tCreating bo:", string(bo.toJson()))
	result, err = dao.Create(&bo)
	if err != nil {
		fmt.Printf("\t\tError: %s\n", err)
	} else {
		fmt.Printf("\t\tResult: %v\n", result)
	}

	// insert another document with duplicated id
	bo = BoApp{Id: "login", ValString: "Authentication application (again)", ValList: []interface{}{"duplicated"}}
	fmt.Println("\tCreating bo:", string(bo.toJson()))
	result, err = dao.Create(&bo)
	if err != nil {
		fmt.Printf("\t\tError: %s\n", err)
	} else {
		fmt.Printf("\t\tResult: %v\n", result)
	}

	fmt.Println(sep)
}

func demoDynamodbFetchItemById(tableName string, itemIds ...string) {
	adc := createAwsDynamodbConnect()
	dao := NewDaoAppDynamodb(adc, tableName)

	fmt.Printf("-== Fetch items by id ==-\n")
	for _, id := range itemIds {
		bo, err := dao.Get(id)
		if err != nil {
			fmt.Printf("\tError while fetching app [%s]: %s\n", id, err)
		} else if bo != nil {
			printApp(bo)
		} else {
			fmt.Printf("\tApp [%s] does not exist\n", id)
		}
	}

	fmt.Println(sep)
}

func demoDynamodbFetchAllItems(tableName string) {
	adc := createAwsDynamodbConnect()
	dao := NewDaoAppDynamodb(adc, tableName)

	fmt.Println("-== Fetch all items in table ==-")
	boList, err := dao.GetAll()
	if err != nil {
		fmt.Printf("\tError while fetching apps: %s\n", err)
	} else {
		for _, bo := range boList {
			printApp(bo)
		}
	}
	fmt.Println(sep)
}

func demoDynamodbDeleteItems(tableName string, itemsIds ...string) {
	adc := createAwsDynamodbConnect()
	dao := NewDaoAppDynamodb(adc, tableName)

	fmt.Println("-== Delete items from table ==-")
	for _, id := range itemsIds {
		bo, err := dao.Get(id)
		if err != nil {
			fmt.Printf("\tError while fetching app [%s]: %s\n", id, err)
		} else if bo == nil {
			fmt.Printf("\tApp [%s] does not exist, no need to delete\n", id)
		} else {
			fmt.Println("\tDeleting bo:", string(bo.toJson()))
			result, err := dao.Delete(bo)
			if err != nil {
				fmt.Printf("\t\tError: %s\n", err)
			} else {
				fmt.Printf("\t\tResult: %v\n", result)
			}
			app, err := dao.Get(id)
			if err != nil {
				fmt.Printf("\t\tError while fetching app [%s]: %s\n", id, err)
			} else if app != nil {
				fmt.Printf("\t\tApp [%s] info: %v\n", app.Id, string(app.toJson()))
			} else {
				fmt.Printf("\t\tApp [%s] no longer exist\n", id)
				result, err = dao.Delete(bo)
				fmt.Printf("\t\tDeleting app [%s] again: %v / %s\n", id, result, err)
			}
		}

	}
	fmt.Println(sep)
}

func demoDynamodbUpdateItems(loc *time.Location, tableName string, itemIds ...string) {
	adc := createAwsDynamodbConnect()
	dao := NewDaoAppDynamodb(adc, tableName)

	fmt.Println("-== Update items from table ==-")
	for _, id := range itemIds {
		t := time.Unix(int64(rand.Int31()), rand.Int63()%1000000000).In(loc)
		bo, err := dao.Get(id)
		if err != nil {
			fmt.Printf("\tError while fetching app [%s]: %s\n", id, err)
		} else if bo == nil {
			fmt.Printf("\tApp [%s] does not exist\n", id)
			bo = &BoApp{
				Id:            id,
				Description:   t.String(),
				ValString:     "(updated)",
				ValTime:       t,
				ValTimeZ:      t,
				ValDate:       t,
				ValDateZ:      t,
				ValDatetime:   t,
				ValDatetimeZ:  t,
				ValTimestamp:  t,
				ValTimestampZ: t,
			}
		} else {
			fmt.Println("\tExisting bo:", string(bo.toJson()))
			bo.Description = t.String()
			bo.ValString += "(updated)"
			bo.ValTime = t
			bo.ValTimeZ = t
			bo.ValDate = t
			bo.ValDateZ = t
			bo.ValDatetime = t
			bo.ValDatetimeZ = t
			bo.ValTimestamp = t
			bo.ValTimestampZ = t
		}
		fmt.Println("\t\tUpdating bo:", string(bo.toJson()))
		result, err := dao.Update(bo)
		if err != nil {
			fmt.Printf("\t\tError while updating app [%s]: %s\n", id, err)
		} else {
			fmt.Printf("\t\tResult: %v\n", result)
			bo, err = dao.Get(id)
			if err != nil {
				fmt.Printf("\t\tError while fetching app [%s]: %s\n", id, err)
			} else if bo != nil {
				fmt.Printf("\t\tApp [%s] info: %v\n", bo.Id, string(bo.toJson()))
			} else {
				fmt.Printf("\t\tApp [%s] does not exist\n", id)
			}
		}
	}
	fmt.Println(sep)
}

func demoDynamodbUpsertItems(loc *time.Location, tableName string, itemIds ...string) {
	adc := createAwsDynamodbConnect()
	dao := NewDaoAppDynamodb(adc, tableName)

	fmt.Printf("-== Upsert items to table ==-\n")
	for _, id := range itemIds {
		t := time.Unix(int64(rand.Int31()), rand.Int63()%1000000000).In(loc)
		bo, err := dao.Get(id)
		if err != nil {
			fmt.Printf("\tError while fetching app [%s]: %s\n", id, err)
		} else if bo == nil {
			fmt.Printf("\tApp [%s] does not exist\n", id)
			bo = &BoApp{
				Id:            id,
				Description:   t.String(),
				ValString:     fmt.Sprintf("(upsert)"),
				ValTime:       t,
				ValTimeZ:      t,
				ValDate:       t,
				ValDateZ:      t,
				ValDatetime:   t,
				ValDatetimeZ:  t,
				ValTimestamp:  t,
				ValTimestampZ: t,
			}
		} else {
			fmt.Println("\tExisting bo:", string(bo.toJson()))
			bo.Description = t.String()
			bo.ValString += fmt.Sprintf("(upsert)")
			bo.ValTime = t
			bo.ValTimeZ = t
			bo.ValDate = t
			bo.ValDateZ = t
			bo.ValDatetime = t
			bo.ValDatetimeZ = t
			bo.ValTimestamp = t
			bo.ValTimestampZ = t
		}
		fmt.Println("\t\tUpserting bo:", string(bo.toJson()))
		result, err := dao.Upsert(bo)
		if err != nil {
			fmt.Printf("\t\tError while upserting app [%s]: %s\n", id, err)
		} else {
			fmt.Printf("\t\tResult: %v\n", result)
			bo, err = dao.Get(id)
			if err != nil {
				fmt.Printf("\t\tError while fetching app [%s]: %s\n", id, err)
			} else if bo != nil {
				fmt.Printf("\t\tApp [%s] info: %v\n", bo.Id, string(bo.toJson()))
			} else {
				fmt.Printf("\t\tApp [%s] does not exist\n", id)
			}
		}
	}
	fmt.Println(sep)
}

func demoDynamodbSelectSortingAndLimit(loc *time.Location, tableName string) {
	adc := createAwsDynamodbConnect()
	initDataDynamodb(adc, tableName)
	dao := NewDaoAppDynamodb(adc, tableName)

	fmt.Println("-== Fetch items from table with sorting and limit ==-")
	n := 100
	fmt.Printf("\tInserting %d docs...\n", n)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("%03d", i)
		t := time.Unix(int64(rand.Int31()), rand.Int63()%1000000000).In(loc)
		bo := BoApp{
			Id:            id,
			Description:   t.String(),
			ValBool:       rand.Int31()%2 == 0,
			ValInt:        rand.Int(),
			ValFloat:      rand.Float64(),
			ValString:     id + " (sorting and limit)",
			ValTime:       t,
			ValTimeZ:      t,
			ValDate:       t,
			ValDateZ:      t,
			ValDatetime:   t,
			ValDatetimeZ:  t,
			ValTimestamp:  t,
			ValTimestampZ: t,
			ValList:       []interface{}{rand.Int31()%2 == 0, i, id},
			ValMap:        map[string]interface{}{"tags": []interface{}{id, i}},
		}
		_, err := dao.Create(&bo)
		if err != nil {
			panic(err)
		}
	}
	startOffset := rand.Intn(n)
	numRows := rand.Intn(10) + 1
	fmt.Printf("\tFetching %d docs, starting from offset %d...\n", numRows, startOffset)
	boList, err := dao.GetN(startOffset, numRows)
	if err != nil {
		fmt.Printf("\t\tError while fetching apps: %s\n", err)
	} else {
		for _, bo := range boList {
			fmt.Printf("\t\tApp [%s] info: %v\n", bo.Id, string(bo.toJson()))
		}
	}
	fmt.Println(sep)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	loc, _ := time.LoadLocation(timeZone)
	fmt.Println("Timezone:", loc)

	tableName := "test_apps"
	demoDynamodbInsertItems(loc, tableName)
	demoDynamodbFetchItemById(tableName, "login", "loggin")
	demoDynamodbFetchAllItems(tableName)
	demoDynamodbDeleteItems(tableName, "login", "loggin")
	demoDynamodbUpdateItems(loc, tableName, "log", "logging")
	demoDynamodbUpsertItems(loc, tableName, "log", "logging")
	demoDynamodbSelectSortingAndLimit(loc, tableName)
}
