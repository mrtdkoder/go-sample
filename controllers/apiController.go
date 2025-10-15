package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Add your model structs here

type ApiPackage struct {
	RequestId   string    `json:"requestId"`
	StatusCode  int       `json:"statusCode"`
	TimeAt      time.Time `json:"time"`
	CheckSum    string    `json:"hashCode"`
	ApiVersion  string    `json:"version"`
	IsEncrypted bool      `json:"encrypted"`
	MessageText string    `json:"messageText"`
	DataType    string    `json:"dataType"`
	Data        string    `json:"data"`
}

func (ap ApiPackage) ToString() string {
	return fmt.Sprintf("RequestId: %s, StatusCode: %d, Time: %s, CheckSum: %s, ApiVersion: %s, IsEncrypted: %t, MessageText: %s, DataType: %s, Data: %s",
		ap.RequestId, ap.StatusCode, ap.TimeAt.Format(time.RFC3339), ap.CheckSum, ap.ApiVersion, ap.IsEncrypted, ap.MessageText, ap.DataType, ap.Data)
}

func (ap *ApiPackage) ToJson() ([]byte, error) {
	return json.Marshal(ap)
}

// base64 encoded data
func (ap *ApiPackage) SetData(data interface{}) {
	ap.Data = ToJsonString(data)
}

// base64 decoded data
func (ap *ApiPackage) GetData(data interface{}) error {
	//fmt.Println(ap.Data)
	err := json.Unmarshal([]byte(ap.Data), data)
	//fmt.Println(data)
	return err
}

func (ap *ApiPackage) DataTo() (any, error) {
	fmt.Println(ap.Data)
	var data any
	print(&data)
	print(data)
	err := json.Unmarshal([]byte(ap.Data), data)
	fmt.Println(data)
	return data, err
}

// func (ap *ApiPackage) GetDataTo(v any, decode bool) error {
// 	result, err := ap.GetData(decode)
// 	*v = &result
// 	return err
// }

func (ap *ApiPackage) Create(msg string) ApiPackage {
	return ApiPackage{
		TimeAt:      time.Now(),
		ApiVersion:  "1.0",
		IsEncrypted: false,
		MessageText: msg,
		DataType:    "application/json",
	}
}

func (ap *ApiPackage) CreateFromStr(str string) (ApiPackage, error) {
	err := json.Unmarshal([]byte(str), &ap)
	if err != nil {
		return *ap, fmt.Errorf("error unmarshaling json to apipackage: %v", err)
	}
	return *ap, nil
}

func (ap *ApiPackage) ExtractDataFromApiPackage(v any) error {
	fmt.Println(ap.Data)
	return json.Unmarshal([]byte(ap.Data), &v)
}

func ToJsonString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

// consists of Api related functions

func CreateNewApiPackage(msg string) ApiPackage {
	return ApiPackage{
		TimeAt:      time.Now(),
		ApiVersion:  "1.0",
		IsEncrypted: false,
		MessageText: msg,
		DataType:    "application/json",
	}
}

func ResponseWrite(w http.ResponseWriter, res *ApiPackage) {
	w.Header().Set("Content-Type", "application/json")
	responseBytes, err := res.ToJson()
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(res.StatusCode)
	w.Write(responseBytes)
}

func ResponseWriterWithData(w http.ResponseWriter, msg string, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	_apiPkg := ApiPackage{}
	_apiPkg = _apiPkg.Create(msg)
	_apiPkg.StatusCode = status
	//_apiPkg.Data = models.ToJsonString(data)
	_apiPkg.SetData(data)
	responseBytes, err := _apiPkg.ToJson()
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Write(responseBytes)
}

func ApiPackageFromRequest(r *http.Request) (ApiPackage, error) {
	var _apiPkg ApiPackage
	//fmt.Print("body:" + ToJsonString(r.Body))
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&_apiPkg)
	if err != nil {
		return _apiPkg, err
	}
	// decode can be done here ..._apiPkg.DecodeData()
	return _apiPkg, nil
}
