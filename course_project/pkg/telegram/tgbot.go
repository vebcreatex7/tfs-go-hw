package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tfs-go-hw/course_project/internal/domain"
)

type TgBot struct {
	bot    *tgbotapi.BotAPI
	chatId int64
}

func NewBot(token string) (*TgBot, error) {

	tg := &TgBot{}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	tg.bot = bot
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		// Ignore any non-Message updates
		if update.Message == nil {
			continue
		}
		// Ignore any non-command Messages.
		if !update.Message.IsCommand() {
			continue
		}
		if update.Message.Command() == "start" {
			tg.chatId = update.Message.Chat.ID
			break
		}
	}
	return tg, nil

}

func (tb *TgBot) SendOrder(order domain.RecordOrder) error {
	str := order.TS.String() + "\n" +
		order.Symbol + "\n" +
		order.Side + "\n" +
		fmt.Sprintf("%d", order.Size) + "\n" +
		fmt.Sprintf("%f", order.Price)
	msg := tgbotapi.NewMessage(tb.chatId, str)
	_, err := tb.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
