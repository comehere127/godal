package sql

import (
	"encoding/json"
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"github.com/btnguyen2k/godal"
	"testing"
)

func TestGenericRowMapperSql_ColumnsList(t *testing.T) {
	name := "TestGenericRowMapperSql_ColumnsList"
	rm := &GenericRowMapperSql{ColumnsListMap: map[string][]string{"test": {"col0", "col1", "col2"}}}

	colsTest := rm.ColumnsList("test")
	if colsTest == nil || len(colsTest) != 3 || colsTest[0] != "col0" || colsTest[1] != "col1" || colsTest[2] != "col2" {
		t.Fatalf("%s failed. StorageId: %s / Column list: %#v", name, "test", colsTest)
	}

	cols := rm.ColumnsList("dummy")
	if cols == nil || len(cols) != 1 || cols[0] != "*" {
		t.Fatalf("%s failed. StorageId: %s / Column list: %#v", name, "test", cols)
	}
}

func testToBo(t *testing.T, name string, rowmapper godal.IRowMapper, table string, row interface{}) {
	colA, colB, colC, col1, col2 := "cola", "ColB", "colC", "Col1", "coL2"
	valA, valB, val1, val2 := "a", "B", int64(1), int64(2)

	bo, err := rowmapper.ToBo(table, row)
	if err != nil || bo == nil {
		t.Fatalf("%s failed: %e / %v", name, err, bo)
	}
	if bo.GboGetAttrUnsafe(colA, reddo.TypeString) != valA ||
		bo.GboGetAttrUnsafe(colB, reddo.TypeString) != valB ||
		bo.GboGetAttrUnsafe(colC, reddo.TypeString) != nil ||
		bo.GboGetAttrUnsafe(col1, reddo.TypeInt).(int64) != val1 ||
		bo.GboGetAttrUnsafe(col2, reddo.TypeInt).(int64) != val2 {
		t.Fatalf("%s failed, Row: %v - Bo: %v", name, row, bo)
	}
}

func TestGenericRowMapperSql_ToBo(t *testing.T) {
	name := "TestGenericRowMapperSql_ToBo"
	table := "table"
	colA, colB, col1, col2 := "cola", "ColB", "Col1", "coL2"
	valA, valB, val1, val2 := "a", "B", int64(1), int64(2)
	rowmapper := &GenericRowMapperSql{}

	{
		row := map[string]interface{}{colA: valA, colB: valB, col1: val1, col2: val2}
		testToBo(t, name, rowmapper, table, row)
		testToBo(t, name, rowmapper, table, &row)
		testToBo(t, name, rowmapper, table+"-not-exists", row)
		row2 := &row
		testToBo(t, name, rowmapper, table, &row2)
	}

	{
		row := fmt.Sprintf(`{"%s": "%v", "%s": "%v", "%s": %v, "%s": %v}`, colA, valA, colB, valB, col1, val1, col2, val2)
		testToBo(t, name, rowmapper, table, row)
		testToBo(t, name, rowmapper, table, &row)
		testToBo(t, name, rowmapper, table+"-not-exists", row)
		row2 := &row
		testToBo(t, name, rowmapper, table, &row2)
	}

	{
		row := []byte(fmt.Sprintf(`{"%s": "%v", "%s": "%v", "%s": %v, "%s": %v}`, colA, valA, colB, valB, col1, val1, col2, val2))
		testToBo(t, name, rowmapper, table, row)
		testToBo(t, name, rowmapper, table, &row)
		testToBo(t, name, rowmapper, table+"-not-exists", row)
		row2 := &row
		testToBo(t, name, rowmapper, table, &row2)
	}

	{
		var row interface{} = nil
		if bo, err := rowmapper.ToBo(table, row); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
		if bo, err := rowmapper.ToBo(table, &row); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
		row2 := &row
		if bo, err := rowmapper.ToBo(table, &row2); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
	}

	{
		var row *string = nil
		if bo, err := rowmapper.ToBo(table, row); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
		if bo, err := rowmapper.ToBo(table, &row); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
		row2 := &row
		if bo, err := rowmapper.ToBo(table, &row2); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
	}

	{
		var row []byte = nil
		if bo, err := rowmapper.ToBo(table, row); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
		if bo, err := rowmapper.ToBo(table, &row); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
		row2 := &row
		if bo, err := rowmapper.ToBo(table, &row2); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
	}

	{
		var row *[]byte = nil
		if bo, err := rowmapper.ToBo(table, row); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
		if bo, err := rowmapper.ToBo(table, &row); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
		row2 := &row
		if bo, err := rowmapper.ToBo(table, &row2); err != nil || bo != nil {
			t.Fatalf("%s failed: %e / %v", name, err, bo)
		}
	}
}

