package document

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
)

type UseCase struct {
	httpClient httpx.Client
	logger     *logger.ZapLogger
}

func NewUseCase(httpClient httpx.Client, zl *logger.ZapLogger) *UseCase {
	return &UseCase{
		httpClient: httpClient,
		logger:     zl,
	}
}

func (u *UseCase) List(ctx context.Context, filter *domain.Pagination) ([]*domain.Document, string, error) {
	url, err := httpx.BuildURL(config.GoogleDriveMetaV3URI, map[string]string{
		"pageSize":  strconv.FormatUint(filter.PageSize, 10),
		"pageToken": filter.PageToken,
		"corpora":   "user",
	})
	if err != nil {
		return nil, "", err
	}

	accessToken, _ := ctx.Value(domain.KeyAccessToken).(string)
	resp, err := u.httpClient.DoRequest(ctx, http.MethodGet, url, nil, http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", accessToken)},
	})
	if err != nil {
		return nil, "", err
	}

	var respData FileListResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode >= 400 && respData.Error != nil {
		u.logger.Error(respData.Error.Message)
		return nil, "", domain.InternalServerError
	}

	var documents []*domain.Document
	for _, file := range respData.Files {
		documents = append(documents, &domain.Document{
			ID:       file.ID,
			Name:     file.Name,
			Type:     file.Type,
			IsFolder: file.Type == domain.MimeTypeFolder,
		})
	}

	return documents, respData.NextPageToken, nil
}

func (u *UseCase) Create(ctx context.Context, fileName, fileType string, driveId ...string) (*domain.Document, error) {
	url, err := httpx.BuildURL(config.GoogleDriveMetaV3URI, map[string]string{
		"uploadType": "multipart",
	})

	payload := map[string]any{
		"kind": "drive#file",
	}

	if len(driveId) > 0 {
		payload["parents"] = driveId[0]
	}

	fileNameParts := strings.Split(fileName, ".")
	if fileType == domain.FileTypeFolder {
		payload["name"] = fileNameParts[0]
		payload["mimeType"] = domain.MimeTypeFolder
	} else {
		payload["mimeType"] = "text/markdown"
		if len(fileNameParts) == 1 || fileNameParts[len(fileNameParts)-1] != "md" {
			fileName = fileName + ".md"
		}
		payload["name"] = fileName
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	accessToken, _ := ctx.Value(domain.KeyAccessToken).(string)
	resp, err := u.httpClient.DoRequest(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", accessToken)},
		"Content-Type":  []string{"application/json"},
	})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, domain.InternalServerError
	}

	return &domain.Document{}, nil
}

func (u *UseCase) Upload(ctx context.Context, contents []byte, fileName string) error {
	url, err := httpx.BuildURL(config.GoogleDriveMediaV3URI, map[string]string{
		"uploadType": "resumable",
	})

	payload := map[string]any{
		"kind": "drive#file",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	accessToken, _ := ctx.Value(domain.KeyAccessToken).(string)
	resp, err := u.httpClient.DoRequest(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", accessToken)},
		"Content-Type":  []string{"application/json"},
	})
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return domain.InternalServerError
	}

	return nil
}

func (u *UseCase) Get(ctx context.Context, id string) *domain.Document {
	return nil
}

func (u *UseCase) Update(ctx context.Context, id string) *domain.Document {
	return nil
}
