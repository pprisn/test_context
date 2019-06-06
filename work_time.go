//Как прервать процесс
package main

import (
    "fmt"
    "log"
    "time"
)


//Рабочий процесс на 20сек.
func work() error {
    for i := 0; i < 10; i++ {
        select {
        case <-time.After(2 * time.Second):
            fmt.Println("Doing some work ", i)
        }
    }
    return nil
}

func main() {
    fmt.Println("Hey, I'm going to do some work")

    ch := make(chan error, 1)
    go func() {
        ch <- work()
    }()


    select {
    case err := <-ch:
        if err != nil {
            log.Fatal("Something went wrong :(", err)
        }
    case <-time.After(30 * time.Second): // прерываем рабочий процесс после указанного времени
        fmt.Println("Life is to short to wait that long")
    }
    // если мы это видим - рабочий процесс полностью отработал без прерывания времени ожидания
    fmt.Println("Если мы это видим - рабочий процесс полностью отработал без прерывания времени ожидания Finished. I'm going home")
}

//Но есть проблема: если моя программа все еще работает, как, 
//например, веб-сервер, даже если я не жду, пока функция work завершит работу, 
//она будет работать и потреблять ресурсы. Поэтому мне нужен способ отменить эту программу.