func TestGenericRowMapperSql_ToRow_Intact(t *testing.T) {
	name := "TestGenericRowMapperSql_ToRow_Intact"
	rm := &GenericRowMapperSql{NameTransformation: NameTransfIntact}
	js := `{"ColA":1,"colb":"a string","COLC":2.3,"colD":true}`
	gbo := godal.NewGenericBo()
	gbo.GboFromJson([]byte(js))

	{
		row, err := rm.ToRow("", gbo)
		if err != nil || row == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		m, ok := row.(map[string]interface{})
		if !ok || len(m) != 4 {
			t.Fatalf("%s failed: row: %#v", name, row)
		}
		if v, e := reddo.ToInt(m["ColA"]); e != nil || v != 1 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "ColA", 1, m["ColA"])
		}
		if v, e := reddo.ToString(m["colb"]); e != nil || v != "a string" {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "colb", "a string", m["colb"])
		}
		if v, e := reddo.ToFloat(m["COLC"]); e != nil || v != 2.3 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "COLC", 2.3, m["COLC"])
		}
		if v, e := reddo.ToBool(m["colD"]); e != nil || v != true {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "colD", true, m["colD"])
		}
	}

	rm.GboFieldToColNameTranslator = map[string]map[string]interface{}{"table_name": {"ColA": "a", "colb": "b", "COLC": "c", "colD": "d"}}
	{
		row, err := rm.ToRow("table_name", gbo)
		if err != nil || row == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		m, ok := row.(map[string]interface{})
		if !ok || len(m) != 4 {
			t.Fatalf("%s failed: row: %#v", name, row)
		}
		if v, e := reddo.ToInt(m["a"]); e != nil || v != 1 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "a", 1, m["a"])
		}
		if v, e := reddo.ToString(m["b"]); e != nil || v != "a string" {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "b", "a string", m["b"])
		}
		if v, e := reddo.ToFloat(m["c"]); e != nil || v != 2.3 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "c", 2.3, m["c"])
		}
		if v, e := reddo.ToBool(m["d"]); e != nil || v != true {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "d", true, m["d"])
		}
	}

	rm.GboFieldToColNameTranslator = map[string]map[string]interface{}{"*": {"ColA": "a", "colb": "b", "COLC": "c", "colD": "d"}}
	{
		row, err := rm.ToRow("", gbo)
		if err != nil || row == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		m, ok := row.(map[string]interface{})
		if !ok || len(m) != 4 {
			t.Fatalf("%s failed: row: %#v", name, row)
		}
		if v, e := reddo.ToInt(m["a"]); e != nil || v != 1 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "a", 1, m["a"])
		}
		if v, e := reddo.ToString(m["b"]); e != nil || v != "a string" {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "b", "a string", m["b"])
		}
		if v, e := reddo.ToFloat(m["c"]); e != nil || v != 2.3 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "c", 2.3, m["c"])
		}
		if v, e := reddo.ToBool(m["d"]); e != nil || v != true {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "d", true, m["d"])
		}
	}
}

