//Simple  Lazy and Very Random Server 
package main

import (
    "fmt"
    "math/rand"
    "net/http"
    "time"
)

func main() {
    http.HandleFunc("/", LazyServer)
    http.ListenAndServe(":1112", nil)
}

// sometimes really fast server, sometimes really slow server
func LazyServer(w http.ResponseWriter, req *http.Request) {
    //Рондомизируем генератор случайных чисел
    rand.Seed(time.Now().UnixNano())
    headOrTails := rand.Intn(2)

    if headOrTails == 0 {
        time.Sleep(7 * time.Second)
        fmt.Fprintf(w, "МЕДЛЕННЫЙ ОТВЕТ Go! slow %v \n", headOrTails)
        fmt.Printf("МЕДЛЕННЫЙ ОТВЕТ Go! slow %v \n",     headOrTails)
        return
    }

    fmt.Fprintf(w, "БЫСТРЫЙ ОТВЕТ Go! quick %v \n", headOrTails)
    fmt.Printf("БЫСТРЫЙ ОТВЕТ Go! quick %v \n",     headOrTails)
    return
}