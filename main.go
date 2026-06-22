package main

import (
	"fmt"
	"log"
	"telegram-bot/config"
	"telegram-bot/database"
	"telegram-bot/handlers"
	"telegram-bot/middleware"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	database.Init(cfg)

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}

	fmt.Printf("Bot authorized on account: %s\n", bot.Self.UserName)

	h := handlers.NewHandler(bot, cfg)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			h.HandleCallback(&update)
			continue
		}

		if update.Message == nil {
			continue
		}

		user := middleware.EnsureUserExists(cfg, &update)
		if user == nil {
			continue
		}

		if !middleware.HasAccess(user) && update.Message.Command() != "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⛔ Нет доступа. Ожидайте подтверждения.")
			bot.Send(msg)
			continue
		}

		switch update.Message.Command() {
		case "start":
			h.Start(&update)
		default:
			switch update.Message.Text {
			case "📋 Профиль":
				h.Profile(&update)
			case "👥 Управление":
				h.ListUsers(&update)
			case "📥 Заявки":
				h.ListPending(&update)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Используйте /start или кнопки меню.")
				bot.Send(msg)
			}
		}
	}
}
