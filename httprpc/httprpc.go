package main

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
)

type server struct {
	functions map[string]FunctionHandler
}

func NewServer() *server {
	return &server{
		functions: make(map[string]FunctionHandler),
	}
}

func (s *server) AddFunction(name string, function interface{}) {
	if name == "" {
		log.Panicf("function name cannot be empty")
	}

	if _, alreadyExists := s.functions[name]; alreadyExists {
		log.Panicf("cannot add function %v, name already used", name)
	}

	if function == nil {
		log.Panicf("function cannot be nil")
	}

	s.functions[name] = GenerateHandler(function)
}

func (s *server) GetFunctionHandler(name string) (handler FunctionHandler, handlerExists bool) {
	handler, handlerExists = s.functions[name]
	return
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" && r.Method != "OPTIONS" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	call := struct {
		FunctionName string           `json:"function"`
		Args         *json.RawMessage `json:"args"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&call); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	handler, handlerExists := s.GetFunctionHandler(call.FunctionName)

	if !handlerExists {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	result, err := handler(call.Args)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

type FunctionHandler func(jsonArgs *json.RawMessage) (interface{}, error)

func GenerateHandler(function interface{}) FunctionHandler {
	t := reflect.TypeOf(function)
	v := reflect.ValueOf(function)

	argList := make([]reflect.Value, t.NumIn())
	argInterfaceList := make([]interface{}, t.NumIn())

	for i := 0; i < t.NumIn(); i++ {
		argList[i] = reflect.New(t.In(i))
		argInterfaceList[i] = argList[i].Interface()
	}

	callArgList := make([]reflect.Value, t.NumIn())

	for i, arg := range argList {
		callArgList[i] = arg.Elem()
	}

	return func(args *json.RawMessage) (result interface{}, err error) {
		if err = json.Unmarshal(*args, &argInterfaceList); err != nil {
			return nil, err
		}

		returnList := make([]interface{}, t.NumOut())

		for i, returnArg := range v.Call(callArgList) {
			returnList[i] = returnArg.Interface()
		}

		return returnList, nil
	}
}
