package services

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

type FileSystemService struct {
	BasePath string
	OwnerID  int
	Capacity int64
	Used     int64
}

type FileInfo struct {
	FID        int       `json:"fid"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	OwnerId    int       `json:"ownerid"`
	IsDir      bool      `json:"isdir"`
	ParentId   int       `json:"parentid"`
	UploadedAt time.Time `json:"uploadedat"`
	UpdatedAt  time.Time `json:"updatedat"`
	HashCode   string    `json:"hashcode"`
}

type FileContent struct {
	FID         int    `json:"fid,omitempty"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	ChunkId     int    `json:"chunkId"`
	TotalChunks int    `json:"totalChunks"`
	Content     string `json:"content,omitempty"`
}

/*			functions 			*/

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func (fs *FileSystemService) findPartFiles(filename string, dir string) ([]string, error) {
	dir = filepath.Clean(dir)
	//fmt.Printf("dir:%s - filename:%s", dir, filename)
	partFiles, err := filepath.Glob(filepath.Join(dir, filename) + ".part*")
	if err != nil {
		return nil, err
	}
	return partFiles, nil
}

func (fs *FileSystemService) deleteFiles(files []string) error {
	for _, fname := range files {
		err := os.Remove(fname)
		if err != nil {
			return err
		}
		//fmt.Print(fname + " deleted")
	}
	return nil
}

// MergeFiles merges multiple files into a single output file
func (fs *FileSystemService) mergeFiles(inputFiles []string, outputFile string) error {
	// Create the output file
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer out.Close()

	// sort file names
	sort.Strings(inputFiles)
	// Process each input file
	for i, inputFile := range inputFiles {
		// Open the input file
		in, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open input file %s: %v", inputFile, err)
		}

		// Copy the file contents to the output file
		_, err = io.Copy(out, in)
		in.Close() // Close immediately after copying

		if err != nil {
			return fmt.Errorf("failed to copy from %s: %v", inputFile, err)
		}

		//fmt.Printf("Copied %d bytes from %s\n", bytesCopied, inputFile)

		// Add a separator between files (optional)
		// if i < len(inputFiles)-1 {
		// 	_, err = out.WriteString("\n\n") // Two newlines as separator
		// 	if err != nil {
		// 		return fmt.Errorf("failed to write separator: %v", err)
		// 	}
		// }
		i++
	}
	return nil
}

func NewFileSystemService(basePath string, ownerID int) *FileSystemService {
	return &FileSystemService{
		BasePath: basePath,
		OwnerID:  ownerID,
	}
}

