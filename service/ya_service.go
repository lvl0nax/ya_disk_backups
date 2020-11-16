package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	yaUrl             = "https://cloud-api.yandex.net/v1/disk"
	applicationFolder = "Приложения" // i have a russian interface
)

type Resource struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Created    string `json:"created"`
	ResourceId string `json:"resource_id"`
	Type       string `json:"type"`
	MimeType   string `json:"mime_type"`
	Embedded   struct {
		Items []Resource `json:"items"`
		Path  string     `json:"path"`
	} `json:"_embedded"`
}

type YaDiskApi interface {
	GetResource(path string) (*Resource, error)
	CreateFolder(newFolder string) error
	UploadFile(localPath, remotePath string) error
	RemoveOldBackups(path string, num int) error
}

type YaService struct {
	token   string
	appName string
	YaDiskApi
}

func NewYaService(token, appName string) *YaService {
	return &YaService{
		token:   token,
		appName: appName,
	}
}

func (s *YaService) apiRequest(path, method string) (*http.Response, error) {
	client := http.Client{}
	url := fmt.Sprintf("%s/%s", yaUrl, path)
	// fmt.Println("Full URL: ", url)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", s.token))
	req.Close = true

	return client.Do(req)
}

func (s *YaService) GetResource(path string) (*Resource, error) {
	res, err := s.apiRequest(fmt.Sprintf("resources?sort=-created&path=%s/%s/%s", applicationFolder, s.appName, path), "GET")
	fmt.Println("Status: ", res.Status)
	if err != nil {
		return nil, err
	}

	var resource *Resource
	err = json.NewDecoder(res.Body).Decode(&resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *YaService) CreateFolder(newFolder string) error {
	path := fmt.Sprintf("resources?path=%s/%s/%s", applicationFolder, s.appName, newFolder)
	res, err := s.apiRequest(path, "PUT")
	if res != nil {
		defer res.Body.Close()
	}

	// fmt.Println("Status: ", res.Status)
	if res.StatusCode == 409 {
		fmt.Println("Folder already exists")
	}

	return err
}

func (s *YaService) DeleteResource(path string) error {
	url := fmt.Sprintf("resources?permanently=true&path=%s", path)
	res, err := s.apiRequest(url, "DELETE")

	if res != nil {
		defer res.Body.Close()
	}

	return err
}

func (s *YaService) UploadFile(localPath, remotePath string) error {
	url, method, err := s.getUploadUrl(remotePath)
	if err != nil {
		return err
	}

	// read local file
	data, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer data.Close()

	// fmt.Printf("METHOD=%s\n", method)
	// fmt.Printf("URL=%s\n", url)
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", s.token))
	req.Close = true
	client := &http.Client{}
	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return err
	}

	return nil
}

type resultUploadUrl struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

func (s *YaService) getUploadUrl(path string) (string, string, error) {
	url := fmt.Sprintf("resources/upload?overwrite=true&path=%s/%s/%s", applicationFolder, s.appName, path)
	res, err := s.apiRequest(url, "GET")
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return "", "", err
	}

	// body, _ := ioutil.ReadAll(res.Body)
	// fmt.Println(string(body))

	var result *resultUploadUrl
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return "", "", err
	}

	return result.Href, result.Method, nil
}

func (s *YaService) RemoveOldBackups(path string, num int) error {
	res, err := s.GetResource(path)
	if err != nil {
		return err
	}

	for i, file := range res.Embedded.Items {
		fmt.Printf("%v %s %s\n", i, file.Name, file.Path)
		if i > num-1 {
			err = s.DeleteResource(file.Path)
			if err != nil {
				fmt.Printf("YaDisk file delete failed. Error: %s\n", err)
			}
		}
	}

	return nil
}
