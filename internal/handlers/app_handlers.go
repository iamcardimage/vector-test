package handlers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"vector/internal/models"
	"vector/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
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

//	func (h *AppHandlers) canMutate(c *fiber.Ctx) (models.AppUser, bool) {
//		u := c.Locals("user").(models.AppUser)
//		if u.Role == "viewer" {
//			c.Status(403).JSON(fiber.Map{"error": "forbidden"})
//			return u, false
//		}
//		return u, true
//	}

// =========Swagger=================
// GetClient godoc
// @Summary Get client information
// @Description Get complete client information including all available fields and second part if available
// @Tags clients
// @Accept json
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} models.GetClientResponse "Complete client information"
// @Failure 400 {object} models.ErrorResponse "Invalid client ID"
// @Failure 404 {object} models.ErrorResponse "Client not found"
// @Router /clients/{id} [get]
func (h *AppHandlers) GetClient(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(models.ErrorResponse{Error: "invalid client id"})
	}

	cur, err := h.appService.GetClientCurrent(id)
	if err != nil {
		return c.Status(404).JSON(models.ErrorResponse{Error: "client not found"})
	}

	response := models.GetClientResponse{

		ID:            cur.ID,
		ClientID:      cur.ClientID,
		Version:       cur.Version,
		ExternalID:    cur.ID,
		ExternalIDStr: cur.ExternalIDStr,

		Surname:    cur.Surname,
		Name:       cur.Name,
		Patronymic: cur.Patronymic,
		Birthday:   cur.Birthday,
		BirthPlace: cur.BirthPlace,

		Inn:             cur.Inn,
		Snils:           cur.Snils,
		PassSeries:      cur.PassSeries,
		PassNumber:      cur.PassNumber,
		PassIssueDate:   cur.PassIssueDate,
		PassIssuer:      cur.PassIssuer,
		PassIssuerCode:  cur.PassIssuerCode,
		DocumentType:    cur.DocumentType,
		DocumentCountry: cur.DocumentCountry,

		ContactEmail: cur.ContactEmail,
		MainPhone:    cur.MainPhone,
		Login:        cur.Login,

		Blocked:            cur.Blocked,
		BlockedReason:      cur.BlockedReason,
		BlockType:          cur.BlockType,
		Male:               cur.Male,
		IsRfResident:       cur.IsRfResident,
		IsRfTaxpayer:       cur.IsRfTaxpayer,
		IsValidInfo:        cur.IsValidInfo,
		QualifiedInvestor:  cur.QualifiedInvestor,
		IsFilled:           cur.IsFilled,
		IsAmericanNational: cur.IsAmericanNational,
		NeedToSetPassword:  cur.NeedToSetPassword,

		LegalCapacity:      cur.LegalCapacity,
		PifsPortfolioCode:  cur.PifsPortfolioCode,
		RiskLevel:          cur.RiskLevel,
		ExternalRiskLevel:  cur.ExternalRiskLevel,
		FillStage:          cur.FillStage,
		IdentificationType: cur.IdentificationType,
		TaxStatus:          cur.TaxStatus,

		EsiaID:       cur.EsiaID,
		AgentID:      cur.AgentID,
		AgentPointID: cur.AgentPointID,

		Country:  cur.Country,
		Region:   cur.Region,
		Index:    cur.Index,
		City:     cur.City,
		Street:   cur.Street,
		House:    cur.House,
		Corps:    cur.Corps,
		Flat:     cur.Flat,
		District: cur.District,

		SignatureType:              cur.SignatureType,
		DataReceivedDigitalProfile: cur.DataReceivedDigitalProfile,
		LockedAt:                   cur.LockedAt,
		CurrentSignInAt:            cur.CurrentSignInAt,
		SignInCount:                cur.SignInCount,

		CreatedLKAt: cur.CreatedLKAt,
		UpdatedLKAt: cur.UpdatedLKAt,
		SyncedAt:    cur.SyncedAt,
		ValidFrom:   cur.ValidFrom,
		ValidTo:     cur.ValidTo,
		IsCurrent:   cur.IsCurrent,

		Hash:                  cur.Hash,
		Status:                cur.Status,
		SecondPartTriggerHash: cur.SecondPartTriggerHash,

		NeedsSecondPart:   cur.NeedsSecondPart,
		SecondPartCreated: cur.SecondPartCreated,
	}

	response.FromCompanySettings = convertJSONToMap(cur.FromCompanySettings)
	response.Settings = convertJSONToMap(cur.Settings)
	response.PersonInfo = convertJSONToMap(cur.PersonInfo)
	response.Manager = convertJSONToMap(cur.Manager)
	response.Checks = convertJSONToMap(cur.Checks)
	response.Note = convertJSONToMap(cur.Note)
	response.AdSource = convertJSONToMap(cur.AdSource)
	response.SignatureAllowedNumbers = convertJSONToMap(cur.SignatureAllowedNumbers)
	response.Raw = convertJSONToMap(cur.Raw)

	if curSP, err := h.appService.GetSecondPartCurrent(id); err == nil {
		response.SecondPart = &struct {
			ClientVersion int        `json:"client_version" example:"1"`
			Version       int        `json:"version" example:"1"`
			Status        string     `json:"status" example:"draft"`
			RiskLevel     string     `json:"risk_level" example:"low"`
			IsCurrent     bool       `json:"is_current" example:"true"`
			DueAt         *time.Time `json:"due_at" swaggertype:"string" format:"date-time"`
		}{
			ClientVersion: curSP.ClientVersion,
			Version:       curSP.Version,
			Status:        curSP.Status,
			RiskLevel:     curSP.RiskLevel,
			IsCurrent:     curSP.IsCurrent,
			DueAt:         curSP.DueAt,
		}
	}

	return c.JSON(response)
}