func (fs *FileSystemService) AddNewFileToDB(fc *FileContent) (int64, error) {
	//fmt.Printf("AddNewFileToDB>name:%s - path:%s - content: %s \n", fc.Name, fc.Path, fc.Content)
	fullPath := path.Join(fs.BasePath, fc.Path)
	if len(fc.Content) == 0 {
		return 0, fmt.Errorf("empty file content")
	}
	println("fullpath:" + fullPath)
	//fmt.Printf("chk: %d/%d - fullpath:%s\n", fc.ChunkId, fc.TotalChunks, fullPath)
	// check chunks if it is the last one save it to database if not just save it to the filesystem with an extention .part123
	var _partFilesCount int
	if fc.TotalChunks > 1 {
		xfullPath := fmt.Sprintf("%s.part%03d", fullPath, fc.ChunkId)
		e := fs.WriteFile(xfullPath, fc.Content, true, true)
		if e != nil {
			return 0, fmt.Errorf("error->anftodb->: %v", e)
		}
		_dir, _file := filepath.Split(fullPath)
		_partfiles, err := fs.findPartFiles(_file, _dir)
		if err != nil {
			return 0, fmt.Errorf("error->findPartFiles: %v", err)
		}
		_partFilesCount = len(_partfiles)
		if _partFilesCount == fc.TotalChunks { //if fc.ChunkId == fc.TotalChunks {
			//mergeFiles()
			err := fs.mergeFiles(_partfiles, fullPath)
			if err != nil {
				return 0, fmt.Errorf("error->mergeFiles: %v", err)
			}
			err = fs.deleteFiles(_partfiles)
			if err != nil {
				return 0, fmt.Errorf("error->deleteFiles: %v", err)
			}
		}
	} else {
		err := fs.WriteFile(fullPath, fc.Content, true, true)
		if err != nil {
			return 0, fmt.Errorf("couldn't write this file: %s", err.Error())
		}
		_partFilesCount = 1
	}

	if _partFilesCount != fc.TotalChunks {
		return 0, nil
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("error->os.stat: %v", err)
	}
	fcSize := info.Size()
	fcIsDir := info.IsDir()
	hashCode := HashPassword(fc.Content)

	var FID int
	row := OpenDB().QueryRow("select FID from Files where FullPath=? ", fullPath)
	e := row.Scan(&FID)
	if e != nil {
		FID = 0
	}
	var insertedId int64
	if FID > 0 {
		result, err := OpenDB().Exec("update files set Name=?, FullPath=?, Size=?, IsDir=?, ParentId=?, HashCode=?, UpdatedAt=? where FID=? and OwnerId=?",
			fc.Name, fullPath, fcSize, fcIsDir, 0, hashCode, time.Now(), FID, fs.OwnerID)
		if err != nil {
			return 0, fmt.Errorf("[editing-file-record]: %s", err.Error())
		}
		_, err = result.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("error in last editing id: %s", err.Error())
		}
		insertedId = int64(FID)
	} else {
		result, err := OpenDB().Exec("insert into files (Name, FullPath, Size, OwnerId, IsDir, ParentId, HashCode, UploadedAt, UpdatedAt) values (?,?,?,?,?,?,?,?,?)",
			fc.Name, fullPath, fcSize, fs.OwnerID, fcIsDir, 0, hashCode, time.Now(), time.Now())
		if err != nil {
			return 0, fmt.Errorf("[insert-file-record]: %s", err.Error())
		}
		insertedId, err = result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("error in last inserted id: %s", err.Error())
		}
	}

	return insertedId, nil

}

func (fs *FileSystemService) MakeDir(path string, all bool) error {
	//parentDir := filepath.Dir(path)
	path = filepath.ToSlash(path)
	parentDir, lastone := filepath.Split(path)
	fmt.Printf("parent: %s - last: %s", parentDir, lastone)
	//fmt.Println("parentDir>" + parentDir)
	var _fid int
	_row := OpenDB().QueryRow("select FID from Files where FullPath=? ", path)
	e := _row.Scan(&_fid)
	if e != nil {
		_fid = 0
	}

	//if !directoryExists(path) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	//}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error-makedir->os.stat: %v", err)
	}
	fcSize := info.Size()
	fcIsDir := info.IsDir()
	var _lastFID int64
	var _err error
	if _fid > 0 {
		result, e := OpenDB().Exec("update files set Name=?, FullPath=?, Size=?, OwnerId=?, IsDir=?, UpdatedAt=? where FID=? ",
			lastone, path, fcSize, fs.OwnerID, fcIsDir, time.Now(), _fid)
		if e != nil {
			return err
		}
		_lastFID, _err = result.LastInsertId()
	} else {
		result, e := OpenDB().Exec("insert into files (Name, FullPath, Size, OwnerId, IsDir, UploadedAt, UpdatedAt) values (?,?,?,?,?,?,?)",
			lastone, path, fcSize, fs.OwnerID, fcIsDir, time.Now(), time.Now())
		if e != nil {
			return err
		}
		_lastFID, _err = result.LastInsertId()
	}

	if _err != nil {
		return fmt.Errorf("%s could not recorded in db", path)
	}
	if _lastFID == 0 {
		return fmt.Errorf("no record has been affacted with %s in db", path)
	}
	return nil
}

