package handlers

import (
	"strconv"
	"strings"
	"vector/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AppHandlers struct {
	appService *service.AppService
}

func NewAppHandlers(appService *service.AppService) *AppHandlers {
	return &AppHandlers{
		appService: appService,
	}
}

// func asIntPtr(u uint) *int {
// 	i := int(u)
// 	return &i
// }

// func (h *AppHandlers) hasRole(u models.AppUser, roles ...string) bool {
// 	for _, r := range roles {
// 		if u.Role == r {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (h *AppHandlers) requireRoles(c *fiber.Ctx, roles ...string) (models.AppUser, bool) {
// 	u := c.Locals("user").(models.AppUser)
// 	if len(roles) == 0 || h.hasRole(u, roles...) || u.Role == models.RoleAdministrator {
// 		return u, true
// 	}
// 	c.Status(403).JSON(fiber.Map{"error": "forbidden"})
// 	return u, false
// }

// func (h *AppHandlers) canMutate(c *fiber.Ctx) (models.AppUser, bool) {
// 	u := c.Locals("user").(models.AppUser)
// 	if u.Role == "viewer" {
// 		c.Status(403).JSON(fiber.Map{"error": "forbidden"})
// 		return u, false
// 	}
// 	return u, true
// }

func (h *AppHandlers) GetClient(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid client id"})
	}

	cur, err := h.appService.GetClientCurrent(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "client not found"})
	}

	var sp fiber.Map
	if curSP, err := h.appService.GetSecondPartCurrent(id); err == nil {
		sp = fiber.Map{
			"client_version": curSP.ClientVersion,
			"version":        curSP.Version,
			"status":         curSP.Status,
			"risk_level":     curSP.RiskLevel,
			"due_at":         curSP.DueAt,
			"is_current":     curSP.IsCurrent,
		}
		if curSP.ClientVersion != cur.Version {
			sp["stale"] = true
		}
	}

	return c.JSON(fiber.Map{
		"id":                  cur.ClientID,
		"version":             cur.Version,
		"surname":             cur.Surname,
		"name":                cur.Name,
		"patronymic":          cur.Patronymic,
		"birthday":            cur.Birthday,
		"birth_place":         cur.BirthPlace,
		"inn":                 cur.Inn,
		"snils":               cur.Snils,
		"created_lk_at":       cur.CreatedLKAt,
		"updated_lk_at":       cur.UpdatedLKAt,
		"pass_issuer_code":    cur.PassIssuerCode,
		"pass_series":         cur.PassSeries,
		"pass_number":         cur.PassNumber,
		"pass_issue_date":     cur.PassIssueDate,
		"pass_issuer":         cur.PassIssuer,
		"contact_email":       cur.ContactEmail,
		"main_phone":          cur.MainPhone,
		"needs_second_part":   cur.NeedsSecondPart,
		"second_part_created": cur.SecondPartCreated,
		"second_part":         sp,
	})
}

func (h *AppHandlers) GetSecondPartHistory(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid client id"})
	}

	vs, err := h.appService.ListSecondPartHistory(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "history: " + err.Error()})
	}

	out := make([]fiber.Map, 0, len(vs))
	for _, v := range vs {
		out = append(out, fiber.Map{
			"client_version": v.ClientVersion,
			"version":        v.Version,
			"status":         v.Status,
			"risk_level":     v.RiskLevel,
			"due_at":         v.DueAt,
			"valid_from":     v.ValidFrom,
			"valid_to":       v.ValidTo,
			"is_current":     v.IsCurrent,
		})
	}

	return c.JSON(fiber.Map{"success": true, "versions": out})
}

func (h *AppHandlers) CreateUser(c *fiber.Ctx) error {
	var in struct {
		Email string `json:"email"`
		Role  string `json:"role"`
		Token string `json:"token,omitempty"`
	}

	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
	}

	u, err := h.appService.CreateUser(in.Email, in.Role, in.Token)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"id": u.ID, "email": u.Email, "role": u.Role, "token": u.Token,
	})
}

func (h *AppHandlers) ListUsers(c *fiber.Ctx) error {
	users, err := h.appService.ListUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "list: " + err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "users": users})
}

func (h *AppHandlers) UpdateUserRole(c *fiber.Ctx) error {
	uid64, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	var in struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&in); err != nil || strings.TrimSpace(in.Role) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "role required"})
	}

	u, err := h.appService.UpdateUserRole(uint(uid64), in.Role)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(u)
}

func (h *AppHandlers) RotateUserToken(c *fiber.Ctx) error {
	uid64, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	u, err := h.appService.RotateUserToken(uint(uid64))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(u)
}

func (h *AppHandlers) DeleteUser(c *fiber.Ctx) error {
	uid64, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	if err := h.appService.DeleteUser(uint(uid64)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}