func TestGenericRowMapperSql_ToRow_LowerCase(t *testing.T) {
	name := "TestGenericRowMapperSql_ToRow_LowerCase"
	rm := &GenericRowMapperSql{NameTransformation: NameTransfLowerCase}
	js := `{"ColA":1,"colb":"a string","COLC":2.3,"colD":true}`
	gbo := godal.NewGenericBo()
	gbo.GboFromJson([]byte(js))

	{
		row, err := rm.ToRow("", gbo)
		if err != nil || row == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		m, ok := row.(map[string]interface{})
		if !ok || len(m) != 4 {
			t.Fatalf("%s failed: row: %#v", name, row)
		}
		if v, e := reddo.ToInt(m["cola"]); e != nil || v != 1 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "cola", 1, m["cola"])
		}
		if v, e := reddo.ToString(m["colb"]); e != nil || v != "a string" {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "colb", "a string", m["colb"])
		}
		if v, e := reddo.ToFloat(m["colc"]); e != nil || v != 2.3 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "colc", 2.3, m["colc"])
		}
		if v, e := reddo.ToBool(m["cold"]); e != nil || v != true {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "cold", true, m["cold"])
		}
	}

	rm.GboFieldToColNameTranslator = map[string]map[string]interface{}{"table_name": {"cola": "a", "colb": "b", "colc": "c", "cold": "d"}}
	{
		row, err := rm.ToRow("table_name", gbo)
		if err != nil || row == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		m, ok := row.(map[string]interface{})
		if !ok || len(m) != 4 {
			t.Fatalf("%s failed: row: %#v", name, row)
		}
		if v, e := reddo.ToInt(m["a"]); e != nil || v != 1 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "a", 1, m["a"])
		}
		if v, e := reddo.ToString(m["b"]); e != nil || v != "a string" {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "b", "a string", m["b"])
		}
		if v, e := reddo.ToFloat(m["c"]); e != nil || v != 2.3 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "c", 2.3, m["c"])
		}
		if v, e := reddo.ToBool(m["d"]); e != nil || v != true {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "d", true, m["d"])
		}
	}

	rm.GboFieldToColNameTranslator = map[string]map[string]interface{}{"*": {"cola": "a", "colb": "b", "colc": "c", "cold": "d"}}
	{
		row, err := rm.ToRow("", gbo)
		if err != nil || row == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		m, ok := row.(map[string]interface{})
		if !ok || len(m) != 4 {
			t.Fatalf("%s failed: row: %#v", name, row)
		}
		if v, e := reddo.ToInt(m["a"]); e != nil || v != 1 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "a", 1, m["a"])
		}
		if v, e := reddo.ToString(m["b"]); e != nil || v != "a string" {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "b", "a string", m["b"])
		}
		if v, e := reddo.ToFloat(m["c"]); e != nil || v != 2.3 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "c", 2.3, m["c"])
		}
		if v, e := reddo.ToBool(m["d"]); e != nil || v != true {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "d", true, m["d"])
		}
	}
}

func TestGenericRowMapperSql_ToRow_UpperCase(t *testing.T) {
	name := "TestGenericRowMapperSql_ToRow_UpperCase"
	rm := &GenericRowMapperSql{NameTransformation: NameTransfUpperCase}
	js := `{"ColA":1,"colb":"a string","COLC":2.3,"colD":true}`
	gbo := godal.NewGenericBo()
	gbo.GboFromJson([]byte(js))

	{
		row, err := rm.ToRow("", gbo)
		if err != nil || row == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		m, ok := row.(map[string]interface{})
		if !ok || len(m) != 4 {
			t.Fatalf("%s failed: row: %#v", name, row)
		}
		if v, e := reddo.ToInt(m["COLA"]); e != nil || v != 1 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "COLA", 1, m["COLA"])
		}
		if v, e := reddo.ToString(m["COLB"]); e != nil || v != "a string" {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "COLB", "a string", m["COLB"])
		}
		if v, e := reddo.ToFloat(m["COLC"]); e != nil || v != 2.3 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "COLC", 2.3, m["COLC"])
		}
		if v, e := reddo.ToBool(m["COLD"]); e != nil || v != true {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "COLD", true, m["COLD"])
		}
	}

	rm.GboFieldToColNameTranslator = map[string]map[string]interface{}{"table_name": {"COLA": "a", "COLB": "b", "COLC": "c", "COLD": "d"}}
	{
		row, err := rm.ToRow("table_name", gbo)
		if err != nil || row == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		m, ok := row.(map[string]interface{})
		if !ok || len(m) != 4 {
			t.Fatalf("%s failed: row: %#v", name, row)
		}
		if v, e := reddo.ToInt(m["a"]); e != nil || v != 1 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "a", 1, m["a"])
		}
		if v, e := reddo.ToString(m["b"]); e != nil || v != "a string" {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "b", "a string", m["b"])
		}
		if v, e := reddo.ToFloat(m["c"]); e != nil || v != 2.3 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "c", 2.3, m["c"])
		}
		if v, e := reddo.ToBool(m["d"]); e != nil || v != true {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "d", true, m["d"])
		}
	}

	rm.GboFieldToColNameTranslator = map[string]map[string]interface{}{"*": {"COLA": "a", "COLB": "b", "COLC": "c", "COLD": "d"}}
	{
		row, err := rm.ToRow("", gbo)
		if err != nil || row == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		m, ok := row.(map[string]interface{})
		if !ok || len(m) != 4 {
			t.Fatalf("%s failed: row: %#v", name, row)
		}
		if v, e := reddo.ToInt(m["a"]); e != nil || v != 1 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "a", 1, m["a"])
		}
		if v, e := reddo.ToString(m["b"]); e != nil || v != "a string" {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "b", "a string", m["b"])
		}
		if v, e := reddo.ToFloat(m["c"]); e != nil || v != 2.3 {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "c", 2.3, m["c"])
		}
		if v, e := reddo.ToBool(m["d"]); e != nil || v != true {
			t.Fatalf("%s failed: expected data[%s] to be %#v but received %#v", name, "d", true, m["d"])
		}
	}
}

