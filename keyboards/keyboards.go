package keyboards

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func MainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Профиль"),
		),
	)
}

func AdminKeyboard() tgbotapi.ReplyKeyboardMarkup {
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

func PendingRequestInline(userID int64) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Одобрить", fmt.Sprintf("approve_%d", userID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отклонить", fmt.Sprintf("reject_%d", userID)),
		),
	)
}

func UserActionInline(userID int64) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👑 Назначить админом", fmt.Sprintf("makeadmin_%d", userID)),
			tgbotapi.NewInlineKeyboardButtonData("👤 Снять админа", fmt.Sprintf("removeadmin_%d", userID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚫 Заблокировать", fmt.Sprintf("block_%d", userID)),
			tgbotapi.NewInlineKeyboardButtonData("🔓 Разблокировать", fmt.Sprintf("unblock_%d", userID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить аккаунт", fmt.Sprintf("delete_%d", userID)),
		),
	)
}
