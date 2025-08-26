package repository

import (
	"context"
	"vector/internal/external"
)

type syncExternalClientAdapter struct {
	client *external.Client
}

func NewSyncExternalAPIClient(client *external.Client) ExternalAPIClient {
	return &syncExternalClientAdapter{client: client}
}

func (a *syncExternalClientAdapter) GetUsersRaw(ctx context.Context, page, perPage int) (*ExternalUsersResponse, error) {
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

func (a *syncExternalClientAdapter) GetContractsRaw(ctx context.Context, page, perPage int) (*ExternalContractsResponse, error) {
	httpResp, err := a.client.GetContractsRaw(ctx, page, perPage)
	if err != nil {
		return nil, err
	}

	return &ExternalContractsResponse{
		Success:     httpResp.Success,
		TotalCount:  httpResp.TotalCount,
		PerPage:     httpResp.PerPage,
		CurrentPage: httpResp.CurrentPage,
		TotalPages:  httpResp.TotalPages,
		Contracts:   httpResp.Contracts,
	}, nil
}
