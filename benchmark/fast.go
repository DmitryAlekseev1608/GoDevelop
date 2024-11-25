package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	// Пакет с функцией ioutil.ReadAll(file) очень медленный,
	// первым делом я начал его заменять. Сначала пробовал io.Copy(w, file),
	// но значительно увеличить показатели у меня не получилось, только немного улучшился показатель B/op
	// ReadAll методы всегда тратят, как я понял много ресурсов, так как читают файл целиком и используются только для
	// маленьких по размеру данных.
	// Поэтому создаем сканер для улучшения показателей с применением буферизации для
	// уменьшения системных вызовов.

	// Было
	// fileContents, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	panic(err)
	//}

	// Стало с использованием сканера. Объявляем fileContents сканером
	fileContents := bufio.NewScanner(file)

	r := regexp.MustCompile("@")
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	// Еще в литературе по задаче советовали вынести очень много занимающую ресурсов операцию
	// regexp.MatchString из цикла и это очень сильно помогло.

	users := make([]map[string]interface{}, 0)

	// Было ранее в коде с излишним преобразованием в строковые данные
	//lines := strings.Split(string(fileContents), "\n")

	// Стало:
	fileContents.Split(bufio.ScanLines)
	// Тут мы напрямую работаем уже со сканером

	// Было:
	//for _, line := range lines {

	// Стало:
	for fileContents.Scan() {
		user := make(map[string]interface{})
		// fmt.Printf("%v %v\n", err, line)
		err := json.Unmarshal(fileContents.Bytes(), &user)

		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	for i, user := range users {

		isAndroid := false
		isMSIE := false

		browsers, ok := user["browsers"].([]interface{})
		//fmt.Print(ok)
		if !ok {
			// log.Println("cant cast browsers")
			continue
		}

		// Делаем компиляцию единожды, а не многократно, объявляем regexp.MustCompile для двух случаев:
		var pattern_Android = regexp.MustCompile("Android")
		var pattern_MSIE = regexp.MustCompile("MSIE")

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			//fmt.Print(browser)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}

			// Было
			//if ok, err := regexp.MatchString("Android", browser); ok && err == nil {

			// Стало:
			if ok := pattern_Android.MatchString(browser); ok {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				// log.Println("cant cast browser to string")
				continue
			}

			// Было
			//if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil {

			// Стало:
			if ok := pattern_MSIE.MatchString(browser); ok {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := r.ReplaceAllString(user["email"].(string), " [at] ")
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
