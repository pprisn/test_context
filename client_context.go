//
//Параллельное выполнение группы запросов и сохранение результатов
//http-запросы к серверу в обычном порядке, если сервер работает медленно, мы игнорируем (отменяем) запрос
//и выполняем быстрый возврат, чтобы мы могли управлять отменой и освободить соединение.
//По материалам блога https://blog.golang.org/context
//
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"flag"
	"golang.org/x/net/context"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

// Переменные для формирования  cpu and memory profile
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

var (
	wg        sync.WaitGroup
	countTime int
)

//var (
// ctx context.Context
// cancel context.CancelFunc
//)

//структура для хранения результатов
type words struct {
	sync.RWMutex //добавить в структуру мьютекс
	found        map[string]string
}

//Инициализация области памяти
func newWords() *words {
	return &words{found: map[string]string{}}
}

//Фиксируем вхождение слова
func (w *words) add(word string, WS string) {
	w.Lock()         //Заблокировать объект
	defer w.Unlock() // По завершению, разблокировать
	WorkStatus, ok := w.found[word]
	if !ok { //т.е. если ID запроса не найдено заводим новый элемент слайса
		w.found[word] = WS
		return
	}
	// слово найдено в очередной раз , увеличим счетчик у элемента слайса
	w.found[word] = WorkStatus + ";" + WS
	return
}

//Метод readlist читает и выводит на экран список всех значений 
func (w *words) readlist() error {
	fmt.Println("Read all worlds:")
	//	w.RLock()         // Блокировка доступа
	//	defer w.RUnlock() //разблокировка доступа
	for word, status := range w.found {
		fmt.Printf("%s;%s\n", word, status)
	}
	return nil
}
//Метод remove() удаляет записи кэша карт var w *words 
//которые имеют завершенный статосом обработки запросов на период проверки 
func (w *words) remove() error {
	w.Lock()         // Блокировка доступа
	defer w.Unlock() //разблокировка доступа
	for word, status := range w.found {
		// Если найдено 1 и более вхождения символа ; в значении элеменнта слайса,
		// считаем , что запрос полностью отработан (получен ответ сервера или установлен статус прерывания по таймауту)
		//
		if strings.Count(status, ";") > 0 {
			delete(w.found, word)
		}
	}
	return nil
}

// main
func main() {
	////////////////////////////////block for cpuprofile///////////////
	flag.Parse()
        var f os.File
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		//		defer pprof.StopCPUProfile()
		defer f.Close()
	}

	////////////////////////////////END block for cpuprofile////////////
	//Создание структуры хранения результатов
	w := newWords()

	//получим 1 раз в минуту результаты работы
	go func( f os.File) {
		for range time.Tick(time.Minute) {
			w.readlist()
			fmt.Println(len(w.found))
			w.remove()
			fmt.Println("Worlds reset:")
			fmt.Println(len(w.found))
			countTime++
			////////////////////////////////block for memprofile///////////////
			if countTime == 4 {
				if *memprofile != "" {
					fm, err := os.Create(*memprofile)
					if err != nil {
						log.Fatal("could not create memory profile: ", err)
					}
					runtime.GC() // get up-to-date statistics
					if err := pprof.WriteHeapProfile(fm); err != nil {
						log.Fatal("could not write memory profile: ", err)
					}
					fm.Close()
					pprof.StopCPUProfile()
					fmt.Println("Save memprofile")

					if *cpuprofile != "" {
						f.Close()
					}
					////////////////////////////////END block for memprofile////////////
				}

			}
		}
	}(f)

	snow := "" //переменная для формирования ID запроса
	for now := range time.Tick(5 * time.Second) {
		//Запускаем параллельные work
		for i := 0; i <= 500; i++ {
			wg.Add(1)
			snow = fmt.Sprintf("ID:%d-%v", i, now)
			go func(i int, now string) {
				// Создание контекста с ограничением времени его жизни в 4 сек
				ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
				defer cancel()
				go work(ctx, now, w)
				wg.Wait()
				//cancel()
			}(i, snow)
		}
		//wg.Wait()
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
	defer close(c)

	req, _ := http.NewRequest("GET", "http://localhost:1112/random", nil)
	go func() {
		resp, err := client.Do(req)
		//	fmt.Printf("Doing http request, %s \n",id)

		//Добавим запись в результат статусов выполнения запросов
		dict.add(id, "StartWork")

		pack := struct {
			r   *http.Response
			err error
		}{resp, err}
		c <- pack
	}()

	//        go func() {
	//          dict.add(id, "StartWork")
	//          resp, err := http.Get("http://localhost:1112")
	//		pack := struct {
	//			r   *http.Response
	//		err error
	//		}{resp, err}
	//		c <- pack
	//	}()

	// Кто первый того и тапки...
	select {
	case <-ctx.Done():
		tr.CancelRequest(req)
		<-c // Wait for client.Do
		//	fmt.Printf("Cancel context, НЕ ДОЖДАЛИСЬ ОТВЕТА СЕРВЕРА на запрос %s\n",id)
		//Добавим результат выполнения запроса со статусом CancelContext
		dict.add(id, "CancelContext")
		//fmt.Println(now.Sub(time.Now())) //замер времени протухания контекста
		return ctx.Err()
	case ok := <-c:
		err := ok.err
		resp_ := ok.r
		if err != nil {
			//fmt.Println("Error ", err)
			dict.add(id, "NoConnection")
			return err
		}
		defer resp_.Body.Close()
		out, error := ioutil.ReadAll(resp_.Body)
		if error != nil {
			//fmt.Println("Error ", err)
			dict.add(id, "NoReadBody")
			return error
		}

		//	fmt.Printf("Server Response %s:  [%s]\n", id,out)

		//Добавим результат выполнения запроса Ответ сервера
		dict.add(id, string(out))
	}

	return nil
}
