package handlers

import (
	"vector/internal/service"

	"github.com/gofiber/fiber/v2"
)

type SyncHandlers struct {
	stagingService  *service.StagingService
	applyService    *service.ApplyService
	fullSyncService *service.FullSyncService
}

func NewSyncHandlers(
	stagingService *service.StagingService,
	applyService *service.ApplyService,
	fullSyncService *service.FullSyncService,
) *SyncHandlers {
	return &SyncHandlers{
		stagingService:  stagingService,
		applyService:    applyService,
		fullSyncService: fullSyncService,
	}
}

func (h *SyncHandlers) SyncStaging(c *fiber.Ctx) error {
	page := c.Locals("page").(int)
	perPage := c.Locals("per_page").(int)

	resp, err := h.stagingService.SyncStaging(c.UserContext(), service.SyncStagingRequest{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	return c.JSON(resp)
}

func (h *SyncHandlers) SyncApply(c *fiber.Ctx) error {
	page := c.Locals("page").(int)
	perPage := c.Locals("per_page").(int)

	resp, err := h.applyService.SyncApply(c.UserContext(), service.SyncApplyRequest{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	return c.JSON(resp)
}

func (h *SyncHandlers) SyncFull(c *fiber.Ctx) error {
	perPage := c.Locals("per_page").(int)

	resp, err := h.fullSyncService.SyncFull(c.UserContext(), service.FullSyncRequest{
		PerPage:       perPage,
		SyncContracts: true,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(resp)
}
