// http://dahernan.github.io/2015/02/04/context-and-cancellation-of-goroutines/
//Для отмены программы мы можем использовать контекстный пакет. 
//Мы должны изменить функцию work, чтобы она принимала аргумент типа 
//context.Context, обычно это первый аргумент.
package main

import (
    "fmt"
    "sync"
    "time"
    "golang.org/x/net/context"
)

var (
    wg sync.WaitGroup
)

func work(ctx context.Context) error {
    defer wg.Done()

    for i := 0; i < 3; i++ {
        select {
        case <-time.After(2 * time.Second):
            fmt.Println("Doing some work ", i)

        // we received the signal of cancelation in this channel    
        case <-ctx.Done():
            fmt.Println("Cancel the context ", i)
            return ctx.Err()
        }
    }
    return nil
}

func main() {   
    // Установим ограничение по таймауту 4 сек для ожидания контекста по запросу
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    fmt.Println("Hey, I'm going to do some work")

    wg.Add(1)
    go work(ctx)
    wg.Wait()

    fmt.Println("Finished. I'm going home")

    //Тест - проверка того что ctx.Done() отрабатывает
    time.Sleep(time.Second * 10)
        select {
        // we received the signal of cancelation in this channel    
        case <-ctx.Done():
            fmt.Println("Cancel the context ")
        }
    //

}

