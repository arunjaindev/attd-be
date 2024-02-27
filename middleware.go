package main

import (
	"net/http"
)

func CORSMiddleware (middleHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			return
		}
		middleHandler.ServeHTTP(w, r)
	})		

}