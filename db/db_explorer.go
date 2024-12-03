package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strconv"
)

// Объявляем структуру Хендлера:
type handler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

// Прописываем маршруты:
func NewDbExplorer(db *sql.DB) (*mux.Router, error) {

	handlers := &handler{
		DB: db,
	}

	router := mux.NewRouter()
	router.HandleFunc("/", handlers.tables).Methods("GET")
	router.HandleFunc("/{table}", handlers.table).Methods("GET")
	router.HandleFunc("/{table}/{id}", handlers.table_id).Methods("GET")
	//На PUT я остановился и не доделал:
	//router.HandleFunc("/{table}/", handlers.table_put).Methods("PUT")

	return router, nil
}

// Обращение по пути "/":
func (h *handler) tables(w http.ResponseWriter, r *http.Request) {

	var table string
	tables := make(map[string][]string)
	result := make(map[string]map[string][]string)

	//Формирование запроса к БД:
	rows, err := h.DB.Query("SHOW TABLES;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Считываем получившиеся данные:
	for rows.Next() {
		rows.Scan(&table)
		tables["tables"] = append(tables["tables"], table)
	}
	rows.Close()

	//Формируем ответ:
	result["response"] = tables
	json.NewEncoder(w).Encode(result)
}

// Обращение по пути "/{table}"
func (h *handler) table(w http.ResponseWriter, r *http.Request) {

	var table string
	tables := []string{}
	result := make(map[string]string)
	result_out := make(map[string]map[string][]map[string]interface{})
	mSum := make(map[string][]map[string]interface{})

	//Выполняем запрос к БД по получению списка таблиц:
	rows, err := h.DB.Query("SHOW TABLES;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Считываем полученные значения:
	for rows.Next() {
		rows.Scan(&table)
		tables = append(tables, table)
	}
	rows.Close()

	//Вытаскиваем параметры для работы из URL:
	vars := mux.Vars(r)
	table_name := vars["table"]

	//Определяем Query по умолчанию:
	limit := 5
	offset := 0
	//Получаем Query, если они заданы:
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))

	//Алгоритм формирования ответа по заданным условиям:
	for _, value := range tables {
		//Если запрошенная таблица равна итерируемой:
		if value == table_name {

			//Выполняем запрос к БД с получением всех данных:
			rows, err = h.DB.Query("SELECT * FROM " + table_name + ";")
			cols, _ := rows.Columns()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			count := 0
			//Формируем количество возвращаемых данных по заданной величине offset и все полученные результаты
			//переписываем в map:
			for rows.Next() {

				if count < offset {
					count++
					continue
				}

				columns := make([]interface{}, len(cols))
				columnPointers := make([]interface{}, len(cols))

				for i, _ := range columns {
					columnPointers[i] = &columns[i]
				}
				rows.Scan(columnPointers...)

				m := make(map[string]interface{})
				for i, colName := range cols {
					val := columnPointers[i].(*interface{})
					vals := *val

					if vals != nil {

						sitf, _ := vals.(interface{})
						sbyte, _ := sitf.([]byte)
						s := string(sbyte)

						if _, err := strconv.Atoi(s); err == nil {
							m[colName], _ = strconv.Atoi(s)
						} else {
							m[colName] = s
						}

					} else {
						m[colName] = vals
					}
				}
				mSum["records"] = append(mSum["records"], m)
				if len(mSum["records"]) == limit {
					break
				}
			}
			rows.Close()

			//Преобразуем полученный результата к требуемым тестом ответам:
			result_out["response"] = mSum
			json.NewEncoder(w).Encode(result_out)
			return
		}
	}
	//Формируем ответ в случае, если запрошенная таблица отсутствует в БД:
	w.WriteHeader(http.StatusNotFound)
	result["error"] = "unknown table"
	json.NewEncoder(w).Encode(result)
}

//Так как все операции по GET запросам одинаковы и содержат похожие логики, то запишим эти алгоритмы в отдельные
//функции:

// Запишем отдельную функцию поиска всех таблиц в БД:
func tables_name(h *handler, w http.ResponseWriter) []string {

	var table string
	tables := []string{}

	//Считываем все таблицы из БД:
	rows, err := h.DB.Query("SHOW TABLES;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	for rows.Next() {
		rows.Scan(&table)
		tables = append(tables, table)
	}
	rows.Close()

	return tables
}

