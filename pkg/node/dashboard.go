package node

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	gzip "github.com/klauspost/pgzip"

	"github.com/temorfeouz/gocho/assets"
	"github.com/temorfeouz/gocho/pkg/config"
)

const (
	tmpFolder = ".tmp"
)

func configHandler(conf *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(conf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	}
}

func nodesHandler(nodeList *list.List) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		nodes := make([]*NodeInfo, 0)
		for el := nodeList.Front(); el != nil; el = el.Next() {
			tmp := el.Value.(*NodeInfo)
			nodes = append(nodes, tmp)
		}

		data, err := json.Marshal(nodes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	}
}

func fileUpload(conf *config.Config) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(2 << 20)
		if err != nil {
			panic(err)
		}

		// file, _, err := r.FormFile("files[]")
		// if err != nil {
		// 	w.Write([]byte(fmt.Sprintf("Error on upload file -> %s", err)))
		// 	return
		// }
		// defer file.Close()
		if r.MultipartForm != nil {
			for _, tmp := range r.MultipartForm.File {
				for _, v := range tmp {

					file, err := v.Open()

					if err != nil {
						w.Write([]byte(fmt.Sprintf("Error on open file -> %s", err)))
						return
					}

					err = saveFile(file, conf.ShareDirectory+r.FormValue("dir"), v.Filename)
					if err != nil {
						w.Write([]byte(fmt.Sprintf("Error on create file -> %s", err)))
						return
					}

					file.Close()

				}
			}
		}
	}
}

func delete(conf *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			panic(err)
		}
		// basePath := strings.TrimRight(conf.ShareDirectory, "/")

		if r.FormValue("elem") == "" || strings.Contains(r.FormValue("elem"), "..") {
			w.Write([]byte("OK"))
			return
		}

		if err := os.RemoveAll(buildFullFilePath(conf.ShareDirectory, r.FormValue("elem"))); err != nil {
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func buildFullFilePath(firstPart, partPath string) string {
	basePath := strings.TrimRight(firstPart, "/")
	partPath = strings.TrimLeft(partPath, "/")

	return fmt.Sprintf("%s/%s", basePath, partPath)

}

func getFilename(path string) string {
	path = strings.TrimRight(path, "/")
	path = strings.TrimLeft(path, "/")
	tmp := strings.Split(path, "/")

	return tmp[len(tmp)-1]
}

func validatePath(p string, conf *config.Config) error {
	if p == "" {
		return errors.New("empty path")
	}
	if strings.Contains(p, "..") {
		return errors.New("Incorrect characters")
	}
	fullPath := buildFullFilePath(conf.ShareDirectory, p)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("File '%s' doesnt exists", fullPath))
	}
	return nil
}
func archive(conf *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		relativePath := r.FormValue("elem")
		if err := validatePath(relativePath, conf); err != nil {
			w.Write([]byte(fmt.Sprintf("On validate -> %s", err.Error())))
			return
		}
		filename := getFilename(relativePath)

		tmpFile := fmt.Sprintf("%s/%s.%s", getTmpFolder(conf), filename, "zip")

		// Get a Buffer to Write To
		outFile, err := os.Create(tmpFile)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("On create tmp file -> %s", err.Error())))
			return
		}
		defer outFile.Close()

		// Create a new zip archive.
		zw := gzip.NewWriter(outFile)

		// Add some files to the archive.
		err = addFiles(zw, buildFullFilePath(conf.ShareDirectory, relativePath), "")

		if err != nil {
			w.Write([]byte(fmt.Sprintf("On archiving -> %s", err.Error())))
			return
		}
		// Make sure to check the error on Close.
		err = zw.Close()
		if err != nil {
			w.Write([]byte(fmt.Sprintf("On closing archive -> %s", err.Error())))
			return
		}

		zipped, err := os.Open(tmpFile)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("On open archive -> %s", err.Error())))
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+filename+".zip")
		w.Header().Set("Content-Type", "application/zip")
		// w.Header().Set("Content-Length", zipped.)

		if _, err := io.Copy(w, zipped); err != nil {
			w.Write([]byte(fmt.Sprintf("On sending archive -> %s", err.Error())))
			return
		}

		if err := os.Remove(tmpFile); err != nil {
			w.Write([]byte(fmt.Sprintf("On removing archive -> %s", err.Error())))
			return
		}
	}
}
func addFiles(w *gzip.Writer, basePath, baseInZip string) error {
	var (
		files []os.FileInfo
		err   error
	)
	stat, err := os.Stat(basePath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		// Open the Directory
		files, err = ioutil.ReadDir(basePath)
		if err != nil {
			return err
		}
	} else {
		tmp := strings.Split(basePath, "/")
		basePath = strings.Join(tmp[0:len(tmp)-1], "/") + "/"
		files = append(files, stat)
	}

	for _, file := range files {
		fmt.Println(basePath + file.Name())
		if !file.IsDir() {
			dat, err := ioutil.ReadFile(basePath + file.Name())
			if err != nil {
				return err
			}

			// Add some files to the archive.
			// f, err := w.Create(baseInZip + file.Name())
			// if err != nil {
			// 	return err
			// }
			_, err = w.Write(dat)
			if err != nil {
				return err
			}
		} else if file.IsDir() {

			// Recurse
			newBase := basePath + file.Name() + "/"
			fmt.Println("Recursing and Adding SubDir: " + file.Name())
			fmt.Println("Recursing and Adding SubDir: " + newBase)

			addFiles(w, newBase, file.Name()+"/")
		}
	}
	return nil
}

func getTmpFolder(conf *config.Config) string {

	tmpFolderPath := buildFullFilePath(os.TempDir(), tmpFolder)
	if _, err := os.Stat(tmpFolderPath); os.IsNotExist(err) {
		os.MkdirAll(tmpFolderPath, os.ModePerm)
	}
	return tmpFolderPath
}

func saveFile(file multipart.File, basePath, filePath string) error {
	basePath = strings.TrimRight(basePath, "/")

	// get info from filepath
	tmp := strings.Split(filePath, "/")
	filename := tmp[len(tmp)-1]
	pathToFile := tmp[0 : len(tmp)-1]

	fullFilePath := fmt.Sprintf("%s/%s", basePath, strings.Join(pathToFile, "/"))
	if len(pathToFile) > 0 {
		os.MkdirAll(fullFilePath, os.ModePerm)
	}

	// here write directly
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		fmt.Sprintf("%s/%s", fullFilePath, filename),
		data, os.ModePerm,
	)
	//saving file
	// return os.Create(fmt.Sprintf("%s/%s", fullFilePath, filename))

	// fqn := fmt.Sprintf("%s/%s", basePath, filePath)
	// get file name and file path

	// os.saveFile(/)
	// return nil, nil
}

func dashboardServe(conf *config.Config, nodeList *list.List) {
	dashboardMux := http.NewServeMux()
	dashboardMux.Handle("/", http.FileServer(assets.AssetFS()))
	dashboardMux.HandleFunc("/api/config", configHandler(conf))
	dashboardMux.HandleFunc("/api/nodes", nodesHandler(nodeList))

	// We don't want the dashboard to be public
	address := "localhost"
	if conf.Debug {
		address = "0.0.0.0"
	}
	fmt.Printf("Starting dashboard at %s\n", address)
	http.ListenAndServe(fmt.Sprintf("%s:%s", address, conf.LocalPort), dashboardMux)
}