func TestGenericRowMapperSql_ToGbo_Intact(t *testing.T) {
	name := "TestGenericRowMapperSql_ToGbo_Intact"
	rm := &GenericRowMapperSql{NameTransformation: NameTransfIntact}
	js := `{"ColA":1,"colb":"a string","COLC":2.3,"colD":true}`
	row := make(map[string]interface{})
	json.Unmarshal([]byte(js), &row)

	{
		gbo, err := rm.ToBo("", row)
		if err != nil || gbo == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		if v, e := gbo.GboGetAttr("ColA", reddo.TypeInt); e != nil || v.(int64) != 1 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "ColA", 1, v)
		}
		if v, e := gbo.GboGetAttr("colb", reddo.TypeString); e != nil || v.(string) != "a string" {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "colb", "a string", v)
		}
		if v, e := gbo.GboGetAttr("COLC", reddo.TypeFloat); e != nil || v.(float64) != 2.3 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "COLC", 2.3, v)
		}
		if v, e := gbo.GboGetAttr("colD", reddo.TypeBool); e != nil || v.(bool) != true {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "colD", true, v)
		}
	}

	rm.ColNameToGboFieldTranslator = map[string]map[string]interface{}{"table_name": {"ColA": "a", "colb": "b", "COLC": "c", "colD": "d"}}
	{
		gbo, err := rm.ToBo("table_name", row)
		if err != nil || gbo == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		if v, e := gbo.GboGetAttr("a", reddo.TypeInt); e != nil || v.(int64) != 1 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "a", 1, v)
		}
		if v, e := gbo.GboGetAttr("b", reddo.TypeString); e != nil || v.(string) != "a string" {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "b", "a string", v)
		}
		if v, e := gbo.GboGetAttr("c", reddo.TypeFloat); e != nil || v.(float64) != 2.3 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "c", 2.3, v)
		}
		if v, e := gbo.GboGetAttr("d", reddo.TypeBool); e != nil || v.(bool) != true {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "d", true, v)
		}
	}

	rm.ColNameToGboFieldTranslator = map[string]map[string]interface{}{"*": {"ColA": "a", "colb": "b", "COLC": "c", "colD": "d"}}
	{
		gbo, err := rm.ToBo("", row)
		if err != nil || gbo == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		if v, e := gbo.GboGetAttr("a", reddo.TypeInt); e != nil || v.(int64) != 1 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "a", 1, v)
		}
		if v, e := gbo.GboGetAttr("b", reddo.TypeString); e != nil || v.(string) != "a string" {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "b", "a string", v)
		}
		if v, e := gbo.GboGetAttr("c", reddo.TypeFloat); e != nil || v.(float64) != 2.3 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "c", 2.3, v)
		}
		if v, e := gbo.GboGetAttr("d", reddo.TypeBool); e != nil || v.(bool) != true {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "d", true, v)
		}
	}
}