// Реализуем отдельный алгоритм поиска всех данных из заданной таблицы:
func response_server(w http.ResponseWriter, r *http.Request, tables []string,
	table_name string, h *handler) map[string][]map[string]interface{} {

	mSum := make(map[string][]map[string]interface{})

	limit := 5
	offset := 0
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))

	for _, value := range tables {

		//Если запрошенная таблица равна итерируемой, то выполним само получение данных:
		if value == table_name {

			rows, err := h.DB.Query("SELECT * FROM " + table_name + ";")
			cols, _ := rows.Columns()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return nil
			}

			count := 0

			for rows.Next() {

				//Учтем ограничение по количеству записей:
				if count < offset {
					count++
					continue
				}

				columns := make([]interface{}, len(cols))
				columnPointers := make([]interface{}, len(cols))

				for i, _ := range columns {
					columnPointers[i] = &columns[i]
				}
				rows.Scan(columnPointers...)

				m := make(map[string]interface{})
				for i, colName := range cols {
					val := columnPointers[i].(*interface{})
					vals := *val

					if vals != nil {

						sitf, _ := vals.(interface{})
						sbyte, _ := sitf.([]byte)
						s := string(sbyte)

						if _, err := strconv.Atoi(s); err == nil {
							m[colName], _ = strconv.Atoi(s)
						} else {
							m[colName] = s
						}

					} else {
						m[colName] = vals
					}
				}
				mSum["records"] = append(mSum["records"], m)
				if len(mSum["records"]) == limit {
					break
				}
			}
			rows.Close()

			//Вернем map в качестве ответа:
			return mSum
		}
	}
	return nil
}

// Алгоритм поиска значений по заданному id:
func find_id(mSum map[string][]map[string]interface{},
	id int) map[string]map[string]interface{} {

	mSumId := make(map[string]map[string]interface{})

	//Проходим по map и записываем итог поиска в mSumId:
	for _, value_1 := range mSum {
		for _, value_2 := range value_1 {

			if value_2["id"] == id {
				mSumId["record"] = value_2
				return mSumId
			}
		}
	}
	return nil
}

// Формируем функцию для ответа по пути /{table}/{id}:
func (h *handler) table_id(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	table_name := vars["table"]
	id, _ := strconv.Atoi(vars["id"])
	mSum := make(map[string][]map[string]interface{})
	mSumId := make(map[string]map[string]interface{})
	result_out := make(map[string]map[string]map[string]interface{})
	result := make(map[string]string)

	//Считываем все таблицы:
	tables := tables_name(h, w)
	//Получаем все данные из заданной таблицы:
	mSum = response_server(w, r, tables, table_name, h)
	//Получаем все данные с заданным id из нужной нам таблицы:
	mSumId = find_id(mSum, id)

	//Формируем ответ:
	if mSumId != nil {
		result_out["response"] = mSumId
		json.NewEncoder(w).Encode(result_out)
	} else {
		w.WriteHeader(http.StatusNotFound)
		result["error"] = "record not found"
		json.NewEncoder(w).Encode(result)
	}
}

//Эту часть я не доделал:
//func (h *handler) table_put(w http.ResponseWriter, r *http.Request) {
//
//	data := make(map[string]interface{})
//	json.NewDecoder(r.Body).Decode(&data)
//	vars := mux.Vars(r)
//	table_name := vars["table"]
//	keys := []string{}
//	value := []string{}
//
//	tables := tables_name(h, w)
//
//	for _, val := range tables {
//		if val == table_name {
//			for k, v := range data {
//				keys = append(keys, k)
//				value = append(value, v.(string))
//			}
//
//			stmt := `INSERT INTO table_name (*keys) VALUES(*value);`
//			h.DB.Exec(stmt)
//		}
//	}
//
//	mSum := make(map[string][]map[string]interface{})
//	mSumId := make(map[string]map[string]interface{})
//	result_out := make(map[string]map[string]map[string]interface{})
//	result := make(map[string]string)
//
//	mSum = response_server(w, r, tables, table_name, h)
//	mSumId = find_id(mSum, 3)
//
//	if mSumId != nil {
//		result_out["response"] = mSumId
//		json.NewEncoder(w).Encode(result_out)
//	} else {
//		w.WriteHeader(http.StatusNotFound)
//		result["error"] = "record not found"
//		json.NewEncoder(w).Encode(result)
//	}
//}
