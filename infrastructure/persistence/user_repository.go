package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	user_management "github.com/aruncs31s/azf/domain/user_management/model"
	"gorm.io/gorm"
)

type UserModel struct {
	ID            string `gorm:"primaryKey;type:varchar(36)"`
	Email         string `gorm:"uniqueIndex;type:varchar(254)"`
	Username      string `gorm:"uniqueIndex;type:varchar(50)"`
	DisplayName   string `gorm:"type:varchar(100)"`
	Status        string `gorm:"type:varchar(20)"`
	Roles         string `gorm:"type:text"` // JSON
	IsAdmin       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
	BlockedReason string `gorm:"type:text"`
	Metadata      string `gorm:"type:text"` // JSON
	OAuthProvider string `gorm:"size:50"`
	OAuthID       string `gorm:"size:255"`
}

func (UserModel) TableName() string {
	return "authz_users"
}

type RoleData struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) user_management.UserRepository {
	return NewGormUserRepository(db)
}

// Helper functions for conversion

func domainToModel(user *user_management.User) (*UserModel, error) {
	if user == nil {
		return nil, errors.New("user cannot be nil")
	}

	// Serialize roles
	roles := user.GetRoles()
	rolesData := make([]RoleData, len(roles))
	for i, role := range roles {
		rolesData[i] = RoleData{
			Name:        role.Name(),
			Permissions: role.Permissions(),
		}
	}
	rolesJSON, err := json.Marshal(rolesData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal roles: %w", err)
	}

	// Serialize metadata
	metadata := user.GetAllMetadata()
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	model := &UserModel{
		ID:            user.GetID(),
		Email:         user.GetEmail(),
		Username:      user.GetUsername(),
		DisplayName:   user.GetDisplayName(),
		Status:        string(user.GetStatus()),
		Roles:         string(rolesJSON),
		IsAdmin:       user.IsAdmin(),
		CreatedAt:     user.GetCreatedAt(),
		UpdatedAt:     user.GetUpdatedAt(),
		LastLoginAt:   user.GetLastLoginAt(),
		BlockedReason: user.GetBlockedReason(),
		Metadata:      string(metadataJSON),
		OAuthProvider: user.GetOAuthProvider(),
		OAuthID:       user.GetOAuthID(),
	}

	return model, nil
}

