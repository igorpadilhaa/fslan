package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"

	"github.com/igorpadilhaa/fslan/fs"
	"github.com/igorpadilhaa/fslan/lookup"
)

type ClientHttp struct {
	hosts lookup.HostTable
	fs    fs.DynamicFs
}

func NewClientHTTP(hostname string, port int) (Client, error) {
	hosts, err := lookup.Start(lookup.ServerDesc{
		Name: hostname,
		Port: port,
	})

	if err != nil {
		return nil, err
	}

	client := ClientHttp{
		hosts: hosts,
		fs:    fs.NewDynamicFs(),
	}

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), fs.HttpHandler(client.fs))
		if err != nil {
			fmt.Println(err)
		}
	}()

	return &client, nil
}

func (client *ClientHttp) Hosts() []string {
	return client.hosts.List()
}

func (client *ClientHttp) GetFile(host, filename string) error {
	serverAddr, err := client.hosts.Resolve(host)
	if err != nil {
		return err
	}
	return downloadFile(filename, "http://"+serverAddr.String()+"/f/"+filename)
}

func (client *ClientHttp) ListFiles(host string) ([]RemoteFile, error) {
	serverAddr, err := client.hosts.Resolve(host)
	if err != nil {
		return nil, err
	}
	return fetchFileList(serverAddr)
}

func (client *ClientHttp) Shared() []LocalFile {
	var files []LocalFile
	for _, file := range client.fs.Files() {
		local := LocalFile{
			Name: file.Name,
			Path: file.Path,
			Mime: file.Mime,
		}
		files = append(files, local)
	}
	return files
}

func (client *ClientHttp) ShareFile(name, path string) {
	client.fs.Add(name, path)
}

func fetchFileList(serverAddr net.Addr) ([]RemoteFile, error) {
	res, err := http.Get("http://" + serverAddr.String() + "/files")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rawData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var files []RemoteFile
	if err := json.Unmarshal(rawData, &files); err != nil {
		return nil, err
	}
	return files, nil
}

func downloadFile(filename, fileUrl string) error {
	res, err := http.Get(fileUrl)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return errors.New("file not found")

	} else if res.StatusCode != http.StatusOK {
		return errors.New("failed to download file")
	}

	mimeType := res.Header.Get("Content-Type")
	extensions, err := mime.ExtensionsByType(mimeType)
	if err != nil {
		return fmt.Errorf("unsupported file type: %w", err)
	}

	fmt.Println("ext: ", extensions[0])
	filename += extensions[0]
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	return err
}
