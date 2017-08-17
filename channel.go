package robonaut

import "github.com/RocketChat/Rocket.Chat.Go.SDK/models"

// GetChannels gets bots channels
func (r *Robonaut) GetChannels() ([]models.Channel, error) {
	channels, err := r.Client.GetChannelsIn()
	if err != nil {
		return nil, err
	}

	return channels, nil
}

// CreateChannel Creates a channel
func (r *Robonaut) CreateChannel(channel *models.Channel) error {
	if err := r.Client.CreateChannel(channel.Name, []string{}); err != nil {
		return err
	}

	return nil
}
