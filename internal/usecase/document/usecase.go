package document

import (
	"context"
	"fmt"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
)

type UseCase struct {
	httpClient httpx.Client
}

func NewUseCase() *UseCase {
	return &UseCase{}
}

func (u *UseCase) List(ctx context.Context, filter *domain.Filter) ([]*domain.Document, error) {
	return nil, nil
}

func (u *UseCase) Get() {}

func (u *UseCase) Update(ctx context.Context, id string, diffs []dmp.Diff) {
	d := dmp.New()
	patches := d.PatchMake(diffs)
	result, _ := d.PatchApply(patches, "Hello World")
	fmt.Println(result[0])
}
