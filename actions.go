package robonaut

import (
	"time"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
)

type ActionSpeak struct {
	Delay   time.Duration
	Channel models.Channel
	Message string
}

func (a ActionSpeak) Do(bot *Robonaut) error {
	if time.Now().Before(bot.MessageLastSent.Add(2 * time.Second)) {
		time.Sleep(2 * time.Second)
	}

	if a.Delay > 0 {
		time.Sleep(a.Delay)
	}

	_, err := bot.Client.SendMessage(&a.Channel, a.Message)
	if err != nil {
		return err
	}

	return nil
}

type ActionReact struct {
	Message  models.Message
	Reaction string
}

func (a ActionReact) Do(bot *Robonaut) error {
	err := bot.Client.ReactToMessage(&a.Message, a.Reaction)
	if err != nil {
		return err
	}

	return nil
}
