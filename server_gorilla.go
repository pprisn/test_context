package main

// Импортируем необходимые зависимости. Мы будем использовать
// пакет из стандартной библиотеки и пакет от gorilla
import (
   "time"
   "fmt"
   "net/http"
   "math/rand"
   "github.com/gorilla/mux"
)

  // Необходимо реализовать хендлер Random. 
  // Этот хендлер просто возвращает сообщение "Not Implemented"

var Random = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){

	rand.Seed(time.Now().UnixNano())
	headOrTails := rand.Intn(2)

	if headOrTails == 0 {
		time.Sleep(7 * time.Second)
		//fmt.Fprintf(w, "МЕДЛЕННЫЙ ОТВЕТ Go! slow %v", headOrTails)
		w.Write([]byte("МЕДЛЕННЫЙ ОТВЕТ Go!"))
		fmt.Printf("МЕДЛЕННЫЙ ОТВЕТ Go! slow %v \n", headOrTails)
		return
	}

	//fmt.Fprintf(w, "БЫСТРЫЙ ОТВЕТ Go! quick %v ", headOrTails)
        w.Write([]byte("БЫСТРЫЙ ОТВЕТ Go! quick!"))
	fmt.Printf("БЫСТРЫЙ ОТВЕТ Go! quick %v \n", headOrTails)
	return
   })


func main() {
    // Инициализируем gorilla/mux роутер
    r := mux.NewRouter()

    // Страница по умолчанию для нашего сайта это простой html.
    r.Handle("/", http.FileServer(http.Dir("./views/")))
    
    // Статику (картинки, скрипти, стили) будем раздавать 
    // по определенному роуту /static/{file} 
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", 
                                http.FileServer(http.Dir("./static/"))))
    r.HandleFunc("/random", Random).Methods("GET")

    // Наше приложение запускается на 1112 порту. 
    // Для запуска мы указываем порт и наш роутер
    http.ListenAndServe(":1112", r)
}