func convertJSONToMap(jsonData datatypes.JSON) *map[string]interface{} {
	if len(jsonData) == 0 {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil
	}

	return &result
}

// @Summary Get second part history for client
// @Description Get history of second part information for specific client
// @Tags clients
// @Accept json
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} map[string]interface{} "Second part history"
// @Failure 400 {object} map[string]interface{} "Invalid client ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /clients/{id}/second-part/history [get]
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

// CreateUser godoc
// @Summary Create new user
// @Description Create a new user with email, role and optional token
// @Tags auth
// @Accept json
// @Produce json
// @Param user body object{email=string,role=string,token=string} true "User data"
// @Success 200 {object} map[string]interface{} "Created user"
// @Failure 400 {object} map[string]interface{} "Invalid input or creation failed"
// @Router /auth/register [post]
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

// ListUsers godoc
// @Summary List all users
// @Description Get list of all users in the system
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List of users"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/users [get]
func (h *AppHandlers) ListUsers(c *fiber.Ctx) error {
	users, err := h.appService.ListUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "list: " + err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "users": users})
}

// UpdateUserRole godoc
// @Summary Update user role
// @Description Update the role of an existing user
// @Tags auth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param role body object{role=string} true "New role"
// @Success 200 {object} map[string]interface{} "Updated user"
// @Failure 400 {object} map[string]interface{} "Invalid input or update failed"
// @Router /auth/users/{id}/role [patch]
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

// RotateUserToken godoc
// @Summary Rotate user token
// @Description Generate new token for an existing user
// @Tags auth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User with new token"
// @Failure 400 {object} map[string]interface{} "Invalid input or rotation failed"
// @Router /auth/users/{id}/rotate-token [post]
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

// DeleteUser godoc
// @Summary Delete user
// @Description Delete an existing user from the system
// @Tags auth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "Success confirmation"
// @Failure 400 {object} map[string]interface{} "Invalid input or deletion failed"
// @Router /auth/users/{id} [delete]
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

// CreateSecondPartDraft godoc
// @Summary Create second part draft
// @Description Create a new second part draft for a client
// @Tags clients
// @Accept json
// @Produce json
// @Param id path int true "Client ID"
// @Param draft body object{risk_level=string,data_override=object} false "Draft data"
// @Success 200 {object} map[string]interface{} "Created second part draft"
// @Failure 400 {object} map[string]interface{} "Invalid input or creation failed"
// @Router /clients/{id}/second-part/draft [post]
func (h *AppHandlers) CreateSecondPartDraft(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid client id"})
	}

	var in struct {
		RiskLevel    *string                 `json:"risk_level,omitempty"`
		DataOverride *map[string]interface{} `json:"data_override,omitempty"`
	}

	if err := c.BodyParser(&in); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid json"})
	}

	// TODO: Get user from auth middleware
	var createdBy *int = nil // This will be set from authenticated user later

	var dataOverrideJSON *datatypes.JSON
	if in.DataOverride != nil {
		jsonBytes, err := json.Marshal(in.DataOverride)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid data_override"})
		}
		dataOverrideJSON = (*datatypes.JSON)(&jsonBytes)
	}

	sp, err := h.appService.CreateSecondPartDraft(id, in.RiskLevel, createdBy, dataOverrideJSON)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":        true,
		"client_version": sp.ClientVersion,
		"version":        sp.Version,
		"status":         sp.Status,
		"risk_level":     sp.RiskLevel,
		"due_at":         sp.DueAt,
		"is_current":     sp.IsCurrent,
	})
}

