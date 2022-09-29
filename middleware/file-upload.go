package middleware

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func UploadFile(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, handler, err := r.FormFile("inputImage")
		if err != nil {
			json.NewEncoder(w).Encode(err)
			return
		}
		defer file.Close()

		tempFile, err := ioutil.TempFile("uploads", "*"+handler.Filename)
		if err != nil {
			json.NewEncoder(w).Encode(err)
			return
		}
		defer tempFile.Close()

		fileBytes, _ := ioutil.ReadAll(file)

		tempFile.Write(fileBytes)

		data := tempFile.Name()
		filename := data[8:]

		ctx := context.WithValue(r.Context(), "dataFile", filename)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
