package handlers

import (
	"fmt"
	"telegram-bot/database"
	"telegram-bot/middleware"
	"telegram-bot/models"
	"telegram-bot/sanitize"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Profile(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ <b>Пользователь не найден</b>\nИспользуйте /start")
		msg.ParseMode = "HTML"
		h.Bot.Send(msg)
		return
	}

	text := fmt.Sprintf(
		"👤 <b>Ваш профиль</b>\n\n"+
			"🆔 ID: <code>%d</code>\n"+
			"👤 Имя: <b>%s %s</b>\n"+
			"🔗 Username: @%s\n"+
			"🎭 Роль: <b>%s</b>\n"+
			"📊 Статус: <b>%s</b>",
		user.ID, user.FirstName, user.LastName, user.Username,
		user.Role, user.Status)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	h.Bot.Send(msg)
}

func (h *Handler) ListPending(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if !h.Cfg.IsAdmin(userID) {
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			msg := tgbotapi.NewMessage(chatID, "⛔ <b>У вас нет прав администратора</b>")
			msg.ParseMode = "HTML"
			h.Bot.Send(msg)
			return
		}
		if !middleware.HasAdminAccess(userID, &user) {
			msg := tgbotapi.NewMessage(chatID, "⛔ <b>У вас нет прав администратора</b>")
			msg.ParseMode = "HTML"
			h.Bot.Send(msg)
			return
		}
	}

	var users []models.User
	database.DB.Where("status = ?", models.StatusPending).Find(&users)

	if len(users) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 <b>Нет ожидающих заявок</b>")
		msg.ParseMode = "HTML"
		h.Bot.Send(msg)
		return
	}

	for _, u := range users {
		text := fmt.Sprintf(
			"📥 <b>Заявка</b>\n\n"+
				"🆔 ID: <code>%d</code>\n"+
				"👤 Имя: <b>%s %s</b>\n"+
				"🔗 Username: @%s",
			u.ID, u.FirstName, u.LastName, u.Username)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Одобрить", fmt.Sprintf("approve_%d", u.ID)),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отклонить", fmt.Sprintf("reject_%d", u.ID)),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboard
		h.Bot.Send(msg)
	}
}

func (h *Handler) ListUsers(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if !h.Cfg.IsAdmin(userID) {
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			msg := tgbotapi.NewMessage(chatID, "⛔ <b>У вас нет прав администратора</b>")
			msg.ParseMode = "HTML"
			h.Bot.Send(msg)
			return
		}
		if !middleware.HasAdminAccess(userID, &user) {
			msg := tgbotapi.NewMessage(chatID, "⛔ <b>У вас нет прав администратора</b>")
			msg.ParseMode = "HTML"
			h.Bot.Send(msg)
			return
		}
	}

	var users []models.User
	database.DB.Find(&users)

	if len(users) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 <b>Нет зарегистрированных пользователей</b>")
		msg.ParseMode = "HTML"
		h.Bot.Send(msg)
		return
	}

	for _, u := range users {
		statusEmoji := "⏳"
		switch u.Status {
		case models.StatusApproved:
			statusEmoji = "✅"
		case models.StatusBlocked:
			statusEmoji = "🚫"
		}

		text := fmt.Sprintf(
			"%s <b>%s</b>\n"+
				"🆔 ID: <code>%d</code>\n"+
				"🎭 Роль: <b>%s</b>",
			statusEmoji, u.FirstName, u.ID, u.Role)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("👑 Админ", fmt.Sprintf("makeadmin_%d", u.ID)),
				tgbotapi.NewInlineKeyboardButtonData("👤 Юзер", fmt.Sprintf("removeadmin_%d", u.ID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🚫 Блок", fmt.Sprintf("block_%d", u.ID)),
				tgbotapi.NewInlineKeyboardButtonData("🔓 Разблок", fmt.Sprintf("unblock_%d", u.ID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", fmt.Sprintf("delete_%d", u.ID)),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboard
		h.Bot.Send(msg)
	}
}

func (h *Handler) HandleCallback(update *tgbotapi.Update) {
	callback := update.CallbackQuery
	adminID := callback.From.ID

	if !h.Cfg.IsAdmin(adminID) {
		var adminUser models.User
		if err := database.DB.First(&adminUser, adminID).Error; err != nil {
			h.answerCallback(callback.ID, "⛔ Нет прав")
			return
		}
		if !middleware.HasAdminAccess(adminID, &adminUser) {
			h.answerCallback(callback.ID, "⛔ Нет прав")
			return
		}
	}

	data := callback.Data
	if !sanitize.CallbackData(data) {
		h.answerCallback(callback.ID, "❓ Неверные данные")
		return
	}

	parts := strings.SplitN(data, "_", 2)
	action := parts[0]
	targetID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		h.answerCallback(callback.ID, "❌ Ошибка ID")
		return
	}

	var user models.User
	if err := database.DB.First(&user, targetID).Error; err != nil {
		h.answerCallback(callback.ID, "❌ Пользователь не найден")
		return
	}

	var response string
	switch action {
	case "approve":
		database.DB.Model(&user).Update("status", models.StatusApproved)
		response = fmt.Sprintf("✅ %s одобрен", user.FirstName)
		h.notifyUser(user.ID, "🎉 <b>Регистрация одобрена!</b>\nДобро пожаловать!")

	case "reject":
		database.DB.Model(&user).Update("status", models.StatusBlocked)
		response = fmt.Sprintf("❌ %s отклонён", user.FirstName)

	case "makeadmin":
		database.DB.Model(&user).Update("role", models.RoleAdmin)
		database.DB.Model(&user).Update("status", models.StatusApproved)
		response = fmt.Sprintf("👑 %s назначен админом", user.FirstName)

	case "removeadmin":
		database.DB.Model(&user).Update("role", models.RoleUser)
		response = fmt.Sprintf("👤 %s снят с админа", user.FirstName)

	case "block":
		database.DB.Model(&user).Update("status", models.StatusBlocked)
		response = fmt.Sprintf("🚫 %s заблокирован", user.FirstName)
		h.notifyUser(user.ID, "🚫 <b>Аккаунт заблокирован</b>\nАдминистратором.")

	case "unblock":
		database.DB.Model(&user).Update("status", models.StatusApproved)
		response = fmt.Sprintf("🔓 %s разблокирован", user.FirstName)
		h.notifyUser(user.ID, "🔓 <b>Аккаунт разблокирован</b>")

	case "delete":
		database.DB.Unscoped().Delete(&user)
		response = fmt.Sprintf("🗑 %s удалён", user.FirstName)
	}

	h.answerCallback(callback.ID, response)
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, response)
	editMsg.ParseMode = "HTML"
	h.Bot.Send(editMsg)
}

func (h *Handler) answerCallback(callbackID, text string) {
	callback := tgbotapi.NewCallback(callbackID, text)
	h.Bot.Send(callback)
}

func (h *Handler) notifyUser(userID int64, text string) {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ParseMode = "HTML"
	h.Bot.Send(msg)
}
