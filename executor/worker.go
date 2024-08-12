package executor

import (
	"reflect"
	"sync"
)

func Worker(wg *sync.WaitGroup, jobs <-chan []interface{}, method interface{}) {
	defer wg.Done()
	methodValue := reflect.ValueOf(method)

	for args := range jobs {
		if methodValue.Kind() != reflect.Func {
			continue
		}

		if len(args) != methodValue.Type().NumIn() {
			continue
		}

		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			if arg == nil {
				in[i] = reflect.Zero(methodValue.Type().In(i))
			} else {
				in[i] = reflect.ValueOf(arg)
			}
		}

		if len(in) == methodValue.Type().NumIn() {
			methodValue.Call(in)
		}
	}
}
