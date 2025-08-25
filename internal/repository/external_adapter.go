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
	resp, err := a.client.GetUsersRaw(ctx, page, perPage)
	if err != nil {
		return nil, err
	}

	return &ExternalUsersResponse{
		Success:     resp.Success,
		TotalCount:  resp.TotalCount,
		PerPage:     resp.PerPage,
		CurrentPage: resp.CurrentPage,
		TotalPages:  resp.TotalPages,
		Users:       resp.Users,
	}, nil
}
