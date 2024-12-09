package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	convo "github.com/adellantado/go-telegram-convo"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var userdata map[int64]map[string]string = make(map[int64]map[string]string)

var convoManager *convo.ConversationManager

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handlerDefault),
		bot.WithMessageTextHandler("/form", bot.MatchTypeExact, handlerForm),
		bot.WithMessageTextHandler("/cancel", bot.MatchTypeExact, handlerCancel),
	}

	// os.Getenv("EXAMPLE_TELEGRAM_BOT_TOKEN")
	b, err := bot.New("7403989705:AAERGGkA5pnVJrESAL15PbKpLI9Ms7R6XBg", opts...)
	if err != nil {
		panic(err)
	}

	convoManager = convo.NewConversationManager()
	convoManager.AddConvoHandlers(map[string][]func(context.Context, *bot.Bot, *models.Update) string{
		"formConvo": {
			callbackAskAge,
			callbackFinish,
		},
	})

	fmt.Println("The bot is running! Press Ctrl+C to terminate!")
	b.Start(ctx)
}

func handlerDefault(ctx context.Context, b *bot.Bot, update *models.Update) {
	if convoManager.Handle(ctx, b, update) {
		return
	}

	chatID := update.Message.Chat.ID
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Type /form to start the form",
	})
}

func handlerCancel(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	convoManager.StopConvo(chatID, "formConvo")

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Canceled",
	})
}

func handlerForm(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	userdata[chatID] = make(map[string]string)

	b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Let's start the form! Type /cancel to cancel\nWhat's your name? (at least 2 characters)",
	})
	convoManager.InitConvo(chatID, "formConvo")
}

func callbackAskAge(ctx context.Context, b *bot.Bot, update *models.Update) string {
	chatID := update.Message.Chat.ID

	if len(update.Message.Text) < 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Please enter a valid name, at least 2 characters",
		})
		return "0"
	}

	userdata[chatID]["name"] = update.Message.Text

	b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "How old are you? (between 18 and 100)",
	})

	return "1"
}

func callbackFinish(ctx context.Context, b *bot.Bot, update *models.Update) string {
	chatID := update.Message.Chat.ID

	age, errAge := strconv.Atoi(update.Message.Text)
	if errAge != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Please enter a valid age",
		})
		return "1"
	}

	if age < 18 || age > 100 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Please enter an age between 18 and 100",
		})
		return "1"
	}

	b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text: fmt.Sprintf("Thank you!\nName: %s\nAge: %d",
			bot.EscapeMarkdown(userdata[chatID]["name"]), age),
	})

	return convo.END
}
