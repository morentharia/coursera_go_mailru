// обращаю ваше внимание - в этом задании запрещены глобальные переменные
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type DBExplorer struct {
	db     *sql.DB
	tables map[string]*tableInfo
	pathRe *regexp.Regexp
}

type tableInfo struct {
	name    string
	columns []*columnInfo
}

type columnInfo struct {
	field      string
	typeName   string
	isNull     bool
	key        string
	defaultVal *string
	extra      string
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	exp := DBExplorer{
		db:     db,
		tables: make(map[string]*tableInfo),
		pathRe: regexp.MustCompile("/?(\\w+)?(?:/(\\d+))?"),
	}
	qTables, err := exp.db.Query("SHOW TABLES")
	if err != nil {
		log.Fatal(err)
	}

	tableInfoList := make([]*tableInfo, 0)
	for qTables.Next() {
		var tableName string
		err = qTables.Scan(&tableName)
		if err != nil {
			log.Fatal(err)
		}
		tableInfoList = append(tableInfoList, &tableInfo{name: tableName})
	}

	for _, t := range tableInfoList {
		qColumns, err := exp.db.Query(fmt.Sprintf("SHOW COLUMNS FROM `%s`", t.name))
		if err != nil {
			log.Fatal(err)
		}

		for qColumns.Next() {
			col := &columnInfo{}
			var isNull string
			err = qColumns.Scan(
				&col.field,
				&col.typeName,
				&isNull,
				&col.key,
				&col.defaultVal,
				&col.extra,
			)
			if isNull == "YES" {
				col.isNull = true
			}
			t.columns = append(t.columns, col)
		}
		exp.tables[t.name] = t
	}
	return exp, nil
}

