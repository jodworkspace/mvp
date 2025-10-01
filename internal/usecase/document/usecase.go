package document

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

func (u *UseCase) List(ctx context.Context, filter *domain.Pagination) ([]*domain.Document, error) {
	url, err := httpx.BuildURL(config.GoogleDriveMetadataURI, map[string]string{
		"pageSize":  strconv.FormatUint(filter.PageSize, 10),
		"pageToken": filter.PageToken,
		"corpora":   "user",
	})
	if err != nil {
		return nil, err
	}

	accessToken, _ := ctx.Value(domain.KeyAccessToken).(string)
	resp, err := u.httpClient.DoRequest(ctx, http.MethodGet, url, nil, http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", accessToken)},
	})
	if err != nil {
		return nil, err
	}

	var respData FileListResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 && respData.Error != nil {
		u.logger.Error(respData.Error.Message)
		return nil, domain.InternalServerError
	}

	var documents []*domain.Document
	for _, file := range respData.Files {
		documents = append(documents, &domain.Document{
			ID:       file.ID,
			Name:     file.Name,
			Type:     file.Type,
			IsFolder: file.Type == domain.FileTypeFolder,
		})
	}

	return documents, nil
}

func (u *UseCase) Get(ctx context.Context, id string) *domain.Document {
	return nil
}
