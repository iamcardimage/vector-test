package repository

import (
	"context"
	"vector/internal/external"
)

type externalClientAdapter struct {
	client *external.Client
}

func NewExternalAPIClient(client *external.Client) ExternalAPIClient {
	return &externalClientAdapter{client: client}
}

func (a *externalClientAdapter) GetUsersRaw(ctx context.Context, page, perPage int) (*ExternalUsersResponse, error) {
	httpResp, err := a.client.GetUsersRaw(ctx, page, perPage)
	if err != nil {
		return nil, err
	}

	return &ExternalUsersResponse{
		Success:     httpResp.Success,
		TotalCount:  httpResp.TotalCount,
		PerPage:     httpResp.PerPage,
		CurrentPage: httpResp.CurrentPage,
		TotalPages:  httpResp.TotalPages,
		Users:       httpResp.Users,
	}, nil
}