// GetContract godoc
// @Summary Get contract information
// @Description Get complete contract information by contract ID
// @Tags contracts
// @Accept json
// @Produce json
// @Param id path int true "Contract ID"
// @Success 200 {object} models.GetContractResponse "Complete contract information"
// @Failure 400 {object} models.ErrorResponse "Invalid contract ID"
// @Failure 404 {object} models.ErrorResponse "Contract not found"
// @Router /contracts/{id} [get]
func (h *AppHandlers) GetContract(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(models.ErrorResponse{Error: "invalid contract id"})
	}

	contract, err := h.appService.GetContract(id)
	if err != nil {
		return c.Status(404).JSON(models.ErrorResponse{Error: "contract not found"})
	}

	response := models.GetContractResponse{
		ID:         contract.ID,
		ExternalID: contract.ExternalID,
		UserID:     contract.UserID,

		InnerCode:         contract.InnerCode,
		Kind:              contract.Kind,
		Status:            contract.Status,
		ContractOwnerType: contract.ContractOwnerType,
		ContractOwnerID:   contract.ContractOwnerID,
		Comment:           contract.Comment,

		IsPersonalInvestAccount:    contract.IsPersonalInvestAccount,
		IsPersonalInvestAccountNew: contract.IsPersonalInvestAccountNew,

		RialtoCode: contract.RialtoCode,
		Anketa:     contract.Anketa,
		OwnerID:    contract.OwnerID,
		UserLogin:  contract.UserLogin,

		CalculatedProfileID: contract.CalculatedProfileID,
		DepoAccountsType:    contract.DepoAccountsType,
		StrategyID:          contract.StrategyID,
		StrategyName:        contract.StrategyName,
		TariffID:            contract.TariffID,
		TariffName:          contract.TariffName,

		CreatedAt: contract.CreatedAt,
		UpdatedAt: contract.UpdatedAt,
		SignedAt:  contract.SignedAt,
		ClosedAt:  contract.ClosedAt,
		SyncedAt:  contract.SyncedAt,

		Hash: contract.Hash,
		Raw:  convertJSONToMap(contract.Raw),
	}

	return c.JSON(response)
}

// ListContracts godoc
// @Summary List contracts
// @Description Get list of contracts with optional filtering
// @Tags contracts
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status (active, closed)"
// @Success 200 {object} models.ListContractsResponse "List of contracts"
// @Failure 400 {object} models.ErrorResponse "Invalid parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /contracts [get]
func (h *AppHandlers) ListContracts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 10)

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	var userID *int
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if uid, err := strconv.Atoi(userIDStr); err == nil {
			userID = &uid
		}
	}

	var status *string
	if statusStr := c.Query("status"); statusStr != "" {
		status = &statusStr
	}

	contracts, total, err := h.appService.ListContracts(page, perPage, userID, status)
	if err != nil {
		return c.Status(500).JSON(models.ErrorResponse{Error: "failed to get contracts: " + err.Error()})
	}

	contractResponses := make([]models.GetContractResponse, len(contracts))
	for i, contract := range contracts {
		contractResponses[i] = models.GetContractResponse{
			ID:         contract.ID,
			ExternalID: contract.ExternalID,
			UserID:     contract.UserID,

			InnerCode:         contract.InnerCode,
			Kind:              contract.Kind,
			Status:            contract.Status,
			ContractOwnerType: contract.ContractOwnerType,
			ContractOwnerID:   contract.ContractOwnerID,
			Comment:           contract.Comment,

			IsPersonalInvestAccount:    contract.IsPersonalInvestAccount,
			IsPersonalInvestAccountNew: contract.IsPersonalInvestAccountNew,

			RialtoCode: contract.RialtoCode,
			Anketa:     contract.Anketa,
			OwnerID:    contract.OwnerID,
			UserLogin:  contract.UserLogin,

			CalculatedProfileID: contract.CalculatedProfileID,
			DepoAccountsType:    contract.DepoAccountsType,
			StrategyID:          contract.StrategyID,
			StrategyName:        contract.StrategyName,
			TariffID:            contract.TariffID,
			TariffName:          contract.TariffName,

			CreatedAt: contract.CreatedAt,
			UpdatedAt: contract.UpdatedAt,
			SignedAt:  contract.SignedAt,
			ClosedAt:  contract.ClosedAt,
			SyncedAt:  contract.SyncedAt,

			Hash: contract.Hash,
			Raw:  convertJSONToMap(contract.Raw),
		}
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	response := models.ListContractsResponse{
		Success:    true,
		Contracts:  contractResponses,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}

	return c.JSON(response)
}
