package robonaut

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/rest"
)

var defaultPassword = "pass@word1"
var defaultDomain = "test.com"

// Robonaut main bot object
type Robonaut struct {
	Name            string
	URL             string
	DebugMode       bool
	Client          *realtime.Client
	Rest            *rest.Client
	Self            *models.User
	MsgChannel      chan models.Message
	MessageListener func(models.Message) error
	MessageLastSent time.Time
	stop            chan bool
	action          chan RobonautAction
	data            RobonautDataStore
}

type RobonautDataStore struct {
	Settings             []models.Setting
	MyChannels           []models.Channel
	Permissions          []models.Permission
	ChannelSubscriptions []models.ChannelSubscription
}

type RobonautAction interface {
	Do(*Robonaut) error
}

type RobonautSpeak struct {
	Type string
}

// SpawnRobonaut Creates a new instance of the bot
func SpawnRobonaut(name string, url string, debug bool) *Robonaut {
	return &Robonaut{Name: name, URL: url, data: RobonautDataStore{}, DebugMode: debug}
}

// Start connects the ddp client
func (r *Robonaut) Start() error {

	serverUrl, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	c, err := realtime.NewClient(serverUrl, r.DebugMode)
	if err != nil {
		log.Println("Failed to connect", err)
		return err
	}

	r.Client = c

	cR := rest.NewClient(serverUrl, r.DebugMode)

	r.Rest = cR

	r.MsgChannel = make(chan models.Message, 100)
	r.action = make(chan RobonautAction, 100)
	r.stop = make(chan bool)

	return nil
}

func (r *Robonaut) Reconnect() {
	r.Client.Reconnect()
}

// EventLoop blocks until a message on stop channel
func (r *Robonaut) EventLoop() {
	go r.processActionQueue()
	go r.processMessageQueue()

	<-r.stop
}

func (r *Robonaut) processMessageQueue() {
	for {
		message := <-r.MsgChannel
		if r.MessageListener != nil {
			// TODO send errors to error handler
			if err := r.MessageListener(message); err != nil {
				log.Println("ERROR:", err)
			}
		}
	}
}

func (r *Robonaut) processActionQueue() {
	for {
		action := <-r.action
		log.Println(r.Name, "Processing Action: ", action)
		err := action.Do(r)
		if err != nil {
			log.Println(err)
		}
	}
}

// Stop Closes out the ddp connection
func (r *Robonaut) Stop() {
	r.stop <- true
	r.Client.Close()
}

// Register Creates an account on the Rocket.Chat server
func (r *Robonaut) Register() error {
	// Register an existing user
	user, err := r.Client.RegisterUser(&models.UserCredentials{Email: fmt.Sprintf("%s@%s", r.Name, defaultDomain), Name: r.Name, Password: defaultPassword})
	if err != nil {
		log.Println("Failed to login", err)
		return err
	}

	r.Self = user

	return err
}

// Login logs into account
func (r *Robonaut) Login(email string, password string) error {
	// Login an existing user
	user, err := r.Client.Login(&models.UserCredentials{Email: email, Password: password})
	if err != nil {
		log.Println("Failed to login", err)
		return err
	}

	r.Self = user

	r.Client.ConnectionOnline()

	if err := r.Rest.Login(&models.UserCredentials{ID: user.ID, Token: user.Token}); err != nil {
		log.Println("Failed to login rest", err)
		return err
	}

	return err
}

func (r *Robonaut) LoginOrRegister() error {
	if err := r.Login(fmt.Sprintf("%s@%s", r.Name, defaultDomain), defaultPassword); err != nil {
		log.Println("Unable to login")
		if err := r.Register(); err != nil {
			log.Println(fmt.Sprintf("Failed to register %v.  \nError: %v", r.Name, err))
			return err
		}
	}

	return nil
}

func (r *Robonaut) AddConnectionStatusListener(listener func(int)) {
	r.Client.AddStatusListener(listener)
}

func (r *Robonaut) GetInitialData() error {
	myChannels, err := r.Client.GetChannelsIn()
	if err != nil {
		return err
	}

	r.data.MyChannels = myChannels

	settings, err := r.Client.GetPublicSettings()
	if err != nil {
		return err
	}

	r.data.Settings = settings

	r.Client.GetUserRoles()

	permissions, err := r.Client.GetPermissions()
	if err != nil {
		return err
	}

	r.data.Permissions = permissions

	// listCustomSounds
	// listEmojiCustom

	channelSubscriptions, err := r.Client.GetChannelSubscriptions()
	if err != nil {
		return nil
	}

	r.data.ChannelSubscriptions = channelSubscriptions

	return nil
}

func (r *Robonaut) Subscribe(sub string, args string) {
	r.Client.Sub(sub, args)
}

func (r *Robonaut) ListenToMyMessages() error {
	r.Client.SubscribeToMyMessages(r.MsgChannel)

	return nil
}

// EstablishSync subscribes to all connections needed
func (r *Robonaut) EstablishSync() error {
	// Subscribe to stuff
	//	r.Client.Sub("meteor.loginServiceConfiguration")
	//	r.Client.Sub("meteor_autoupdate_clientVersions")

	//	r.Client.Sub("roles")
	//	r.Client.Sub("userData")
	//	r.Client.Sub("activeUsers")

	//	r.Client.Sub("stream-notify-logged", "roles-changed")
	//	r.Client.Sub("stream-notify-logged", "Users:NameChanged")
	//	r.Client.Sub("stream-notify-logged", "updatedAvatar")

	//	r.Client.Sub("stream-notify-all", "public-settings-changed")
	//	r.Client.Sub("stream-notify-all", "updateCustomSound")
	//	r.Client.Sub("stream-notify-all", "deleteCustomSound")

	//	r.Client.Sub("stream-notify-logged", "updateEmojiCustom")
	//	r.Client.Sub("stream-notify-logged", "deleteEmojiCustom")

	r.Client.Sub("stream-notify-user", fmt.Sprintf("%s/message", r.Self.ID))
	//	r.Client.Sub("stream-notify-user", fmt.Sprintf("%s/otr", r.Self.ID))
	//	r.Client.Sub("stream-notify-user", fmt.Sprintf("%s/webrtc", r.Self.ID))
	r.Client.Sub("stream-notify-user", fmt.Sprintf("%s/notification", r.Self.ID))

	r.Client.Sub("stream-notify-user", fmt.Sprintf("%s/rooms-changed", r.Self.ID))
	r.Client.Sub("stream-notify-user", fmt.Sprintf("%s/subscriptions-changed", r.Self.ID))

	// robo.Client.GetChannelRoles(roomId)
	// robo.Client.LoadHistory(roomId)

	//r.Client.Sub("stream-notify-room", fmt.Sprintf("%s/deleteMessage", roomId))
	//r.Client.Sub("stream-notify-room", fmt.Sprintf("%s/typing", roomId))

	return nil
}

// ListenToCommsChannel subscribes to channel for new messages
func (r *Robonaut) ListenToCommsChannel(channel *models.Channel) error {
	err := r.Client.SubscribeToMessageStream(channel, r.MsgChannel)
	if err != nil {
		return err
	}

	return nil
}

// Speak sends a message to a channel
func (r *Robonaut) Speak(channel *models.Channel, message string) error {
	r.action <- ActionSpeak{Channel: *channel, Message: message}
	log.Println("Speak added to Action Queue")

	return nil
}
