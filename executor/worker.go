package executor

import (
	"reflect"
	"sync"
)

func Worker(wg *sync.WaitGroup, jobs <-chan []interface{}, method interface{}) {
	defer wg.Done()
	methodValue := reflect.ValueOf(method)

	for args := range jobs {
		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			in[i] = reflect.ValueOf(arg)
		}
		methodValue.Call(in)
	}
}
