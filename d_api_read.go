package uadmin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

func dAPIReadHandler(w http.ResponseWriter, r *http.Request, s *Session) {
	var err error
	var rowsCount int64

	urlParts := strings.Split(r.URL.Path, "/")
	modelName := urlParts[0]
	model, _ := NewModel(modelName, false)
	params := getURLArgs(r)
	schema, _ := getSchema(modelName)

	// Check permission
	allow := false
	if disableReader, ok := model.Interface().(APIDisabledReader); ok {
		allow = disableReader.APIDisabledRead(r)
		// This is a "Disable" method
		allow = !allow
		if !allow {
			w.WriteHeader(401)
			ReturnJSON(w, r, map[string]interface{}{
				"status":  "error",
				"err_msg": "Permission denied",
			})
			return
		}
	}
	if publicReader, ok := model.Interface().(APIPublicReader); ok {
		allow = publicReader.APIPublicRead(r)
	}
	if !allow && s != nil {
		allow = s.User.GetAccess(modelName).Read
	}
	if !allow {
		w.WriteHeader(401)
		ReturnJSON(w, r, map[string]interface{}{
			"status":  "error",
			"err_msg": "Permission denied",
		})
		return
	}

	// Check if log is required
	log := APILogRead
	if logReader, ok := model.Interface().(APILogReader); ok {
		log = logReader.APILogRead(r)
	}

	if len(urlParts) == 2 {
		// Read Multiple
		var m interface{}

		SQL := "SELECT {FIELDS} FROM {TABLE_NAME}"
		if val, ok := params["$distinct"]; ok && val == "1" {
			SQL = "SELECT DISTINCT {FIELDS} FROM {TABLE_NAME}"
		}

		tableName := schema.TableName
		SQL = strings.Replace(SQL, "{TABLE_NAME}", tableName, -1)

		f, customSchema := getQueryFields(r, params, tableName)
		if f != "" {
			SQL = strings.Replace(SQL, "{FIELDS}", f, -1)
		} else {
			SQL = strings.Replace(SQL, "{FIELDS}", "*", -1)
		}

		join := getQueryJoin(r, params, tableName)
		if join != "" {
			SQL += " " + join
		}

		// Get filters from request
		q, args := getFilters(r, params, tableName, &schema)

		// Apply List Modifier from Schema
		if schema.ListModifier != nil {
			lmQ, lmArgs := schema.ListModifier(&schema, &s.User)

			if lmQ != "" {
				if q != "" {
					q += " AND "
				}

				// Add extra filters from list modifier
				q += lmQ
				args = append(args, lmArgs...)
			}
		}
		if q != "" {
			SQL += " WHERE " + q
		}

		groupBy := getQueryGroupBy(r, params)
		if groupBy != "" {
			SQL += " GROUP BY " + groupBy
		}
		order := getQueryOrder(r, params)
		if order != "" {
			SQL += " ORDER BY " + order
		}
		limit := getQueryLimit(r, params)
		if limit != "" {
			SQL += " LIMIT " + limit
		}
		offset := getQueryOffset(r, params)
		if offset != "" {
			SQL += " OFFSET " + offset
		}

		if DebugDB {
			Trail(DEBUG, SQL)
			Trail(DEBUG, "%#v", args)
		}

		//var rows *sql.Rows

		if !customSchema {
			mArray, _ := NewModelArray(modelName, true)
			m = mArray.Interface()
		} else {
			m = []map[string]interface{}{}
		}

		driver, supported := sqlDialect[Database.Type]
		if !supported {
			panic(fmt.Errorf("dAPIReadHandler database '%v' not supported", Database.Type))
		}

		rowsCount, m, err = driver.apiRead(SQL, args, m, customSchema)
		if err != nil {
			w.WriteHeader(500)
			ReturnJSON(w, r, map[string]interface{}{
				"status":  "error",
				"err_msg": "Unable to execute SQL. " + err.Error(),
			})
			Trail(ERROR, "SQL: %v\nARGS: %v", SQL, args)
			return
		}

		// Preload
		if params["$preload"] == "1" {
			mList := reflect.ValueOf(m)
			for i := 0; i < mList.Elem().Len(); i++ {
				Preload(mList.Elem().Index(i).Addr().Interface())
			}
		}

		// Process M2M
		getQueryM2M(params, m, customSchema, modelName)

		returnDAPIJSON(w, r, map[string]interface{}{
			"status": "ok",
			"result": m,
		}, params, "read", model.Interface())
		go func() {
			if log {
				createAPIReadLog(modelName, 0, rowsCount, params, &s.User, r)
			}
		}()
		return
	} else if len(urlParts) == 3 {
		// Read One
		m, _ := NewModel(modelName, true)
		Get(m.Interface(), "id = ?", urlParts[2])
		rowsCount = 0

		var i interface{}
		if int(GetID(m)) != 0 {
			i = m.Interface()
			rowsCount = 1
		}

		if params["$preload"] == "1" {
			Preload(m.Interface())
		}

		returnDAPIJSON(w, r, map[string]interface{}{
			"status": "ok",
			"result": i,
		}, params, "read", model.Interface())
		go func() {
			if log {
				createAPIReadLog(modelName, int(GetID(m)), rowsCount, map[string]string{"id": urlParts[2]}, &s.User, r)
			}
		}()
	} else {
		// Error: Unknown format
		w.WriteHeader(404)
		ReturnJSON(w, r, map[string]interface{}{
			"status":  "error",
			"err_msg": "invalid format (" + r.URL.Path + ")",
		})
		return
	}
}

func createAPIReadLog(modelName string, ID int, rowsCount int64, params map[string]string, user *User, r *http.Request) {
	vals := map[string]interface{}{
		"params":     params,
		"rows_count": rowsCount,
		"_IP":        r.RemoteAddr,
	}
	output, _ := json.Marshal(vals)

	log := Log{
		Username:  user.Username,
		Action:    Action(0).Read(),
		TableName: modelName,
		TableID:   ID,
		Activity:  string(output),
	}
	log.Save()

}
