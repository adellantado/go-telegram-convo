package convo

import (
	"context"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	END string = "end"
)

type ConversationManager struct {
	registeredConvos map[string]*Conversation
	activeConvos     map[int64]*Conversation
}

func NewConversationManager() *ConversationManager {
	return &ConversationManager{
		registeredConvos: make(map[string]*Conversation),
		activeConvos:     make(map[int64]*Conversation),
	}
}

func (cm *ConversationManager) AddConvo(convoName string, convo *Conversation) {
	cm.registeredConvos[convoName] = convo
}

func (cm *ConversationManager) AddConvoHandlers(convos map[string][]func(context.Context, *bot.Bot, *models.Update) string) {
	for convoName, hadlers := range convos {
		newConvo := NewConversation("0")
		for i, hadler := range hadlers {
			newConvo.AddHandler(strconv.Itoa(i), hadler)
		}
		cm.AddConvo(convoName, newConvo)
	}
}

func (cm *ConversationManager) InitConvo(chatID int64, convoName string) {
	activeConvo, exists := cm.registeredConvos[convoName]
	if exists {
		cm.activeConvos[chatID] = activeConvo
	}
}

func (cm *ConversationManager) StopConvo(chatID int64, convoName string) {
	activeConvo, exists := cm.registeredConvos[convoName]
	if exists {
		activeConvo.ResetState(chatID)
		delete(cm.activeConvos, chatID)
	}
}

func (cm *ConversationManager) Handle(ctx context.Context, b *bot.Bot, update *models.Update) bool {
	chatID := update.Message.Chat.ID
	activeConvo, exists := cm.activeConvos[chatID]
	if exists {
		state := activeConvo.HandleUpdate(ctx, b, update)
		if state == END {
			delete(cm.activeConvos, chatID)
		}
		return true
	}
	return false
}

type Conversation struct {
	stateHandlers map[string]func(context.Context, *bot.Bot, *models.Update) string
	userStates    map[int64]string
	defaultState  string
}

func NewConversation(defaultState string) *Conversation {
	return &Conversation{
		stateHandlers: make(map[string]func(context.Context, *bot.Bot, *models.Update) string),
		userStates:    make(map[int64]string),
		defaultState:  defaultState,
	}
}

func (c *Conversation) AddHandler(state string, handler func(context.Context, *bot.Bot, *models.Update) string) {
	c.stateHandlers[state] = handler
}

func (c *Conversation) HandleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) string {
	chatID := update.Message.Chat.ID
	currentState, exists := c.userStates[chatID]
	if !exists {
		currentState = c.defaultState
	}

	handler, handlerExists := c.stateHandlers[currentState]
	if !handlerExists {
		c.userStates[chatID] = c.defaultState
		return END
	}

	nextState := handler(ctx, b, update)

	if nextState == END {
		c.userStates[chatID] = c.defaultState
	} else {
		c.userStates[chatID] = nextState
	}
	return nextState
}

func (c *Conversation) ResetState(chatID int64) {
	c.userStates[chatID] = c.defaultState
}