func modelToDomain(model *UserModel) (*user_management.User, error) {
	if model == nil {
		return nil, errors.New("model cannot be nil")
	}

	// Create user with basic info
	user, err := user_management.NewUser(model.ID, model.Email, model.Username, model.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Set status (this updates updatedAt, will correct later)
	status, err := user_management.NewUserStatus(model.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}
	user.SetStatus(status)

	// Set roles
	var rolesData []RoleData
	if model.Roles != "" {
		if err := json.Unmarshal([]byte(model.Roles), &rolesData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal roles: %w", err)
		}
	}
	roles := make([]*user_management.UserRole, 0, len(rolesData))
	for _, rd := range rolesData {
		role, err := user_management.NewUserRole(rd.Name, rd.Permissions)
		if err != nil {
			return nil, fmt.Errorf("invalid role data: %w", err)
		}
		roles = append(roles, role)
	}
	user.SetRoles(roles)

	// Set admin
	user.SetIsAdmin(model.IsAdmin)

	// Set timestamps (correct the updatedAt set by SetStatus)
	user.SetCreatedAt(model.CreatedAt)
	user.SetUpdatedAt(model.UpdatedAt)
	user.SetLastLoginAt(model.LastLoginAt)
	user.SetBlockedReason(model.BlockedReason)

	// Set metadata
	var metadata map[string]interface{}
	if model.Metadata != "" {
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		metadata = make(map[string]interface{})
	}
	user.SetAllMetadata(metadata)

	// Set OAuth fields
	user.SetOAuthProvider(model.OAuthProvider)
	user.SetOAuthID(model.OAuthID)

	return user, nil
}

// Implement UserReader

func (r *GormUserRepository) GetByID(ctx context.Context, userID string) (*user_management.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return modelToDomain(&model)
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*user_management.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return modelToDomain(&model)
}

func (r *GormUserRepository) GetByUsername(ctx context.Context, username string) (*user_management.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return modelToDomain(&model)
}

func (r *GormUserRepository) GetByOAuthID(ctx context.Context, provider, oauthID string) (*user_management.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("oauth_provider = ? AND oauth_id = ?", provider, oauthID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by OAuth ID: %w", err)
	}
	return modelToDomain(&model)
}

func (r *GormUserRepository) Search(ctx context.Context, query string, filter *user_management.UserSearchFilter) (*user_management.UserSearchResult, error) {
	db := r.db.WithContext(ctx).Model(&UserModel{})

	// Apply filters
	if filter.Status != nil {
		db = db.Where("status = ?", string(*filter.Status))
	}
	if filter.IsAdmin != nil {
		db = db.Where("is_admin = ?", *filter.IsAdmin)
	}
	if filter.RoleName != nil {
		// Since roles are JSON, this is tricky. For simplicity, use LIKE on the JSON string
		db = db.Where("roles LIKE ?", "%"+*filter.RoleName+"%")
	}
	if filter.CreatedAfter != nil {
		db = db.Where("created_at > ?", *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		db = db.Where("created_at < ?", *filter.CreatedBefore)
	}
	if filter.LastLoginAfter != nil {
		db = db.Where("last_login_at > ?", *filter.LastLoginAfter)
	}

	// Search query on username, displayName, email
	if query != "" {
		db = db.Where("username LIKE ? OR display_name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	// Count total
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Apply pagination
	limit := 20 // default
	offset := 0
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	if filter.Offset > 0 {
		offset = filter.Offset
	}
	db = db.Limit(limit).Offset(offset)

	// Fetch models
	var models []UserModel
	if err := db.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert to domain
	users := make([]*user_management.User, 0, len(models))
	for _, model := range models {
		user, err := modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to domain: %w", err)
		}
		users = append(users, user)
	}

	hasMore := int64(len(users)) == int64(limit) && int64(offset+len(users)) < total

	return &user_management.UserSearchResult{
		Users:   users,
		Total:   total,
		HasMore: hasMore,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func (r *GormUserRepository) ListAll(ctx context.Context, limit, offset int) ([]*user_management.User, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	var models []UserModel
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*user_management.User, 0, len(models))
	for _, model := range models {
		user, err := modelToDomain(&model)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert model to domain: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

func (r *GormUserRepository) GetAdmins(ctx context.Context) ([]*user_management.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Where("is_admin = ?", true).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}

	users := make([]*user_management.User, 0, len(models))
	for _, model := range models {
		user, err := modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to domain: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *GormUserRepository) GetByRole(ctx context.Context, roleName string) ([]*user_management.User, error) {
	var models []UserModel
	// Simple LIKE search in JSON
	if err := r.db.WithContext(ctx).Where("roles LIKE ?", "%"+roleName+"%").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}

	users := make([]*user_management.User, 0, len(models))
	for _, model := range models {
		user, err := modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to domain: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *GormUserRepository) GetByStatus(ctx context.Context, status user_management.UserStatus) ([]*user_management.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Where("status = ?", string(status)).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by status: %w", err)
	}

	users := make([]*user_management.User, 0, len(models))
	for _, model := range models {
		user, err := modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to domain: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Implement UserWriter

func (r *GormUserRepository) Create(ctx context.Context, user *user_management.User) (*user_management.User, error) {
	model, err := domainToModel(user)
	if err != nil {
		return nil, fmt.Errorf("failed to convert domain to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, errors.New("user with this email or username already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return modelToDomain(model)
}

func (r *GormUserRepository) Update(ctx context.Context, user *user_management.User) (*user_management.User, error) {
	model, err := domainToModel(user)
	if err != nil {
		return nil, fmt.Errorf("failed to convert domain to model: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return modelToDomain(model)
}

func (r *GormUserRepository) Delete(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", userID).Delete(&UserModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *GormUserRepository) Block(ctx context.Context, userID string, reason string) (*user_management.User, error) {
	// First get the user
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Block in domain
	if err := user.Block(reason); err != nil {
		return nil, err
	}

	// Update
	return r.Update(ctx, user)
}

func (r *GormUserRepository) Unblock(ctx context.Context, userID string) (*user_management.User, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := user.Unblock(); err != nil {
		return nil, err
	}

	return r.Update(ctx, user)
}

func (r *GormUserRepository) AssignRole(ctx context.Context, userID string, role *user_management.UserRole) (*user_management.User, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := user.AssignRole(role); err != nil {
		return nil, err
	}

	return r.Update(ctx, user)
}

func (r *GormUserRepository) RemoveRole(ctx context.Context, userID string, role *user_management.UserRole) (*user_management.User, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := user.RemoveRole(role); err != nil {
		return nil, err
	}

	return r.Update(ctx, user)
}

func (r *GormUserRepository) PromoteToAdmin(ctx context.Context, userID string) (*user_management.User, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := user.PromoteToAdmin(); err != nil {
		return nil, err
	}

	return r.Update(ctx, user)
}

func (r *GormUserRepository) DemoteFromAdmin(ctx context.Context, userID string) (*user_management.User, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := user.DemoteFromAdmin(); err != nil {
		return nil, err
	}

	return r.Update(ctx, user)
}
