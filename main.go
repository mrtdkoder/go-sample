package main

import (
	"crypto/sha1"
	"fmt"
	"go-sample/controllers"
	"io"
	"net/http"
	"os"
)

// type LocalApiPackage models.ApiPackage
// https://www.flightradar24.com/VATOZ21/3c696f0c
func main() {

	var os_name string
	var err error
	var http_server = &http.Server{
		Addr: ":8081",

		// Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 	fmt.Fprintf(w, "Hello, %s! You are running this application on %s\n", "Go Developers", r.Host)
		// 	fmt.Println(" > Hostname:", r.Host)
		// 	fmt.Println(" > Remote Address:", r.RemoteAddr)
		// 	fmt.Println(" > Request Method:", r.Method)
		// 	fmt.Println(" > Request URI:", r.RequestURI)
		// 	fmt.Println(" > User Agent:", r.UserAgent())
		// 	fmt.Println(" > Request Headers:")
		// 	for key, value := range r.Header {
		// 		fmt.Printf("\t %s : %s\n", key, value)
		// 	}
		// 	fmt.Println(" > Request Protocol:", r.Proto)
		// 	fmt.Println(" > Request Content Length:", r.ContentLength)
		// 	fmt.Println(" > Request Remote Address:", r.RemoteAddr)
		// 	//fmt.Println(" > Request Host:", r.GetBody)
		// 	fmt.Println("Received a request from", r.RemoteAddr)
		// }),
	}

	http_server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s! You are running this application on %s\n", "Go Developers", r.Host)
		fmt.Println(" > Hostname:", r.Host)
	})

	go http_server.ListenAndServe()
	name := "Go Developers"
	fmt.Println("Azure for", name)
	fmt.Scanf("%s", &name)
	fmt.Println("Welcome to Azure, dear", name)
	os_name, err = os.Hostname()

	fmt.Println("Running on", os_name)
	if err != nil {
		fmt.Println(" > Error retrieving hostname:", err)
	} else {
		fmt.Println(" > Hostname retrieved successfully")
	}

	// var http_client = &http.Client{}
	// http_client.Timeout = 10 // Set a timeout for the HTTP client
	// fmt.Println("Starting the Go HTTP server...")
	// if rs, err := http_client.Get("http://localhost:8081/"); err != nil {
	// 	println(" > Error connecting to the server:", err)
	// } else {
	// 	println(" > Successfully connected to the server %s", &rs.Body)
	// }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// result := createApiPackage("12345", "Hello, World!")
		// result.Data = fmt.Sprintf("Hello, %s! You are running this application on %s\n", name, os_name)
		// result.CheckSum = sha1sum(result.Data)

		// //sha1sum(fmt.Sprintf("Hello, %s! You are running this application on %s\n", name, os_name))
		// //fmt.Fprintf(w, "Hello, %s! You are running this application on %s\n", name, os_name)
		// //fmt.Fprintf(w, "%+v", result)
		// w.Header().Set("Content-Type", "application/json")
		// w.Header().Set("X-Request-ID", result.RequestId)
		// w.Header().Set("X-API-Version", result.ApiVersion)
		// fmt.Println(" > toString():", result.ToString())
		// resultJson, _ := json.Marshal(result)
		// fmt.Fprintf(w, "%s", string(resultJson))
		// json.NewEncoder(w).Encode(result)
		// json.NewDecoder(r.Body).Decode(&result)
		// fmt.Println("Received a request from", r.RemoteAddr)
	})

	http.HandleFunc("/users/add", controllers.AddUser)
	http.HandleFunc("GET /users/list/{s}", controllers.ListUsers)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Application is healthy!\n")
		// bodyBytes, _ := io.ReadAll(r.Body)
		// fmt.Println(" > Health check request body:", string(bodyBytes))
	})

	http.HandleFunc("/login", controllers.Login)

	http.HandleFunc("/upload", controllers.UploadFile)

	http.HandleFunc("/userfiles", controllers.ListDir)
	// Start the HTTP server

	fmt.Println("Starting HTTP server...")
	go http.ListenAndServe(":8080", nil)
	// _http := http.Server{
	// 	Addr:    ":8080",
	// 	Handler: nil,
	// }
	// _http.ListenAndServe()
	// _http.Close()

	fmt.Println("Server is running on port 8080")
	var cmd string
	fmt.Print("type Q for quit: ")
	for {
		fmt.Scanln(&cmd)
		if cmd == "Q" || cmd == "q" {
			fmt.Println("Shutting down the server...")
			break
		}
	}

}

// func createApiPackage(requestId string, messageText string) models.ApiPackage {
// 	return models.ApiPackage{
// 		RequestId:   requestId,
// 		TimeAt:      time.Now(),
// 		ApiVersion:  "1.0",
// 		IsEncrypted: false,
// 		MessageText: messageText,
// 		DataType:    "text/plain",
// 	}
// }

// Define a local type that aliases models.ApiPackage

func sha1sum(s string) string {
	// This function is a placeholder for SHA-1 checksum calculation.
	// In a real application, you would implement the actual logic here.
	h := sha1.New()
	io.WriteString(h, s)
	//fmt.Printf("% x", h.Sum(nil))
	return string(h.Sum(nil))
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse multipart form with a max memory of 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form data"+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		http.Error(w, "Error retrieving the file"+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst, err := os.Create("./uploaded_" + handler.Filename)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error writing the file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success","filename":"%s"}`, handler.Filename)
}
