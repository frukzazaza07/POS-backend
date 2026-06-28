package service

import (
	"errors"
	"pos-backend/internal/middleware"
	"pos-backend/internal/models"
	"pos-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Password string          `json:"password"`
	Role     models.UserRole `json:"role"`
}

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	if !user.IsActive {
		return "", nil, errors.New("account is inactive")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *AuthService) Register(req RegisterRequest) (*models.User, error) {
	if existing, _ := s.userRepo.FindByEmail(req.Email); existing != nil {
		return nil, errors.New("email already registered")
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New("name, email, and password are required")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := req.Role
	if role == "" {
		role = models.RoleCashier
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
		Role:     role,
		IsActive: true,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}
