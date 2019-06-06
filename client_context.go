//Небольшое приложение для параллельного выполнения группы запросов и получения результатов
//http-запросы к серверу в обычном порядке, если сервер работает медленно, мы игнорируем (отменяем) запрос 
//и выполняем быстрый возврат, чтобы мы могли управлять отменой и освободить соединение.

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"strings"

	"golang.org/x/net/context"
)

var (
	wg sync.WaitGroup
)

//структура для хранения результатов
type words struct {
	sync.Mutex //добавить в структуру мьютекс
	found map[string]string
}


//Инициализация области памяти
func newWords() *words { 
	return &words{found: map[string]string{}}
    }


//Фиксируем вхождение слова
func (w *words) add(word string, WS string){
	w.Lock()        //Заблокировать объект
	defer w.Unlock() // По завершению, разблокировать
	WorkStatus, ok := w.found[word]
	if !ok { //т.е. если ID запроса не найдено заводим новый элемент слайса
		w.found[word] = WS
		return
	}
	// слово найдено в очередной раз , увеличим счетчик у элемента слайса
	w.found[word] = WorkStatus +";"+WS
}


//Вывести список
func (w *words) readlist() error{
	fmt.Println("Read all worlds:")
	w.Lock() // Блокировка доступа
	for word, status := range w.found {
		fmt.Printf("%s;%s\n", word, status)
	}
	w.Unlock() //разблокировка доступа

	return nil
}


func (w *words) remove() error{
    w.Lock() // Блокировка доступа
    for word, status := range w.found {
        // Если найдено 1 и более вхождения символа ; в значении элеменнта слайса,  
        // считаем , что запрос полностью отработан (получен ответ сервера или установлен статус прерывания по таймауту)
	// 
        if strings.Count(status, ";") > 0 {
            delete(w.found,word)
        }
    }
    w.Unlock() //разблокировка доступа
    return nil
}


// main 
func main() {
        //Создание структуры хранения результатов
	w := newWords()

        //получим 1 раз в минуту результаты работы
	go func() {
		for _ = range time.Tick(time.Minute) {
			w.readlist()
                        fmt.Println(len(w.found))
			w.remove()
			fmt.Println("Worlds reset:")
                        fmt.Println(len(w.found))
		}
	}()

        for now := range time.Tick( 1 * time.Second) {
          //Запускаем параллельные work 
          for i:=0; i<= 10; i++ {
		wg.Add(1)
		go func(i int, now string) {
			// Создание контекста с ограничением времени его жизни в 4 сек
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()
			id := fmt.Sprintf("ID:%d-%s",i,now)
			go work(ctx, id, w )
			wg.Wait()
		}(i, fmt.Sprintf("%v",now))
	  }
	}
	fmt.Println("Finished.")
}

// work() - функция выполнения запроса и получения результата.
// Результатом работы является запись в структуру значения ID-идентификатора запроса 
// и результата ответа сервера или 
// статус прерывания работы при достижении ограничения времени жизни контекста запроса
// Параметры: 
// ctx context.Context - контекст запроса
// id string идентификатор запроса
// dict *words - указатель на структуру хранения результатов выполнения запросов
func work(ctx context.Context, id string, dict *words) error {
	defer wg.Done()
	//Формируем структуру заголовков запроса
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	// канал для распаковки данных anonymous struct to pack and unpack data in the channel
	c := make(chan struct {
		r   *http.Response
		err error
	}, 1)

	req, _ := http.NewRequest("GET", "http://localhost:1112", nil)
	go func() {
		resp, err := client.Do(req)
	//	fmt.Printf("Doing http request, %s \n",id)
              
              //Добавим запись в результат статусов выполнения запросов
               dict.add(id,"StartWork")

		pack := struct {
			r   *http.Response
			err error
		}{resp, err}
		c <- pack
	}()
	
        // Кто первый того и тапки...	
	select {
	case <-ctx.Done():
		tr.CancelRequest(req)
		<-c // Wait for client.Do
	//	fmt.Printf("Cancel context, НЕ ДОЖДАЛИСЬ ОТВЕТА СЕРВЕРА на запрос %s\n",id)
              //Добавим результат выполнения запроса со статусом CancelContext
                dict.add( id,"CancelContext")
		return ctx.Err()
	case ok := <-c:
		err := ok.err
		resp_ := ok.r
		if err != nil {
			fmt.Println("Error ", err)
			return err
		}
		defer resp_.Body.Close()
		out, _ := ioutil.ReadAll(resp_.Body)
	//	fmt.Printf("Server Response %s:  [%s]\n", id,out)

              //Добавим результат выполнения запроса Ответ сервера
               dict.add(id, string(out))

	}
    
	return nil
}
