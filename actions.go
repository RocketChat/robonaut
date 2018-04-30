package robonaut

import (
	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
)

type ActionSpeak struct {
	Channel models.Channel
	Message string
}

func (a ActionSpeak) Do(bot *Robonaut) error {
	_, err := bot.Client.SendMessage(&a.Channel, a.Message)
	if err != nil {
		return err
	}

	return nil
}