func (fs *FileSystemService) WriteFile(filePath string, content string, decode bool, overwrite bool) error {
	fullPath := filepath.ToSlash(filepath.Clean(filePath)) //filepath.Join(fs.BasePath, filePath)
	var newContent []byte
	if decode {
		decodedData, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return err
		}
		newContent = decodedData
	}
	parentDir := filepath.Dir(fullPath)
	//fmt.Println("parentDir>" + parentDir)
	//if !directoryExists(parentDir) {
	//os.MkdirAll(parentDir, 0755)
	err := fs.MakeDir(parentDir, true)
	if err != nil {
		return fmt.Errorf("error->makedir-> %v", err)
	}
	//}
	if !fileExists(fullPath) {
		return os.WriteFile(fullPath, newContent, 0644)
	} else if overwrite {
		return os.WriteFile(fullPath, newContent, 0644)
	} else {
		return fmt.Errorf("file already exists")
	}
}

func (fs *FileSystemService) WriteFileByFC(fc *FileContent, overwrite bool) error {
	//fmt.Printf("WriteFile>name:%s - path:%s - content: %s \n", fc.Name, fc.Path, fc.Content)
	fullPath := filepath.Join(fs.BasePath, fc.Path)
	//filepath.Dir()
	_content, err := base64.StdEncoding.DecodeString(fc.Content)
	if err != nil {
		return err
	}
	//_content := []byte(fc.Content)
	// fmt.Println("writing file")
	// fmt.Printf("fullpath: %s \n", fullPath)
	// fmt.Printf("dir: %s \n", filepath.Dir(fullPath))
	// fmt.Printf("base: %s \n", filepath.Base(fullPath))
	//fmt.Printf("abs: %s \n", filepath.Abs(fullPath))
	//fmt.Println(_content)
	parentDir := filepath.Dir(fullPath)
	//fmt.Println("parentDir>" + parentDir)
	if !directoryExists(parentDir) {
		os.MkdirAll(parentDir, 0755)
	}

	if fc.TotalChunks > 1 {
		partFile := fmt.Sprint(fullPath, ".prt", fc.ChunkId)
		if !fileExists(partFile) {
			return os.WriteFile(partFile, _content, 0644)
		} else if overwrite {
			return os.WriteFile(partFile, _content, 0644)
		}
	} else {
		if !fileExists(fullPath) {
			return os.WriteFile(fullPath, _content, 0644)
		} else if overwrite {
			return os.WriteFile(fullPath, _content, 0644)
		}
	}

	return fmt.Errorf("file already exists")
}

func (fs *FileSystemService) DeleteFile(fileDetail *FileInfo) error {
	fullPath := filepath.Join(fs.BasePath, fileDetail.Path, fileDetail.Name)
	return os.Remove(fullPath)
}

func (fs *FileSystemService) DeleteAll(fileDetail *FileInfo) error {
	fullPath := filepath.Join(fs.BasePath, fileDetail.Path, fileDetail.Name)
	return os.RemoveAll(fullPath)
}

func (fs *FileSystemService) GetFileContent(fileDetail *FileInfo) ([]byte, error) {
	fullPath := filepath.Join(fs.BasePath, fileDetail.Path, fileDetail.Name)
	_content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return []byte(base64.StdEncoding.EncodeToString(_content)), nil
}

// func (fs *FileSystemService) MakeDir(dirPath string) error {
// 	os.ReadDir()
// 	return os.MkdirAll(dirPath, 0750)
// }

func (fs *FileSystemService) ListDir(dir string, filter string) ([]FileInfo, error) {
	dirPath := filepath.Join(fs.BasePath, dir)
	// Read directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	var results []FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		fd := FileInfo{
			Name:       info.Name(),
			Path:       filepath.Join(dirPath, info.Name()),
			OwnerId:    0,
			Size:       info.Size(),
			IsDir:      info.IsDir(),
			UploadedAt: info.ModTime(),
		}
		results = append(results, fd)
	}

	return results, nil
}

// func (fs *FileService) ListFiles() []*FileService       {}
// func (fs *FileService) MakeDir(dirPath string) bool     {}
// func (fs *FileService) DeleteDir(dirPath string) bool   {}
