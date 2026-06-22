package handlers

import (
	"fmt"
	"telegram-bot/config"
	"telegram-bot/database"
	"telegram-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	Bot *tgbotapi.BotAPI
	Cfg *config.Config
}

func NewHandler(bot *tgbotapi.BotAPI, cfg *config.Config) *Handler {
	return &Handler{Bot: bot, Cfg: cfg}
}

func (h *Handler) Start(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	var user models.User
	result := database.DB.First(&user, userID)

	if result.Error != nil {
		user = models.User{
			ID:        userID,
			FirstName: update.Message.From.FirstName,
			LastName:  update.Message.From.LastName,
			Username:  update.Message.From.UserName,
			Status:    models.StatusPending,
			Role:      models.RoleUser,
		}
		database.DB.Create(&user)

		msg := tgbotapi.NewMessage(chatID, "👋 Добро пожаловать!\n\nВы зарегистрированы. Ожидайте подтверждения от администратора.")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		h.Bot.Send(msg)
		h.notifyAdmins(&user)
		return
	}

	switch user.Status {
	case models.StatusPending:
		msg := tgbotapi.NewMessage(chatID, "⏳ Ожидайте подтверждения вашей регистрации администратором.")
		h.Bot.Send(msg)
	case models.StatusBlocked:
		msg := tgbotapi.NewMessage(chatID, "🚫 Ваш аккаунт заблокирован. Обратитесь к администратору.")
		h.Bot.Send(msg)
	case models.StatusApproved:
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ С возвращение, %s!", user.FirstName))
		if h.Cfg.IsAdmin(userID) {
			msg.ReplyMarkup = adminKeyboard()
		} else {
			msg.ReplyMarkup = userKeyboard()
		}
		h.Bot.Send(msg)
	}
}

func (h *Handler) notifyAdmins(user *models.User) {
	text := fmt.Sprintf("📥 Новая заявка на регистрацию\n\nID: %d\nИмя: %s %s\nUsername: @%s",
		user.ID, user.FirstName, user.LastName, user.Username)

	msg := tgbotapi.NewMessage(config.SuperAdminUID, text)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Одобрить", fmt.Sprintf("approve_%d", user.ID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отклонить", fmt.Sprintf("reject_%d", user.ID)),
		),
	)
	msg.ReplyMarkup = keyboard
	h.Bot.Send(msg)
}

func userKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Профиль"),
		),
	)
}

func adminKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Профиль"),
			tgbotapi.NewKeyboardButton("👥 Управление"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📥 Заявки"),
		),
	)
}
