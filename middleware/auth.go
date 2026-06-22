package middleware

import (
	"telegram-bot/config"
	"telegram-bot/database"
	"telegram-bot/models"
	"telegram-bot/sanitize"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserContext struct {
	User    *models.User
	IsAdmin bool
}

func GetUser(update *tgbotapi.Update) (*models.User, bool) {
	var user models.User
	result := database.DB.First(&user, update.Message.From.ID)
	if result.Error != nil {
		return nil, false
	}
	return &user, true
}

func HasAccess(user *models.User) bool {
	if user == nil {
		return false
	}
	return user.Status == models.StatusApproved || user.Role == models.RoleSuperAdmin
}

func GetUserID(update *tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

func EnsureUserExists(cfg *config.Config, update *tgbotapi.Update) *models.User {
	userID := GetUserID(update)
	if userID == 0 {
		return nil
	}

	if userID == config.SuperAdminUID {
		return &models.User{
			ID:        userID,
			Status:    models.StatusApproved,
			Role:      models.RoleSuperAdmin,
			FirstName: "SuperAdmin",
		}
	}

	var user models.User
	result := database.DB.First(&user, userID)
	if result.Error != nil {
		user = models.User{
			ID:        userID,
			Status:    models.StatusPending,
			Role:      models.RoleUser,
			FirstName: sanitize.Name(update.Message.From.FirstName),
			LastName:  sanitize.Name(update.Message.From.LastName),
			Username:  sanitize.Username(update.Message.From.UserName),
		}
		database.DB.Create(&user)
	}
	return &user
}
