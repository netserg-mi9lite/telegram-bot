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

	sendStartupMessage(bot)

	h := handlers.NewHandler(bot, cfg)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			data := update.CallbackQuery.Data
			if data == "register_start" {
				fakeUpdate := tgbotapi.Update{
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: update.CallbackQuery.Message.Chat.ID},
						From: update.CallbackQuery.From,
					},
				}
				h.RegisterStart(&fakeUpdate)
				answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
				bot.Send(answer)
			} else {
				h.HandleCallback(&update)
			}
			continue
		}

		if update.Message == nil {
			continue
		}

		if update.Message.Contact != nil {
			h.HandleContact(&update)
			continue
		}

		user := middleware.EnsureUserExists(cfg, &update)
		if user == nil {
			continue
		}

		cmd := update.Message.Command()
		text := update.Message.Text

		if h.Cfg.IsAdmin(user.ID) {
			switch cmd {
			case "start":
				h.Start(&update)
			default:
				switch text {
				case "📋 Профиль":
					h.Profile(&update)
				case "👥 Управление":
					h.ListUsers(&update)
				case "📥 Заявки":
					h.ListPending(&update)
				default:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "💡 Используйте /start или кнопки меню.")
					bot.Send(msg)
				}
			}
			continue
		}

		switch cmd {
		case "start":
			h.Start(&update)
		case "register":
			h.RegisterStart(&update)
		default:
			if text == "❌ Отмена" {
				h.RegisterCancel(&update)
				continue
			}
			if !middleware.HasAccess(user) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"⛔ <b>Нет доступа</b>\n\n"+
						"Используйте /register для регистрации.")
				msg.ParseMode = "HTML"
				bot.Send(msg)
				continue
			}
			switch text {
			case "📋 Профиль":
				h.Profile(&update)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "💡 Используйте /start или кнопки меню.")
				bot.Send(msg)
			}
		}
	}
}

func sendStartupMessage(bot *tgbotapi.BotAPI) {
	text := fmt.Sprintf(
		"🚀 <b>%s v%s</b>\n\n"+
			"📡 Сервер запущен\n"+
			"🌐 IP: <code>%s</code>\n"+
			"✅ Статус: <b>online</b>",
		config.AppName, config.AppVersion,
		config.GetServerIP(),
	)

	msg := tgbotapi.NewMessage(config.SuperAdminUID, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}
