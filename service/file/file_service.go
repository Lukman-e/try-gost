package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"

	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/gofiber/fiber/v2"
)

type FileReponse struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	Metadata  struct {
		Size int64 `json:"size"`
	} `json:"metadata"`
}

type FileService interface {
	// UploadFile func uploads file to supabase bucket.
	UploadFile(fileHeader *multipart.FileHeader) (fileURL string, err error)

	// RemoveFile func deletes a file from supabase bucket.
	RemoveFile(fileName string) (err error)

	// GetFilesList func get list of files from supabase bucket.
	GetFilesList() (files []map[string]any, err error)
}

type FileServiceImpl struct {
	PublicURL    string
	ListFilesURL string
	UploadURL    string
	DeleteURL    string
	Token        string
}

var (
	fileService     *FileServiceImpl
	fileServiceOnce sync.Once
)

func NewFileService() FileService {
	fileServiceOnce.Do(func() {
		config := env.Configuration()
		baseURL := config.BucketURL + "/storage/v1/object/"
		fileService = &FileServiceImpl{
			PublicURL:    baseURL + "public/" + config.BucketName + "/",
			ListFilesURL: baseURL + "list/" + config.BucketName,
			UploadURL:    baseURL + config.BucketName + "/",
			DeleteURL:    baseURL + config.BucketName,
			Token:        config.BucketToken,
		}
	})
	return fileService
}

func (c FileServiceImpl) UploadFile(fileHeader *multipart.FileHeader) (fileURL string, err error) {
	fileName := fileHeader.Filename
	file, headerErr := fileHeader.Open()
	if headerErr != nil {
		return "", headerErr
	}
	defer file.Close()

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	fileField, formErr := writer.CreateFormFile("file", fileName)
	if formErr != nil {
		return "", formErr
	}
	_, copyErr := io.Copy(fileField, file)
	if copyErr != nil {
		return "", copyErr
	}
	writer.Close()
	url := c.UploadURL + fileName
	request, newReqErr := http.NewRequest(http.MethodPost, url, requestBody)
	if newReqErr != nil {
		return "", newReqErr
	}

	request.Header.Set(fiber.HeaderAuthorization, "Bearer "+c.Token)
	request.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())
	FileServiceImpl := &http.Client{}
	respUpload, doErr := FileServiceImpl.Do(request)
	if doErr != nil {
		return "", doErr
	}
	defer respUpload.Body.Close()

	if respUpload.StatusCode != http.StatusOK {
		return "", responseErrHandler(respUpload)
	}

	link := c.PublicURL + fileName
	reqTestGet, reqErr := http.NewRequest(http.MethodGet, link, nil)
	if reqErr != nil {
		return "", reqErr
	}
	respTestGet, doErr := FileServiceImpl.Do(reqTestGet)
	if doErr != nil {
		return "", doErr
	}
	defer respTestGet.Body.Close()
	if respTestGet.StatusCode != http.StatusOK {
		return "", responseErrHandler(respTestGet)
	}
	return link, nil
}

func (c FileServiceImpl) RemoveFile(fileName string) (err error) {
	body := map[string]interface{}{
		"prefixes": fileName,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodDelete, c.DeleteURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	request.Header.Set(fiber.HeaderAuthorization, "Bearer "+c.Token)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	FileServiceImpl := &http.Client{}
	response, err := FileServiceImpl.Do(request)
	if err != nil {
		return responseErrHandler(response)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return responseErrHandler(response)
	}

	respBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return readErr
	}
	var resp []FileReponse
	if unmarshalErr := json.Unmarshal(respBody, &resp); unmarshalErr != nil {
		return unmarshalErr
	}
	if len(resp) < 1 {
		return fiber.NewError(fiber.StatusNotFound, "file/s not found")
	}
	return nil
}

func (c FileServiceImpl) GetFilesList() (files []map[string]any, err error) {
	type sortBy struct {
		Column string `json:"column"`
		Order  string `json:"order"`
	}
	type listReqBody struct {
		Limit         int    `json:"limit"`
		Offset        int    `json:"offset"`
		Prefix        string `json:"prefix"`
		SortByOptions sortBy `json:"sortBy"`
	}

	body := listReqBody{
		Limit:  999,
		Offset: 1,
		Prefix: "",
		SortByOptions: sortBy{
			Column: "name",
			Order:  "asc",
		},
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	listFileURL := c.ListFilesURL
	request, err := http.NewRequest(http.MethodPost, listFileURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	request.Header.Set(fiber.HeaderAuthorization, "Bearer "+c.Token)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	FileServiceImpl := &http.Client{}
	response, err := FileServiceImpl.Do(request)
	if err != nil {
		return nil, responseErrHandler(response)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, responseErrHandler(response)
	}

	respBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}
	var listResp []FileReponse

	if unmarshalErr := json.Unmarshal(respBody, &listResp); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	for _, list := range listResp {
		// Calculate the size in megabytes
		sizeInMB := float64(list.Metadata.Size) / 1024 / 1024
		formattedSize := fmt.Sprintf("%.4f MB", sizeInMB)
		file := map[string]interface{}{
			"name":        list.Name,
			"uploaded_at": list.CreatedAt,
			"size_mb":     formattedSize,
		}
		files = append(files, file)
	}

	return files, nil
}

func responseErrHandler(resp *http.Response) (err error) {
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}
	var errResp struct {
		Message    string `json:"message"`
		Error      string `json:"error"`
		StatusCode string `json:"statusCode"`
	}
	if unmarshalErr := json.Unmarshal(respBody, &errResp); unmarshalErr != nil {
		return unmarshalErr
	}
	statusCode, convErr := strconv.Atoi(errResp.StatusCode)
	if convErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed conv: "+convErr.Error())
	}
	return fiber.NewError(statusCode, errResp.Message+", "+errResp.Error)
}
