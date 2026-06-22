package handlers

import (
	"fmt"
	"telegram-bot/config"
	"telegram-bot/database"
	"telegram-bot/models"
	"telegram-bot/sanitize"

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

	if h.Cfg.IsAdmin(userID) {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✅ <b>Добро пожаловать, SuperAdmin!</b>\n\n"+
				"🚀 %s v%s",
			config.AppName, config.AppVersion))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = adminKeyboard()
		h.Bot.Send(msg)
		return
	}

	var user models.User
	result := database.DB.First(&user, userID)

	if result.Error != nil {
		msg := tgbotapi.NewMessage(chatID,
			"👋 <b>Добро пожаловать!</b>\n\n"+
				"Для доступа к боту необходимо зарегистрироваться.\n\n"+
				"⚠️ <b>Внимание:</b> при регистрации будет передана ваша карточка контакта (номер телефона).")
		msg.ParseMode = "HTML"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📝 Зарегистрироваться", "register_start"),
			),
		)
		msg.ReplyMarkup = keyboard
		h.Bot.Send(msg)
		return
	}

	switch user.Status {
	case models.StatusPending:
		msg := tgbotapi.NewMessage(chatID,
			"⏳ <b>Ожидайте подтверждения</b>\n\n"+
				"Ваша заявка на регистрацию передана администратору.")
		msg.ParseMode = "HTML"
		h.Bot.Send(msg)
	case models.StatusBlocked:
		msg := tgbotapi.NewMessage(chatID,
			"🚫 <b>Аккаунт заблокирован</b>\n\n"+
				"Обратитесь к администратору.")
		msg.ParseMode = "HTML"
		h.Bot.Send(msg)
	case models.StatusApproved:
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ <b>С возвращение, %s!</b>", user.FirstName))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = userKeyboard()
		h.Bot.Send(msg)
	}
}

func (h *Handler) RegisterStart(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	var user models.User
	result := database.DB.First(&user, userID)
	if result.Error == nil {
		msg := tgbotapi.NewMessage(chatID, "ℹ️ <b>Вы уже зарегистрированы</b>")
		msg.ParseMode = "HTML"
		h.Bot.Send(msg)
		return
	}

	contactBtn := tgbotapi.NewKeyboardButtonContact("📱 Отправить контакт")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(contactBtn),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(chatID,
		"📝 <b>Регистрация</b>\n\n"+
			"Отправьте свою карточку контакта для завершения регистрации.\n\n"+
			"⚠️ <b>Внимание:</b> будет передан ваш номер телефона.")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard
	h.Bot.Send(msg)
}

func (h *Handler) HandleContact(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	contact := update.Message.Contact

	if contact.UserID != userID {
		msg := tgbotapi.NewMessage(chatID, "❌ <b>Ошибка</b>\nМожно отправить только свой контакт.")
		msg.ParseMode = "HTML"
		h.Bot.Send(msg)
		return
	}

	var existing models.User
	if err := database.DB.First(&existing, userID).Error; err == nil {
		msg := tgbotapi.NewMessage(chatID, "ℹ️ <b>Вы уже зарегистрированы</b>")
		msg.ParseMode = "HTML"
		h.Bot.Send(msg)
		return
	}

	phone := sanitize.String(contact.PhoneNumber)
	firstName := sanitize.Name(contact.FirstName)
	lastName := sanitize.Name(contact.LastName)

	user := models.User{
		ID:        userID,
		FirstName: firstName,
		LastName:  lastName,
		Username:  sanitize.Username(update.Message.From.UserName),
		Status:    models.StatusPending,
		Role:      models.RoleUser,
	}
	database.DB.Create(&user)

	msg := tgbotapi.NewMessage(chatID,
		"✅ <b>Регистрация завершена!</b>\n\n"+
			"📋 Ваша заявка передана администратору.\n"+
			"⏳ Ожидайте подтверждения.")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	h.Bot.Send(msg)

	h.notifyAdmins(&user, phone)
}

func (h *Handler) RegisterCancel(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	msg := tgbotapi.NewMessage(chatID, "❌ <b>Регистрация отменена</b>")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	h.Bot.Send(msg)
}

func (h *Handler) notifyAdmins(user *models.User, phone string) {
	text := fmt.Sprintf(
		"📥 <b>Новая заявка на регистрацию</b>\n\n"+
			"🆔 ID: <code>%d</code>\n"+
			"👤 Имя: <b>%s %s</b>\n"+
			"🔗 Username: @%s\n"+
			"📱 Телефон: <code>%s</code>",
		user.ID, user.FirstName, user.LastName, user.Username, phone)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Одобрить", fmt.Sprintf("approve_%d", user.ID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отклонить", fmt.Sprintf("reject_%d", user.ID)),
		),
	)

	var admins []models.User
	database.DB.Where("role = ? AND status = ?", models.RoleAdmin, models.StatusApproved).Find(&admins)

	if len(admins) == 0 {
		msg := tgbotapi.NewMessage(config.SuperAdminUID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboard
		h.Bot.Send(msg)
		return
	}

	for _, admin := range admins {
		msg := tgbotapi.NewMessage(admin.ID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboard
		h.Bot.Send(msg)
	}
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
