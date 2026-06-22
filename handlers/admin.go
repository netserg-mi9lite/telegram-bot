package handlers

import (
	"fmt"
	"telegram-bot/database"
	"telegram-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Profile(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Пользователь не найден. Используйте /start")
		h.Bot.Send(msg)
		return
	}

	text := fmt.Sprintf("👤 Ваш профиль\n\n"+
		"ID: %d\n"+
		"Имя: %s %s\n"+
		"Username: @%s\n"+
		"Роль: %s\n"+
		"Статус: %s",
		user.ID, user.FirstName, user.LastName, user.Username,
		user.Role, user.Status)

	msg := tgbotapi.NewMessage(chatID, text)
	h.Bot.Send(msg)
}

func (h *Handler) ListPending(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if !h.Cfg.IsAdmin(userID) {
		msg := tgbotapi.NewMessage(chatID, "⛔ У вас нет прав администратора.")
		h.Bot.Send(msg)
		return
	}

	var users []models.User
	database.DB.Where("status = ?", models.StatusPending).Find(&users)

	if len(users) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 Нет ожидающих заявок.")
		h.Bot.Send(msg)
		return
	}

	for _, u := range users {
		text := fmt.Sprintf("📥 Заявка\n\nID: %d\nИмя: %s %s\nUsername: @%s",
			u.ID, u.FirstName, u.LastName, u.Username)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Одобрить", fmt.Sprintf("approve_%d", u.ID)),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отклонить", fmt.Sprintf("reject_%d", u.ID)),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		h.Bot.Send(msg)
	}
}

func (h *Handler) ListUsers(update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	if !h.Cfg.IsAdmin(userID) {
		msg := tgbotapi.NewMessage(chatID, "⛔ У вас нет прав администратора.")
		h.Bot.Send(msg)
		return
	}

	var users []models.User
	database.DB.Find(&users)

	if len(users) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 Нет зарегистрированных пользователей.")
		h.Bot.Send(msg)
		return
	}

	text := "👥 Все пользователи:\n\n"
	for _, u := range users {
		statusEmoji := "⏳"
		switch u.Status {
		case models.StatusApproved:
			statusEmoji = "✅"
		case models.StatusBlocked:
			statusEmoji = "🚫"
		}

		text += fmt.Sprintf("%s %s (ID: %d) - %s\n", statusEmoji, u.FirstName, u.ID, u.Role)

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
		msg.ReplyMarkup = keyboard
		h.Bot.Send(msg)
		text = ""
	}
}

func (h *Handler) HandleCallback(update *tgbotapi.Update) {
	callback := update.CallbackQuery
	adminID := callback.From.ID

	if !h.Cfg.IsAdmin(adminID) {
		h.answerCallback(callback.ID, "⛔ Нет прав")
		return
	}

	var action, targetID string
	data := callback.Data

	switch {
	case len(data) > 8 && data[:8] == "approve_":
		action = "approve"
		targetID = data[8:]
	case len(data) > 7 && data[:7] == "reject_":
		action = "reject"
		targetID = data[7:]
	case len(data) > 10 && data[:10] == "makeadmin_":
		action = "makeadmin"
		targetID = data[10:]
	case len(data) > 12 && data[:12] == "removeadmin_":
		action = "removeadmin"
		targetID = data[12:]
	case len(data) > 6 && data[:6] == "block_":
		action = "block"
		targetID = data[6:]
	case len(data) > 8 && data[:8] == "unblock_":
		action = "unblock"
		targetID = data[8:]
	case len(data) > 7 && data[:7] == "delete_":
		action = "delete"
		targetID = data[7:]
	default:
		h.answerCallback(callback.ID, "❓ Неизвестное действие")
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
		h.notifyUser(user.ID, "🎉 Ваша регистрация одобрена! Добро пожаловать!")

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
		h.notifyUser(user.ID, "🚫 Ваш аккаунт заблокирован администратором.")

	case "unblock":
		database.DB.Model(&user).Update("status", models.StatusApproved)
		response = fmt.Sprintf("🔓 %s разблокирован", user.FirstName)
		h.notifyUser(user.ID, "🔓 Ваш аккаунт разблокирован.")

	case "delete":
		database.DB.Unscoped().Delete(&user)
		response = fmt.Sprintf("🗑 %s удалён", user.FirstName)
	}

	h.answerCallback(callback.ID, response)
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, response)
	h.Bot.Send(editMsg)
}

func (h *Handler) answerCallback(callbackID, text string) {
	callback := tgbotapi.NewCallback(callbackID, text)
	h.Bot.Send(callback)
}

func (h *Handler) notifyUser(userID int64, text string) {
	msg := tgbotapi.NewMessage(userID, text)
	h.Bot.Send(msg)
}