func (exp DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathParsed := exp.pathRe.FindStringSubmatch(r.URL.Path)
	tableName, tableID := pathParsed[1], pathParsed[2]

	if tableName != "" {
		if !exp.isTableExists(tableName) {
			JSON(w, http.StatusNotFound, "", nil, "unknown table")
			return
		}
	}

	ctx := context.WithValue(context.Background(), "tableName", tableName)
	ctx = context.WithValue(ctx, "tableID", tableID)

	switch r.Method {
	case "GET":
		if tableName == "" && tableID == "" {
			exp.showTablesHandler(w, r)
		} else if tableName != "" && tableID == "" {
			exp.selectHandler(ctx, w, r)
		} else if tableName != "" && tableID != "" {
			exp.selectByID(ctx, w, r)
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
	case "POST":
		if tableName != "" && tableID != "" {
			exp.updateHandler(ctx, w, r)
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
	case "PUT":
		if tableName != "" && tableID == "" {
			exp.insertHandler(ctx, w, r)
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
	case "DELETE":
		if tableName != "" && tableID != "" {
			exp.deleteHandler(ctx, w, r)
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}

func JSON(w http.ResponseWriter, status int, responseName string, data interface{}, errorText string) {
	type response struct {
		Response map[string]interface{} `json:"response,omitempty"`
		Error    string                 `json:"error,omitempty"`
	}
	w.Header().Set("Content-Type", "application/json")
	var resp response
	if errorText != "" {
		resp = response{
			Error: errorText,
		}
	} else {
		resMap := make(map[string]interface{})
		resMap[responseName] = data
		resp = response{
			Response: resMap,
		}
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Write(jsonResp)
}

func (exp *DBExplorer) isTableExists(tableName string) bool {
	_, ok := exp.tables[tableName]
	return ok
}

func (exp *DBExplorer) TableInfo(tableName string) *tableInfo {
	table, _ := exp.tables[tableName]
	return table
}

func (exp *DBExplorer) ColumnsInfo(tableName string) []*columnInfo {
	table, _ := exp.tables[tableName]
	return table.columns
}

func (exp *DBExplorer) ColumnNames(tableName string) []string {
	table, _ := exp.tables[tableName]
	colNames := make([]string, 0)
	for _, col := range table.columns {
		colNames = append(colNames, col.field)
	}
	return colNames
}

func (exp *DBExplorer) showTablesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := exp.db.Query("SHOW TABLES")
	if err != nil {
		JSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	resp := make([]string, 0)
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			JSON(w, http.StatusInternalServerError, "", nil, err.Error())
			return
		}
		resp = append(resp, tableName)
	}
	JSON(w, http.StatusOK, "tables", resp, "")
}

func (exp *DBExplorer) selectHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tableName := ctx.Value("tableName").(string)

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 100
	}

	resp, err := exp.selectAll(tableName, offset, limit)
	if err != nil {
		JSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	JSON(w, http.StatusOK, "records", resp, "")
}

func (exp *DBExplorer) selectByID(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tableName := ctx.Value("tableName").(string)
	id, _ := strconv.Atoi(ctx.Value("tableID").(string))
	resp, err := exp.selectWhereIDeq(tableName, id)
	if err != nil {
		JSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	if len(resp) == 0 {
		JSON(w, http.StatusNotFound, "", nil, "record not found")
		return
	}

	JSON(w, http.StatusOK, "record", resp[0], "")
}

func (exp *DBExplorer) selectAll(tableName string, offset, limit int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(
		`SELECT * FROM %s LIMIT %d OFFSET %d`,
		tableName, limit, offset,
	)

	return exp.newExecuteQuery(tableName, query)
}

func (exp *DBExplorer) selectWhereIDeq(tableName string, id int) ([]map[string]interface{}, error) {
	colNames := exp.ColumnNames(tableName)

	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE %s = %d`,
		strings.Join(colNames, ", "),
		tableName,
		colNames[0],
		id,
	)

	return exp.newExecuteQuery(tableName, query)
}

func (exp *DBExplorer) newExecuteQuery(tableName string, query string) ([]map[string]interface{}, error) {
	rows, err := exp.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dest := make([]interface{}, 0)

	columnNames := exp.ColumnNames(tableName)
	columnTypes, _ := rows.ColumnTypes()
	for _, item := range columnTypes {
		switch item.DatabaseTypeName() {
		case "INT":
			dest = append(dest, new(int))
		case "VARCHAR", "TEXT":
			dest = append(dest, new(sql.NullString))
		default:
		}
	}
	resp := make([]map[string]interface{}, 0)
	for rows.Next() {
		rows.Scan(dest...)
		row := make(map[string]interface{}, 0)
		for i, item := range dest {
			switch v := item.(type) {
			case *int:
				row[columnNames[i]] = *v
			case *sql.NullString:
				if v.Valid {
					row[columnNames[i]] = v.String
				} else {
					row[columnNames[i]] = nil
				}
			}
		}
		resp = append(resp, row)
	}

	return resp, nil
}

func (exp *DBExplorer) updateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var err error

	tableName := ctx.Value("tableName").(string)

	id, _ := strconv.Atoi(ctx.Value("tableID").(string))

	data := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		JSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	err = exp.validate(tableName, data)
	if err != nil {
		JSON(w, http.StatusBadRequest, "", nil, err.Error())
		return
	}

	rowsAffected, err := exp.update(tableName, id, data)
	if err != nil {
		JSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	JSON(w, http.StatusOK, "updated", rowsAffected, "")
}

func (exp *DBExplorer) validate(tableName string, data map[string]interface{}) error {
	table := exp.TableInfo(tableName)
	for _, column := range table.columns {
		val, ok := data[column.field]
		if !ok {
			continue
		}
		if val == nil {
			if column.isNull {
				continue
			}
			return fmt.Errorf("field %s have invalid type", column.field)
		}

		switch reflect.TypeOf(val).Name() {
		case "string":
			switch column.typeName {
			case "varchar(255)", "text":
				continue
			}

		}
		return fmt.Errorf("field %s have invalid type", column.field)
	}
	return nil
}

func (exp *DBExplorer) update(tableName string, id int, data map[string]interface{}) (int64, error) {
	columns := exp.ColumnsInfo(tableName)

	placeholders := make([]string, 0)
	values := make([]interface{}, 0)

	for k, v := range data {
		values = append(values, v)
		placeholders = append(placeholders, fmt.Sprintf("%v = ?", k))
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = %d",
		tableName,
		strings.Join(placeholders, ", "),
		columns[0].field,
		id,
	)
	res, err := exp.db.Exec(query, values...)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}

func (exp *DBExplorer) insertHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tableName := ctx.Value("tableName").(string)
	table := exp.TableInfo(tableName)

	data := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		JSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	lastID, err := exp.insert(tableName, data)
	if err != nil {
		JSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	JSON(w, http.StatusOK, table.columns[0].field, lastID, "")
}

func (exp *DBExplorer) insert(tableName string, data map[string]interface{}) (int64, error) {
	columns := exp.ColumnsInfo(tableName)

	values := make([]interface{}, 0)
	for i := 1; i < len(columns); i++ {
		val, ok := data[columns[i].field]
		if !ok {
			if columns[i].isNull {
				val = nil
			} else {
				val = ""
			}
		}
		values = append(values, val)
	}
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(exp.ColumnNames(tableName)[1:], ", "),
		"?"+strings.Repeat(", ?", len(values)-1),
	)
	res, err := exp.db.Exec(query, values...)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

func (exp *DBExplorer) deleteHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tableName := ctx.Value("tableName").(string)
	id, _ := strconv.Atoi(ctx.Value("tableID").(string))

	rowsAffected, err := exp.delete(tableName, id)
	if err != nil {
		JSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	JSON(w, http.StatusOK, "deleted", rowsAffected, "")
}

func (exp *DBExplorer) delete(tableName string, id int) (int64, error) {
	table := exp.TableInfo(tableName)

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = ?",
		table.name,
		table.columns[0].field,
	)
	res, err := exp.db.Exec(query, id)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}
