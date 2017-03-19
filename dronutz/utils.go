package dronutz

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	uuid "github.com/nu7hatch/gouuid"
)

func readJSON(stream io.ReadCloser, target interface{}) {
	defer stream.Close()
	decoder := json.NewDecoder(stream)
	err := decoder.Decode(target)
	if err != nil {
		panic(err)
	}
}

func writeJSON(stream io.Writer, data interface{}) {
	encoder := json.NewEncoder(stream)
	err := encoder.Encode(data)
	if err != nil {
		panic(err)
	}
}

func writeErrorResponse(res http.ResponseWriter, err error) {
	fmt.Println("Error checking status: ", err)
	res.WriteHeader(http.StatusInternalServerError)
	return
}

func guid() string {
	uid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return uid.String()
}
