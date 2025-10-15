package controllers

import (
	"fmt"
	"go-sample/services"
	"net/http"
	"strings"
)

var fss = services.FileSystemService{
	BasePath: "./userfiles/",
}

func checkAndPrepareRequest(w http.ResponseWriter, r *http.Request, method string, fileSS *services.FileSystemService) (ApiPackage, error) {
	if r.Method != method {
		_tmp := CreateNewApiPackage("please login")
		_tmp.StatusCode = http.StatusMethodNotAllowed
		ResponseWrite(w, &_tmp)
		return _tmp, fmt.Errorf("method not allowed")
	}
	usr, err := AuthenticateUserWithToken(r)
	if usr == nil {
		_tmp := CreateNewApiPackage("cannot retrive user data")
		_tmp.StatusCode = http.StatusNetworkAuthenticationRequired
		ResponseWrite(w, &_tmp)
		//w.Write([]byte("please login"))
		return _tmp, fmt.Errorf("cannot retrive user data: %v", err)
	}

	_apiPkg, err := ApiPackageFromRequest(r)
	if err != nil {
		ap := CreateNewApiPackage("error:" + err.Error())
		ap.StatusCode = 400
		ResponseWrite(w, &ap)
		return _apiPkg, err
	}

	fileSS.BasePath = "./userfiles/" + usr.HomeDir
	fileSS.OwnerID = usr.UID

	return _apiPkg, nil
}

func UploadFile(w http.ResponseWriter, r *http.Request) {
	_apiPkg, err := checkAndPrepareRequest(w, r, http.MethodPost, &fss)
	if err != nil {
		return
	}

	newStr := strings.ReplaceAll(_apiPkg.Data, "'", "\"")
	_apiPkg.Data = newStr
	var aa services.FileContent
	err = _apiPkg.GetData(&aa)
	if err != nil {
		ap := _apiPkg.Create("error getting data from request:" + err.Error())
		ap.StatusCode = 403
		ResponseWrite(w, &ap)
		return
	}

	//var fss services.FileSystemService
	// fmt.Println(aa)
	// fmt.Printf("upload>name:%s - path:%s - content: %s \n", aa.Name, aa.Path, aa.Content)
	if aa.ChunkId < 1 {
		aa.ChunkId = 1
	}
	if aa.TotalChunks < 1 {
		aa.TotalChunks = 1
	}
	fid, err := fss.AddNewFileToDB(&aa)
	if err != nil {
		ap := _apiPkg.Create("error saving request into db:" + err.Error())
		ap.StatusCode = 401
		ResponseWrite(w, &ap)
		return
	}

	ap := _apiPkg.Create(fmt.Sprintf("new file has been recorded to db successfully. fid:%d", (fid)))
	ap.StatusCode = 200
	ResponseWrite(w, &ap)
}

// list files and directories in given path
func ListDir(w http.ResponseWriter, r *http.Request) {
	_apiPkg, err := checkAndPrepareRequest(w, r, http.MethodGet, &fss)
	if err != nil {
		return
	}

	newStr := strings.ReplaceAll(_apiPkg.Data, "'", "\"")
	_apiPkg.Data = newStr

	var args map[string]string                     //make(map[string]string)
	err = _apiPkg.ExtractDataFromApiPackage(&args) //.DataTo() //_apiPkg.GetData(args)
	if err != nil {
		ap := _apiPkg.Create("error getting data from apipackage:" + err.Error())
		ap.StatusCode = 403
		ResponseWrite(w, &ap)
		return
	}

	//fmt.Printf("args: %s - %s", args["dir"], args["filter"])

	finfos, err := fss.ListDir(args["dir"], args["filter"])
	if err != nil {
		ap := _apiPkg.Create("error getting files info on directories:" + err.Error())
		ap.StatusCode = 403
		ResponseWrite(w, &ap)
		return
	}

	ap := _apiPkg.Create("files info data")
	ap.StatusCode = 200
	ap.DataType = "json/fileinfo[]"
	ap.SetData(finfos)
	ResponseWrite(w, &ap)

}
