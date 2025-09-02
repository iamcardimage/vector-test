package handlers

import (
	"strings"
	"vector/internal/middleware"
	"vector/internal/models"
	"vector/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AuthHandlers struct {
	authService *service.AuthService
}

func NewAuthHandlers(authService *service.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Вход в систему с email и паролем, возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Данные для входа"
// @Success 200 {object} models.LoginResponse "Успешная аутентификация"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 401 {object} map[string]interface{} "Неверный email или пароль"
// @Router /auth/login [post]
func (h *AuthHandlers) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Неверный формат данных",
			"success": false,
		})
	}

	// Базовая валидация
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Email и пароль обязательны",
			"success": false,
		})
	}

	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// CreateUser godoc
// @Summary Создание нового пользователя (только для администраторов)
// @Description Создать нового пользователя в системе. Доступно только администраторам.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "Данные пользователя"
// @Success 201 {object} map[string]interface{} "Пользователь создан"
// @Failure 400 {object} map[string]interface{} "Неверные данные"
// @Failure 403 {object} map[string]interface{} "Недостаточно прав"
// @Router /auth/users [post]
func (h *AuthHandlers) CreateUser(c *fiber.Ctx) error {
	// Получаем текущего пользователя
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error":   "Пользователь не аутентифицирован",
			"success": false,
		})
	}

	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Неверный формат данных",
			"success": false,
		})
	}

	// Базовая валидация
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" ||
		strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" ||
		strings.TrimSpace(req.Role) == "" {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Все обязательные поля должны быть заполнены",
			"success": false,
		})
	}

	user, err := h.authService.CreateUser(req, currentUser.Role)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    user.PublicUser(),
		"message": "Пользователь успешно создан",
	})
}

// GetProfile godoc
// @Summary Получить профиль текущего пользователя
// @Description Получить информацию о текущем аутентифицированном пользователе
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Профиль пользователя"
// @Failure 401 {object} map[string]interface{} "Пользователь не аутентифицирован"
// @Router /auth/profile [get]
func (h *AuthHandlers) GetProfile(c *fiber.Ctx) error {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error":   "Пользователь не аутентифицирован",
			"success": false,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    currentUser.PublicUser(),
	})
}

// GetRoles godoc
// @Summary Получить список доступных ролей
// @Description Получить список всех доступных ролей в системе
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Список ролей"
// @Router /auth/roles [get]
func (h *AuthHandlers) GetRoles(c *fiber.Ctx) error {
	roles := []map[string]string{
		{
			"value":       models.RoleAdministrator,
			"label":       "Администратор",
			"description": "Полный доступ к системе",
		},
		{
			"value":       models.RolePodft,
			"label":       "Отдел ПОДФТ",
			"description": "Работа с проверками ПОД/ФТ",
		},
		{
			"value":       models.RoleClientManagement,
			"label":       "Клиентский отдел",
			"description": "Работа с клиентами",
		},
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    roles,
	})
}