func TestGenericRowMapperSql_ToGbo_LowerCase(t *testing.T) {
	name := "TestGenericRowMapperSql_ToGbo_LowerCase"
	rm := &GenericRowMapperSql{NameTransformation: NameTransfLowerCase}
	js := `{"ColA":1,"colb":"a string","COLC":2.3,"colD":true}`
	row := make(map[string]interface{})
	json.Unmarshal([]byte(js), &row)

	{
		gbo, err := rm.ToBo("", row)
		if err != nil || gbo == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		if v, e := gbo.GboGetAttr("cola", reddo.TypeInt); e != nil || v.(int64) != 1 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "cola", 1, v)
		}
		if v, e := gbo.GboGetAttr("colb", reddo.TypeString); e != nil || v.(string) != "a string" {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "colb", "a string", v)
		}
		if v, e := gbo.GboGetAttr("colc", reddo.TypeFloat); e != nil || v.(float64) != 2.3 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "colc", 2.3, v)
		}
		if v, e := gbo.GboGetAttr("cold", reddo.TypeBool); e != nil || v.(bool) != true {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "cold", true, v)
		}
	}

	rm.ColNameToGboFieldTranslator = map[string]map[string]interface{}{"table_name": {"cola": "a", "colb": "b", "colc": "c", "cold": "d"}}
	{
		gbo, err := rm.ToBo("table_name", row)
		if err != nil || gbo == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		if v, e := gbo.GboGetAttr("a", reddo.TypeInt); e != nil || v.(int64) != 1 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "a", 1, v)
		}
		if v, e := gbo.GboGetAttr("b", reddo.TypeString); e != nil || v.(string) != "a string" {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "b", "a string", v)
		}
		if v, e := gbo.GboGetAttr("c", reddo.TypeFloat); e != nil || v.(float64) != 2.3 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "c", 2.3, v)
		}
		if v, e := gbo.GboGetAttr("d", reddo.TypeBool); e != nil || v.(bool) != true {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "d", true, v)
		}
	}

	rm.ColNameToGboFieldTranslator = map[string]map[string]interface{}{"*": {"cola": "a", "colb": "b", "colc": "c", "cold": "d"}}
	{
		gbo, err := rm.ToBo("", row)
		if err != nil || gbo == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		if v, e := gbo.GboGetAttr("a", reddo.TypeInt); e != nil || v.(int64) != 1 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "a", 1, v)
		}
		if v, e := gbo.GboGetAttr("b", reddo.TypeString); e != nil || v.(string) != "a string" {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "b", "a string", v)
		}
		if v, e := gbo.GboGetAttr("c", reddo.TypeFloat); e != nil || v.(float64) != 2.3 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "c", 2.3, v)
		}
		if v, e := gbo.GboGetAttr("d", reddo.TypeBool); e != nil || v.(bool) != true {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "d", true, v)
		}
	}
}

func TestGenericRowMapperSql_ToGbo_UpperCase(t *testing.T) {
	name := "TestGenericRowMapperSql_ToGbo_UpperCase"
	rm := &GenericRowMapperSql{NameTransformation: NameTransfUpperCase}
	js := `{"ColA":1,"colb":"a string","COLC":2.3,"colD":true}`
	row := make(map[string]interface{})
	json.Unmarshal([]byte(js), &row)

	{
		gbo, err := rm.ToBo("", row)
		if err != nil || gbo == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		if v, e := gbo.GboGetAttr("COLA", reddo.TypeInt); e != nil || v.(int64) != 1 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "COLA", 1, v)
		}
		if v, e := gbo.GboGetAttr("COLB", reddo.TypeString); e != nil || v.(string) != "a string" {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "COLB", "a string", v)
		}
		if v, e := gbo.GboGetAttr("COLC", reddo.TypeFloat); e != nil || v.(float64) != 2.3 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "COLC", 2.3, v)
		}
		if v, e := gbo.GboGetAttr("COLD", reddo.TypeBool); e != nil || v.(bool) != true {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "COLD", true, v)
		}
	}
	
	rm.ColNameToGboFieldTranslator = map[string]map[string]interface{}{"table_name": {"COLA": "a", "COLB": "b", "COLC": "c", "COLD": "d"}}
	{
		gbo, err := rm.ToBo("table_name", row)
		if err != nil || gbo == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		if v, e := gbo.GboGetAttr("a", reddo.TypeInt); e != nil || v.(int64) != 1 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "a", 1, v)
		}
		if v, e := gbo.GboGetAttr("b", reddo.TypeString); e != nil || v.(string) != "a string" {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "b", "a string", v)
		}
		if v, e := gbo.GboGetAttr("c", reddo.TypeFloat); e != nil || v.(float64) != 2.3 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "c", 2.3, v)
		}
		if v, e := gbo.GboGetAttr("d", reddo.TypeBool); e != nil || v.(bool) != true {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "d", true, v)
		}
	}

	rm.ColNameToGboFieldTranslator = map[string]map[string]interface{}{"*": {"COLA": "a", "COLB": "b", "COLC": "c", "COLD": "d"}}
	{
		gbo, err := rm.ToBo("", row)
		if err != nil || gbo == nil {
			t.Fatalf("%s failed: error: %#v", name, err)
		}
		if v, e := gbo.GboGetAttr("a", reddo.TypeInt); e != nil || v.(int64) != 1 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "a", 1, v)
		}
		if v, e := gbo.GboGetAttr("b", reddo.TypeString); e != nil || v.(string) != "a string" {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "b", "a string", v)
		}
		if v, e := gbo.GboGetAttr("c", reddo.TypeFloat); e != nil || v.(float64) != 2.3 {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "c", 2.3, v)
		}
		if v, e := gbo.GboGetAttr("d", reddo.TypeBool); e != nil || v.(bool) != true {
			t.Fatalf("%s failed: expected attr[%s] to be %#v but received %#v", name, "d", true, v)
		}
	}
}
