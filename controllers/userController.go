package controllers

import (
	"encoding/json"
	"fmt"
	"go-sample/models"
	"go-sample/services"
	"net/http"
	"strings"
)

func AddUser(w http.ResponseWriter, r *http.Request) {
	var _apiPkg ApiPackage

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var err error
	_apiPkg, err = ApiPackageFromRequest(r)
	// bodyBytes, _ := io.ReadAll(r.Body)
	// fmt.Println(" > Health check request body:", string(bodyBytes))
	//json.NewEncoder(w).Encode(_apiPkg)
	// decoder := json.NewDecoder(r.Body)
	// err := decoder.Decode(&_apiPkg)

	//fmt.Println(" > Decoded API Package:", _apiPkg.Data)
	if err != nil {
		http.Error(w, "Invalid request body"+err.Error(), http.StatusBadRequest)
		//return
	}

	newStr := strings.ReplaceAll(_apiPkg.Data, "'", "\"")
	//println(" > New user data string:", newStr)
	var newUser services.User
	err = json.Unmarshal([]byte(newStr), &newUser)
	if err != nil {
		http.Error(w, "Invalid user data"+err.Error(), http.StatusBadRequest)
		return
	}

	var result int64
	result, err = services.AddUser(&newUser)
	if err != nil {
		http.Error(w, "Failed to add user to database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_apiPkg = _apiPkg.Create("User added successfully")
	_apiPkg.StatusCode = http.StatusCreated
	_apiPkg.Data = models.ToJsonString(result)
	// w.WriteHeader(http.StatusCreated)
	// responseBytes, err := json.Marshal(_apiPkg)
	// if err != nil {
	// 	http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
	// 	return
	// }
	//w.Write(responseBytes)
	// Best practice to check the error from Write
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusCreated)
	// w.Write(responseBytes)
	// or simply:
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(_apiPkg)

	//fmt.Fprintf(w, "%s", responseBytes)

	// Now newUser is populated from the request body
	// ... further processing ...
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s := r.PathValue("s")
	users, err := services.GetUsers(s)
	if err != nil {
		ResponseWriterWithData(w, "Failed to retrieve users: "+err.Error(), http.StatusInternalServerError, nil)
		return
	}

	ResponseWriterWithData(w, "List of users", http.StatusOK, users)
}

func AuthenticateUserWithToken(r *http.Request) (*services.User, error) {
	bearerToken := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
	if bearerToken == "" {
		return nil, fmt.Errorf("missing or invalid Authorization header")
	}
	user, err := services.AuthenticateUserWithToken(bearerToken)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func tryLogin(rq *http.Request) (*services.User, error) {
	fmt.Println(" > Login request received from", rq.RemoteAddr)
	user, _ := AuthenticateUserWithToken(rq)
	if user != nil {
		return user, nil
	}
	apiPck, err := ApiPackageFromRequest(rq)
	if err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}
	var userCred services.LoginCredentials
	err = apiPck.ExtractDataFromApiPackage(&userCred)
	if err != nil {
		return nil, fmt.Errorf("invalid login credentials data: %v", err)
	}
	user, err = services.AuthenticateUser(userCred.EMail, userCred.Password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	return user, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user *services.User
	fmt.Println(" > Login request received from", r.RemoteAddr)
	user, _ = AuthenticateUserWithToken(r)
	if user != nil {
		//fmt.Println(" > User already authenticated:", user.AuthToken)
		ResponseWriterWithData(w, "User already authenticated", http.StatusOK, user.GetTokenInfo())
		return
	}
	apiPck, err := ApiPackageFromRequest(r)
	if err != nil {
		ResponseWriterWithData(w, "Invalid request body:"+err.Error(), http.StatusBadRequest, nil)
		return
	}

	var userCred services.LoginCredentials
	err = apiPck.ExtractDataFromApiPackage(&userCred)
	if err != nil {
		ResponseWriterWithData(w, "Invalid user data:"+err.Error(), http.StatusBadRequest, nil)
		return
	}
	fmt.Println(" > Authenticating user:", userCred.EMail, "password:", userCred.Password)
	user, err = services.AuthenticateUser(userCred.EMail, userCred.Password)
	if err != nil {
		//log.Println(" > Authentication failed for user:", userCred.EMail, err)
		ResponseWriterWithData(w, "Authentication failed:"+err.Error(), http.StatusUnauthorized, nil)
		return
	}

	ResponseWriterWithData(w, "Login successful", http.StatusOK, user.GetTokenInfo())

}
