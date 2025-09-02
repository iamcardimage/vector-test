package service

import (
	"errors"
	"time"
	"vector/internal/models"
	"vector/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo  repository.UserRepository
	jwtConfig models.JWTConfig
}

func NewAuthService(userRepo repository.UserRepository, jwtConfig models.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

func (s *AuthService) Login(email, password string) (*models.LoginResponse, error) {
	// Найти пользователя по email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("неверный email или пароль")
	}

	if !user.IsActive {
		return nil, errors.New("пользователь деактивирован")
	}

	if !user.CheckPassword(password) {
		return nil, errors.New("неверный email или пароль")
	}

	token, expiresAt, err := s.generateJWT(user)
	if err != nil {
		return nil, errors.New("ошибка создания токена")
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.PublicUser(),
	}, nil
}

func (s *AuthService) CreateUser(req models.CreateUserRequest, creatorRole string) (*models.AppUser, error) {

	if creatorRole != models.RoleAdministrator {
		return nil, errors.New("недостаточно прав для создания пользователя")
	}

	if !isValidRole(req.Role) {
		return nil, errors.New("недопустимая роль")
	}

	user := models.AppUser{
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		MiddleName: req.MiddleName,
		Role:       req.Role,
		IsActive:   true,
	}

	if err := user.HashPassword(req.Password); err != nil {
		return nil, errors.New("ошибка хеширования пароля")
	}

	createdUser, err := s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return &createdUser, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*models.AppUser, error) {

	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неверный метод подписи токена")
		}
		return []byte(s.jwtConfig.SecretKey), nil
	})

	if err != nil {
		return nil, errors.New("неверный токен")
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("неверный токен")
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if !user.IsActive {
		return nil, errors.New("пользователь деактивирован")
	}

	return &user, nil
}

func (s *AuthService) generateJWT(user models.AppUser) (string, int64, error) {
	expiresAt := time.Now().Add(s.jwtConfig.TokenDuration)

	claims := &models.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "vector-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtConfig.SecretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

func isValidRole(role string) bool {
	switch role {
	case models.RoleAdministrator, models.RolePodft, models.RoleClientManagement:
		return true
	}
	return false
}
