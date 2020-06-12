package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	result, err := RunInteractive(func(query string, num int) ([]ListItem, error) {
		time.Sleep(1000 * time.Millisecond)
		results := make([]ListItem, num-len(query))
		for i := 0; i < len(results); i++ {
			results[i] = ListItem{
				ID:    fmt.Sprintf("%v:%v", query, i),
				Label: fmt.Sprintf("Result %d/%d for \"%v\"", i+1, num, query),
				Score: float32(i),
			}
		}
		return results, nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("%v\n", result)
	}
}
