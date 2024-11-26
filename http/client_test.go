package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

//Объявляем необходимые для работы структуры данных:

//Users - это структура, которая содержит все данные по пользователям, нужная чтобы распарсить атрибутом root в xml;
//Users_after - это структура наборов данных по пользователям, но без значения row, объявленная для удобства;
//User_my - структура данных по одному пользователю, необходимая для выполнения парсинга;
//Users_after_result - структура объявленная на основании структуры из client.go сокращенного набора данных по
//пользователям;
//Test_case - структура моих тестов для удобства.

type Users struct {
	RowID []User_my `xml:"row"`
}

type Users_after []User_my

const userElementName = "row"

type User_my struct {
	Id            int    `xml:"id"`
	Guid          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
	Name          string
}

type Users_after_result []User

type Test_case struct {
	Src         SearchRequest
	number_test int
	numder_data int
}

// Функцию main я задал, но не использовал. Просто для порядка и на всякий случай закомментил:
func main() {
	//http.HandleFunc("/", SearchServer)
	//http.ListenAndServe(":8080", nil)
}

// Данный тест я написал для проверки функции Decoder, что парсит xml. Оставил ее тут закомментированной.
//func TestDecoder(t *testing.T) {
//
//	users_f := DecoderXml()
//	slices := []int{}
//	for _, userNode := range users_f.RowID {
//		slices = append(slices, userNode.Id)
//	}
//	if slices[len(slices)-1] == 34 && slices[0] == 0 {
//		fmt.Println("ОК")
//	}
//}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	var serch SearchRequest

	//На входе в сервер проверяем сразу значение токена:
	if r.Header["Accesstoken"][0] == "12345" {

		//Считываем имеющиеся значения параметров из полученного запрос на сервер:
		Query := r.URL.Query()["query"][0]
		OrderField := r.URL.Query()["order_field"][0]
		OrderBy, _ := strconv.Atoi(r.URL.Query()["order_by"][0])

		serch.Query = Query
		serch.OrderField = OrderField
		serch.OrderBy = OrderBy

		//Используем функцию DecoderXml() для выполнения парсинга xml:
		users_f := DecoderXml()

		//Используем функцию поиска и сортировки данных в имеющемся .xml:
		result, err := SearchDate(users_f, serch)

		//Анализируем полученные ошибки с сервера:

		//1. Ошибка, когда поле Query содержит только числа, что не верно для Name и About, таким образом я смог сделать
		//работу с StatusInternalServerError:
		if err.Error() == "Request is empty" {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		//2. Ошибка долгого времени выполнения запроса. Тут я задал определенное значение Query, чтобы получить ошибку
		//"Time is out" в функции SearchDate:
		if err.Error() == "Time is out" {
			time.Sleep(5 * time.Second)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		//3. Ошибка для случая, когда значение поля order_field задано неверно:
		if err.Error() == "Bad order_field" {

			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			var eror SearchErrorResponse
			eror.Error = "ErrorBadOrderField"
			jsonResp, _ := json.Marshal(eror)
			w.Write(jsonResp)
			return
		}

		//4. Ошибка наличия опечатки в поле order_field? сзаданная, чтобы попасть по определенной ветки в FindUser:
		if err.Error() == "Specify order_field" {

			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			var eror SearchErrorResponse
			eror.Error = "Specify order_field"
			jsonResp, _ := json.Marshal(eror)
			w.Write(jsonResp)
			return
		}

		//5. Ошибка StatusBadRequest, сгенированная для передачи битого json:
		if err.Error() == "Bad json" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		//6. Успешное выполнение запроса:
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, _ := json.Marshal(result)
		w.Write(jsonResp)

	} else {
		//Случай не верной авторизации:
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}
}

// Функция DecoderXml() для парсинга заданного .xml:
func DecoderXml() Users {
	var (
		users_f Users
		user_f  User_my
	)

	//Считываем сам dataset.xml
	f, err := os.Open("dataset.xml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	//Считываем .xml
	d := xml.NewDecoder(f)

	//Производим парсинг через объявление переменной интерфейсом Token и при помощи применения свитча с типами:
	for t, _ := d.Token(); t != nil; t, _ = d.Token() {
		switch se := t.(type) {

		case xml.StartElement:
			if se.Name.Local == userElementName {
				d.DecodeElement(&user_f, &se)
				users_f.RowID = append(users_f.RowID, user_f)
			}
		}
	}
	return users_f
}

// Функция поиска данных по Query и их сортировки:
func SearchDate(users_f Users, serch SearchRequest) (Users_after, error) {

	//Логика у меня такая я объявляю новые ошибки, которые потом считывает SearchServer и потом формирует уже ответ
	//сервера с указанием .WriteHeader и так далее.

	//1. Для начала посмотрим, что за тип данных у полученного Query, если он представляет собой только числа, что
	//не может быть верным для полей Name и About, то будет ошибка на сервере http.StatusInternalServerError.
	//Я больше не придумал способов вызвать данную ошибку, чтобы попасть по определенному пути в FindUser в тестах.
	_, err_type := strconv.Atoi(serch.Query)
	if err_type == nil {
		err := errors.New("Request is empty")
		return nil, err
	}

	//2. Данная ошибка мне нужна для того, чтобы задать time.Sleep в SearchServer и получить ошибку
	//timeout for limit. Не получилось иначе мне смоделировать долгую работу сервера. Только задав определенное значение
	//Query в запросе.
	if serch.Query == "Time is out for realization" {
		err := errors.New("Time is out")
		return nil, err
	}

	//3. Ошибка для генерации битого json:
	if serch.Query == "Bad json for error" {
		err := errors.New("Bad json")
		return nil, err
	}

	var (
		find_users Users_after
		find_user  User_my
	)

	//Тут отражен метод поиска требуемых значений. Соответственно имеем два варианта формирования с и без задания Query.
	//Если Query не задан, то я просто формирую find_users без row - так удобнее работать, чтобы к ней постоянно не
	//обращаться в дальнейшем.
	if serch.Query != "" {
		for _, userNode := range users_f.RowID {
			if serch.Query == userNode.Name || strings.Contains(userNode.About, serch.Query) {

				find_user = userNode
				find_user.Name = userNode.FirstName + " " + userNode.LastName
				find_users = append(find_users, find_user)
			}
		}
	} else {
		for _, userNode := range users_f.RowID {
			find_user = userNode
			find_user.Name = userNode.FirstName + " " + userNode.LastName
			find_users = append(find_users, find_user)
		}
	}

	//Далее реализуем метод сортировки полученных данных по полю с помощью sort.Slice для различных условий задачи.
	if serch.OrderField == "Id" || serch.OrderField == "Age" || serch.OrderField == "Name" {
		if len(find_users) > 1 && serch.OrderBy != 0 {

			switch serch.OrderField {
			case "Id":

				switch serch.OrderBy {

				case 1:
					sort.Slice(find_users, func(i, j int) bool {
						return find_users[i].Id < find_users[j].Id
					})

				case -1:
					sort.Slice(find_users, func(i, j int) bool {
						return find_users[i].Id > find_users[j].Id
					})
				}

			case "Age":

				switch serch.OrderBy {

				case 1:
					sort.Slice(find_users, func(i, j int) bool {
						return find_users[i].Age < find_users[j].Age
					})

				case -1:
					sort.Slice(find_users, func(i, j int) bool {
						return find_users[i].Age > find_users[j].Age
					})
				}

			case "Name":

				switch serch.OrderBy {

				case 1:
					sort.Slice(find_users, func(i, j int) bool {
						return find_users[i].Name < find_users[j].Name
					})

				case -1:
					sort.Slice(find_users, func(i, j int) bool {
						return find_users[i].Name > find_users[j].Name
					})
				}

			case "":

				switch serch.OrderBy {

				case 1:
					sort.Slice(find_users, func(i, j int) bool {
						return find_users[i].Name < find_users[j].Name
					})

				case -1:
					sort.Slice(find_users, func(i, j int) bool {
						return find_users[i].Name > find_users[j].Name
					})
				}
			}
		}
	} else {

		//В том случае, если у нас поля заданы с опечатками я решил реализовать данный функционал. Логика такая, что
		//если запрос OrderField содержит наименование поля с лишними символами сервер ему говорит о наличии опечатки
		//в запросе. Также объявляем новую ошибку и по ней уже в SerchServer даем ответ сервера.
		if strings.Contains(serch.OrderField, "Id") || strings.Contains(serch.OrderField, "Age") ||
			strings.Contains(serch.OrderField, "Name") {

			err := errors.New("Specify order_field")
			return nil, err
		}

		//Если запрос вообще не похож на нужное поле OrderField, то дается другая ошибка. В SearchServer она
		//обрабатывается, как http.StatusBadRequest.
		err := errors.New("Bad order_field")
		return nil, err
	}

	err := errors.New("ОК")
	return find_users, err
}

// Основной тест по проверке работы сервера:
func TestSearchServer(t *testing.T) {

	var (
		find_users Users_after_result
		find_user  User
	)

	cases := []Test_case{

		// Набор тестовых данных для проверки логики кода функции SerchDate (ищет нужные поля и сортирует), где я
		//менял разные числа и анализировал покрытия и результаты. Все тесты по этой части, так как их не нужно было
		//делать по заданию я удалил, так как они все однотипные и не интересные.
		{
			Src: SearchRequest{
				Limit:      10,
				Offset:     10,
				Query:      "Aliquip",
				OrderField: "Name",
				OrderBy:    -1,
			},
			number_test: 1,
			numder_data: 1},

		// Для проверки допустимости параметра Limit в случае, когда он меньше 0:
		{
			Src: SearchRequest{
				Limit:      -5,
				Offset:     10,
				Query:      "Aliquip",
				OrderField: "Name",
				OrderBy:    -1,
			},
			number_test: 2,
			numder_data: 2},

		// Для проверки допустимости параметра Limit в слчае, когда он больше 25:
		{
			Src: SearchRequest{
				Limit:      30,
				Offset:     10,
				Query:      "Aliquip",
				OrderField: "Name",
				OrderBy:    -1,
			},
			number_test: 2,
			numder_data: 3},

		// Проверка по допустимости значения offset
		{
			Src: SearchRequest{
				Limit:      1,
				Offset:     -3,
				Query:      "Aliquip",
				OrderField: "Name",
				OrderBy:    -1,
			},
			number_test: 2,
			numder_data: 4},

		// Проверка поля OrderField при указании недопустимого значения:
		{
			Src: SearchRequest{
				Limit:      1,
				Offset:     1,
				Query:      "Aliquip",
				OrderField: "Epsent",
				OrderBy:    -1,
			},
			number_test: 2,
			numder_data: 5},

		// В этом тесте я задал недопустимое значение поля Query и проверил работу своей функции SerchDate (ищет нужные
		// поля и сортирует)
		{
			Src: SearchRequest{
				Limit:      1,
				Offset:     1,
				Query:      "Tambov",
				OrderField: "Id",
				OrderBy:    -1,
			},
			number_test: 2,
			numder_data: 6},

		// Нужно было для того, чтобы зайти в условие if len(data) == req.Limit:
		{
			Src: SearchRequest{
				Limit:      3,
				Offset:     1,
				Query:      "Aliquip",
				OrderField: "Name",
				OrderBy:    1,
			},
			number_test: 2,
			numder_data: 7},

		// Здесь я просто не задал URL для команды FindUser (3 номер теста ниже) и получил unknown error Get:
		{
			Src: SearchRequest{
				Limit:      3,
				Offset:     1,
				Query:      "Aliquip",
				OrderField: "Name",
				OrderBy:    1,
			},
			number_test: 3,
			numder_data: 8},

		// Здесь я указал не верный токен:
		{
			Src: SearchRequest{
				Limit:      3,
				Offset:     1,
				Query:      "Aliquip",
				OrderField: "Name",
				OrderBy:    1,
			},
			number_test: 4,
			numder_data: 9},

		// Тут я задал условие для получения SearchServer fatal error ввиде, что если Query число, что не может быть для
		//Name и About, то сервер выдает ошибку.
		{
			Src: SearchRequest{
				Limit:      3,
				Offset:     1,
				Query:      "12",
				OrderField: "Name",
				OrderBy:    1,
			},
			number_test: 2,
			numder_data: 10},

		// Тут у меня прописано условие для timeout, я просто сделал условие на сервере, что при получении ниже
		//указанного значения Query мы на сервере заходим в цикл, где стоит таймер на 5 секунд. Другого способа
		//увеличить время я не нашел (пробовал разные запросы):
		{
			Src: SearchRequest{
				Limit:      3,
				Offset:     1,
				Query:      "Time is out for realization",
				OrderField: "Name",
				OrderBy:    1,
			},
			number_test: 2,
			numder_data: 11},

		// Тут прописано условие для получения неизвестной ошибки unknown bad request error: Specify order_field.
		///Логика была следующая, что если человек пишет Name, Age или Id с лишними символами, то сервер сообщает, что в
		//запросе имеется опечатка.
		{
			Src: SearchRequest{
				Limit:      3,
				Offset:     1,
				Query:      "Aliquip",
				OrderField: "Namer",
				OrderBy:    1,
			},
			number_test: 2,
			numder_data: 12},

		//	Генерация ошибки с битым json:
		{
			Src: SearchRequest{
				Limit:      3,
				Offset:     1,
				Query:      "Bad json for error",
				OrderField: "Name",
				OrderBy:    1,
			},
			number_test: 2,
			numder_data: 12},
	}

	// Заходим в цикл для проверки выше написанных тестов:

	for num := 0; num < len(cases); num++ {

		// 1 тест для проверки моей функции SerchDate с условием возврата, где проверяется сортировка;
		// 2 тест общий просто запускает сервер и возвращает ошибку;
		// 3 тест в котором был запущен сервер, но данные по его URL не были переданы в FindUsers;
		// 4 тест где я передаю не верный токен.
		// Процесс запуска сервера и обращения к функции FindUsers:

		switch cases[num].number_test {

		case 1:

			ts := httptest.NewServer(http.HandlerFunc(SearchServer))
			var scl SearchClient
			scl.URL = ts.URL
			scl.AccessToken = "12345"

			result, err := scl.FindUsers(cases[num].Src)
			if err != nil {
				continue
			}

			for _, userNode := range result.Users {
				find_user = userNode
				find_users = append(find_users, find_user)
			}

			for num := 0; num < len(find_users)-1; num++ {
				if find_users[num].Name < find_users[num+1].Name {
					fmt.Println("Набор данных номер:", cases[num].numder_data)
					fmt.Println("Номер теста:", cases[num].number_test)
					fmt.Println("Error")
					break
				}
			}
			//fmt.Println(find_users)
			ts.Close()

		case 2:

			ts := httptest.NewServer(http.HandlerFunc(SearchServer))
			var scl SearchClient
			scl.URL = ts.URL
			scl.AccessToken = "12345"

			_, err := scl.FindUsers(cases[num].Src)

			if err != nil {
				fmt.Println("Набор данных номер:", cases[num].numder_data)
				fmt.Println("Номер теста:", cases[num].number_test)
				fmt.Println(err)
				continue
			}

			//fmt.Println(find_users)
			ts.Close()

		case 3:

			ts := httptest.NewServer(http.HandlerFunc(SearchServer))
			var scl SearchClient
			scl.AccessToken = "12345"

			_, err := scl.FindUsers(cases[num].Src)

			if err != nil {
				fmt.Println("Набор данных номер:", cases[num].numder_data)
				fmt.Println("Номер теста:", cases[num].number_test)
				fmt.Println(err)
				continue
			}

			//fmt.Println(find_users)
			ts.Close()

		case 4:

			ts := httptest.NewServer(http.HandlerFunc(SearchServer))
			var scl SearchClient
			scl.URL = ts.URL
			scl.AccessToken = "12"

			_, err := scl.FindUsers(cases[num].Src)

			if err != nil {
				fmt.Println("Набор данных номер:", cases[num].numder_data)
				fmt.Println("Номер теста:", cases[num].number_test)
				fmt.Println(err)
				continue
			}

			//fmt.Println(find_users)
			ts.Close()
		}
	}
}
