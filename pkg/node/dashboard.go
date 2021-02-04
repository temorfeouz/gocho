package node

import (
	"archive/tar"
	"bufio"
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
		filename := buildFullFilePath(conf.ShareDirectory, relativePath)

		zr, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		wrt := tar.NewWriter(zr)
		//zr.SetConcurrency(1048576, runtime.NumCPU())
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		defer zr.Close()
		defer wrt.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="`+getFilename(relativePath)+` .tar.gz"`))

		fi, err := os.Stat(filename)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			// walk through every file in the folder
			filepath.Walk(filename, func(file string, fi os.FileInfo, err error) error {
				// generate tar header
				header, err := tar.FileInfoHeader(fi, file)
				if err != nil {
					return err
				}
				// must provide real name
				// (see https://golang.org/src/archive/tar/common.go?#L626)
				header.Name = filepath.ToSlash(file)

				// write header
				if err := wrt.WriteHeader(header); err != nil {
					return err
				}
				// if not a dir, write file content
				if !fi.IsDir() {
					data, err := os.Open(file)
					if err != nil {
						return err
					}
					if _, err := io.Copy(wrt, data); err != nil {
						return err
					}
				}
				return nil
			})
		case mode.IsRegular():
			if f, err := os.OpenFile(filename, os.O_RDONLY, 0755); err == nil {
				fi, err2 := f.Stat()
				if err2 != nil {
					fmt.Errorf("%s", err)
					return
				}
				log.Printf("processing %s file size %v", filename, ByteCountSI(fi.Size()))
				fr := bufio.NewReader(f)
				fp, _ := filepath.Abs(filename)
				header, _ := tar.FileInfoHeader(fi, fp)
				header.Name = filepath.ToSlash(filename)
				wrt.WriteHeader(header)
				if _, err := io.Copy(wrt, fr); err != nil {
					fmt.Errorf("err io.copy %s - %s", filename, err)
				}
				defer f.Close()
			}
		}
	}
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
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

	// create file for saving
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", fullFilePath, filename), os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, file)

	return err
}

func dashboardServe(conf *config.Config, nodeList *list.List) {
	dashboardMux := http.NewServeMux()
	dashboardMux.Handle("/", http.FileServer(assets.AssetFS()))
	dashboardMux.HandleFunc("/api/config", configHandler(conf))
	dashboardMux.HandleFunc("/api/nodes", nodesHandler(nodeList))

	// We don't want the dashboard to be public
	// address := "localhost"
	// if conf.Debug {
	//address := "0.0.0.0"
	// }
	fmt.Printf("Starting dashboard at %s:%s\n", conf.Interface, conf.LocalPort)
	http.ListenAndServe(fmt.Sprintf("%s:%s", conf.Interface, conf.LocalPort), dashboardMux)
}
