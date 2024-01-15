package fs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (fs DynamicFs) handleInfo(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	jsonString, err := json.Marshal(fs.Files())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Header().Set("content-type", "application/json")
	res.Write(jsonString)
}

func HttpHandler(fs DynamicFs) http.Handler {
	mux := http.NewServeMux()
	
	mux.Handle("/f/", http.StripPrefix("/f/", http.FileServer(http.FS(fs))))
	mux.HandleFunc("/files", fs.handleInfo)

	return mux
}
