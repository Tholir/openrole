// Package plugin provides the WASM plugin runtime for openrole using GopherJS.
// This file defines the JavaScript plugin interface contracts.
package plugin

import (
	"context"
	"fmt"
)

// Plugin represents the main interface that all openrole plugins must implement.
// Plugins are compiled Go code that executes in the browser via WebAssembly.
type Plugin interface {
	// Info returns plugin metadata.
	Info() PluginInfo

	// Init is called once when the plugin is first loaded.
	Init(ctx context.Context) error

	// RegisterHooks registers all plugin hooks.
	RegisterHooks(ctx context.Context) error

	// Close is called when the plugin is unloaded.
	Close(ctx context.Context) error
}

// PluginInfo contains metadata about a plugin.
type PluginInfo struct {
	// ID is the unique identifier for this plugin.
	ID string `json:"id"`
	// Name is the human-readable name.
	Name string `json:"name"`
	// Version is the plugin version in semver format.
	Version string `json:"version"`
	// Description describes what the plugin does.
	Description string `json:"description"`
	// Author is the plugin author.
	Author string `json:"author"`
	// Tags are search keywords for the plugin.
	Tags []string `json:"tags"`
	// Homepage is the plugin's website URL.
	Homepage string `json:"homepage,omitempty"`
}

// Tool represents a plugin-provided tool that can be called by the agent.
type Tool struct {
	// Name is the unique identifier for this tool within the plugin.
	Name string `json:"name"`
	// Description explains what the tool does.
	Description string `json:"description"`
	// InputSchema defines the expected JSON schema for tool input.
	InputSchema ToolSchema `json:"input_schema"`
	// OutputSchema defines the JSON schema for tool output.
	OutputSchema ToolSchema `json:"output_schema,omitempty"`
}

// ToolSchema represents a JSON schema for tool input/output.
type ToolSchema struct {
	// Type is the JSON type (object, string, number, boolean, array).
	Type string `json:"type"`
	// Properties define the fields for object types.
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	// Required lists required field names.
	Required []string `json:"required,omitempty"`
	// Description provides documentation.
	Description string `json:"description,omitempty"`
}

// PropertySchema describes a single property in a schema.
type PropertySchema struct {
	// Type is the JSON type of this property.
	Type string `json:"type"`
	// Description explains this property.
	Description string `json:"description,omitempty"`
	// Enum lists allowed values.
	Enum []interface{} `json:"enum,omitempty"`
	// Default provides a default value.
	Default interface{} `json:"default,omitempty"`
}

// ToolResult is the result of a tool call.
type ToolResult struct {
	// Success indicates whether the tool executed successfully.
	Success bool `json:"success"`
	// Output contains the tool's output data.
	Output interface{} `json:"output,omitempty"`
	// Error contains error information if Success is false.
	Error *ToolError `json:"error,omitempty"`
}

// ToolError describes a tool execution error.
type ToolError struct {
	// Code is a machine-readable error code.
	Code string `json:"code"`
	// Message is the human-readable error message.
	Message string `json:"message"`
	// Details contains additional error context.
	Details interface{} `json:"details,omitempty"`
}

// ToolHandler is a function that handles tool calls.
type ToolHandler func(ctx context.Context, input map[string]interface{}) ToolResult

// EventHook is a function that handles events.
type EventHook func(ctx context.Context, event Event) error

// Event represents an event in the openrole system.
type Event struct {
	// Type identifies the event type.
	Type EventType `json:"type"`
	// SessionID is the session that generated the event.
	SessionID string `json:"session_id"`
	// Data contains event-specific payload.
	Data interface{} `json:"data,omitempty"`
	// Timestamp is when the event occurred.
	Timestamp int64 `json:"timestamp"`
}

// EventType identifies the type of an event.
type EventType string

const (
	// EventSessionStarted is fired when a new session begins.
	EventSessionStarted EventType = "session.started"
	// EventSessionEnded is fired when a session ends.
	EventSessionEnded EventType = "session.ended"
	// EventMessageSent is fired when a message is sent.
	EventMessageSent EventType = "message.sent"
	// EventMessageReceived is fired when a message is received.
	EventMessageReceived EventType = "message.received"
	// EventToolCalled is fired when a tool is invoked.
	EventToolCalled EventType = "tool.called"
	// EventToolResult is fired when a tool returns.
	EventToolResult EventType = "tool.result"
	// EventError is fired when an error occurs.
	EventError EventType = "error"
	// EventVoteRequested is fired when a vote is requested.
	EventVoteRequested EventType = "vote.requested"
	// EventVoteCompleted is fired when a vote is completed.
	EventVoteCompleted EventType = "vote.completed"
	// EventDiceRolled is fired when dice are rolled.
	EventDiceRolled EventType = "dice.rolled"
	// EventCharacterSheetUpdated is fired when a character sheet changes.
	EventCharacterSheetUpdated EventType = "character.sheet.updated"
)

// ChatMessage represents a message in the chat.
type ChatMessage struct {
	// ID is the unique message identifier.
	ID string `json:"id"`
	// Role is the sender role (user, assistant, system, tool).
	Role MessageRole `json:"role"`
	// Content is the message content.
	Content string `json:"content"`
	// Metadata contains additional message data.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MessageRole identifies the role of a message sender.
type MessageRole string

const (
	// RoleUser represents a human user.
	RoleUser MessageRole = "user"
	// RoleAssistant represents the AI assistant.
	RoleAssistant MessageRole = "assistant"
	// RoleSystem represents a system message.
	RoleSystem MessageRole = "system"
	// RoleTool represents a tool message.
	RoleTool MessageRole = "tool"
)

// ChatHeaders contains metadata about a chat request.
type ChatHeaders struct {
	// SessionID is the current session.
	SessionID string `json:"session_id"`
	// Model is the model being used.
	Model string `json:"model"`
	// Provider is the provider being used.
	Provider string `json:"provider"`
	// Timestamp is when the request was made.
	Timestamp int64 `json:"timestamp"`
	// CustomHeaders contains plugin-specific headers.
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// ChatParams contains parameters for a chat request.
type ChatParams struct {
	// Messages are the conversation messages.
	Messages []ChatMessage `json:"messages"`
	// SystemPrompt is the system prompt.
	SystemPrompt string `json:"system_prompt,omitempty"`
	// Temperature controls randomness.
	Temperature float64 `json:"temperature,omitempty"`
	// MaxTokens limits response length.
	MaxTokens int `json:"max_tokens,omitempty"`
	// Tools are available tools.
	Tools []Tool `json:"tools,omitempty"`
	// Stop is custom stop sequences.
	Stop []string `json:"stop,omitempty"`
}

// ChatResponse is the response from a chat completion.
type ChatResponse struct {
	// Message is the generated message.
	Message ChatMessage `json:"message"`
	// FinishReason explains why the response ended.
	FinishReason string `json:"finish_reason,omitempty"`
	// Usage contains token usage information.
	Usage *Usage `json:"usage,omitempty"`
}

// Usage contains token usage statistics.
type Usage struct {
	// PromptTokens is tokens used in the prompt.
	PromptTokens int `json:"prompt_tokens"`
	// CompletionTokens is tokens in the completion.
	CompletionTokens int `json:"completion_tokens"`
	// TotalTokens is total tokens used.
	TotalTokens int `json:"total_tokens"`
}

// VotePrompt represents a voting prompt shown to users.
type VotePrompt struct {
	// ID is the unique vote identifier.
	ID string `json:"id"`
	// Question is the voting question.
	Question string `json:"question"`
	// Options are the available choices.
	Options []VoteOption `json:"options"`
	// MultiSelect allows multiple selections.
	MultiSelect bool `json:"multi_select"`
	// Deadline is when voting closes (0 for no deadline).
	Deadline int64 `json:"deadline,omitempty"`
	// Metadata contains additional context.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// VoteOption represents a single voting option.
type VoteOption struct {
	// ID is the option identifier.
	ID string `json:"id"`
	// Label is the display text.
	Label string `json:"label"`
	// Description explains the option.
	Description string `json:"description,omitempty"`
}

// VoteResult contains the outcome of a vote.
type VoteResult struct {
	// VoteID is the vote that was conducted.
	VoteID string `json:"vote_id"`
	// SelectedOptions are the chosen option IDs.
	SelectedOptions []string `json:"selected_options"`
	// Tally is the raw vote count per option.
	Tally map[string]int `json:"tally"`
	// WinnerIDs are the winning option IDs.
	WinnerIDs []string `json:"winner_ids"`
}

// DiceRoll represents a dice roll request.
type DiceRoll struct {
	// ID is the unique roll identifier.
	ID string `json:"id"`
	// Formula is the dice notation (e.g., "2d6+3").
	Formula string `json:"formula"`
	// Reason is why dice are being rolled.
	Reason string `json:"reason,omitempty"`
	// Metadata contains additional context.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DiceResult contains the outcome of a dice roll.
type DiceResult struct {
	// RollID is the original roll request ID.
	RollID string `json:"roll_id"`
	// Total is the sum of all dice.
	Total int `json:"total"`
	// Individual are the individual die results.
	Individual []int `json:"individual"`
	// Modifiers are any applied modifiers.
	Modifiers int `json:"modifiers"`
	// Rolls are the raw roll data.
	Rolls []DiceRollDetail `json:"rolls"`
}

// DiceRollDetail contains details about a single dice roll.
type DiceRollDetail struct {
	// Faces is the number of faces on the die.
	Faces int `json:"faces"`
	// Result is the rolled value.
	Result int `json:"result"`
	// Kept indicates if this die was kept (for advantage/disadvantage).
	Kept bool `json:"kept"`
}

// CharacterSheet represents a TTRPG character sheet.
type CharacterSheet struct {
	// ID is the character identifier.
	ID string `json:"id"`
	// Name is the character name.
	Name string `json:"name"`
	// Class is the character class.
	Class string `json:"class"`
	// Level is the character level.
	Level int `json:"level"`
	// Race is the character race.
	Race string `json:"race"`
	// Background is the character background.
	Background string `json:"background"`
	// Alignment is the character alignment.
	Alignment string `json:"alignment"`
	// Stats are the ability scores.
	Stats CharacterStats `json:"stats"`
	// HP is current and max hit points.
	HP CharacterHP `json:"hp"`
	// AC is armor class.
	AC int `json:"ac"`
	// Speed is movement speed in feet.
	Speed int `json:"speed"`
	// Skills are the character's skills.
	Skills []Skill `json:"skills"`
	// Inventory is the character's items.
	Inventory []Item `json:"inventory"`
	// Notes is free-form notes.
	Notes string `json:"notes,omitempty"`
	// Metadata contains additional data.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CharacterStats contains ability scores.
type CharacterStats struct {
	// Strength score.
	STR int `json:"str"`
	// Dexterity score.
	DEX int `json:"dex"`
	// Constitution score.
	CON int `json:"con"`
	// Intelligence score.
	INT int `json:"int"`
	// Wisdom score.
	WIS int `json:"wis"`
	// Charisma score.
	CHA int `json:"cha"`
}

// Modifier returns the ability modifier for a score.
func (s CharacterStats) Modifier(stat string) int {
	var score int
	switch stat {
	case "str":
		score = s.STR
	case "dex":
		score = s.DEX
	case "con":
		score = s.CON
	case "int":
		score = s.INT
	case "wis":
		score = s.WIS
	case "cha":
		score = s.CHA
	default:
		return 0
	}
	return (score - 10) / 2
}

// CharacterHP contains hit point information.
type CharacterHP struct {
	// Current is the current HP.
	Current int `json:"current"`
	// Max is the maximum HP.
	Max int `json:"max"`
	// Temp is temporary HP.
	Temp int `json:"temp"`
}

// Skill represents a character skill.
type Skill struct {
	// Name is the skill name.
	Name string `json:"name"`
	// Ability is the associated ability score.
	Ability string `json:"ability"`
	// Proficient indicates proficiency.
	Proficient bool `json:"proficient"`
	// Expertise indicates expertise (double proficiency).
	Expertise bool `json:"expertise"`
	// Bonus is any additional bonus.
	Bonus int `json:"bonus"`
}

// Item represents an inventory item.
type Item struct {
	// Name is the item name.
	Name string `json:"name"`
	// Quantity is how many.
	Quantity int `json:"quantity"`
	// Weight is the weight in pounds.
	Weight float64 `json:"weight,omitempty"`
	// Description describes the item.
	Description string `json:"description,omitempty"`
	// Equipped indicates if currently equipped.
	Equipped bool `json:"equipped"`
	// Attuned indicates if attuned.
	Attuned bool `json:"attuned"`
	// Legendary indicates a legendary item.
	Legendary bool `json:"legendary,omitempty"`
}

// HookRegistry manages plugin hooks.
type HookRegistry struct {
	// toolHooks are hooks for specific tools.
	toolHooks map[string]ToolHandler
	// eventHooks are hooks for specific events.
	eventHooks map[EventType][]EventHook
	// chatMessageHooks are hooks for chat messages.
	chatMessageHooks []ChatMessageHook
	// chatParamsHooks are hooks for chat parameters.
	chatParamsHooks []ChatParamsHook
	// chatHeadersHooks are hooks for chat headers.
	chatHeadersHooks []ChatHeadersHook
}

// ToolHandler defines a function that handles tool calls.
type ToolHandlerFunc func(ctx context.Context, input map[string]interface{}) ToolResult

// EventHook defines a function that handles events.
type EventHookFunc func(ctx context.Context, event Event) error

// ChatMessageHook modifies chat messages.
type ChatMessageHook func(ctx context.Context, msg ChatMessage) (ChatMessage, error)

// ChatParamsHook modifies chat parameters before sending.
type ChatParamsHook func(ctx context.Context, params ChatParams) (ChatParams, error)

// ChatHeadersHook modifies chat headers.
type ChatHeadersHook func(ctx context.Context, headers ChatHeaders) (ChatHeaders, error)

// NewHookRegistry creates a new hook registry.
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		toolHooks:       make(map[string]ToolHandler),
		eventHooks:      make(map[EventType][]EventHook),
		chatMessageHooks: make([]ChatMessageHook, 0),
		chatParamsHooks: make([]ChatParamsHook, 0),
		chatHeadersHooks: make([]ChatHeadersHook, 0),
	}
}

// RegisterTool registers a tool handler.
func (r *HookRegistry) RegisterTool(name string, handler ToolHandler) {
	r.toolHooks[name] = handler
}

// RegisterEvent registers an event hook.
func (r *HookRegistry) RegisterEvent(eventType EventType, hook EventHook) {
	r.eventHooks[eventType] = append(r.eventHooks[eventType], hook)
}

// RegisterChatMessage registers a chat message hook.
func (r *HookRegistry) RegisterChatMessage(hook ChatMessageHook) {
	r.chatMessageHooks = append(r.chatMessageHooks, hook)
}

// RegisterChatParams registers a chat params hook.
func (r *HookRegistry) RegisterChatParams(hook ChatParamsHook) {
	r.chatParamsHooks = append(r.chatParamsHooks, hook)
}

// RegisterChatHeaders registers a chat headers hook.
func (r *HookRegistry) RegisterChatHeaders(hook ChatHeadersHook) {
	r.chatHeadersHooks = append(r.chatHeadersHooks, hook)
}

// HandleTool handles a tool call through registered hooks.
func (r *HookRegistry) HandleTool(ctx context.Context, name string, input map[string]interface{}) ToolResult {
	handler, ok := r.toolHooks[name]
	if !ok {
		return ToolResult{
			Success: false,
			Error: &ToolError{
				Code:    "TOOL_NOT_FOUND",
				Message: fmt.Sprintf("tool %q not found", name),
			},
		}
	}
	return handler(ctx, input)
}

// HandleEvent handles an event through registered hooks.
func (r *HookRegistry) HandleEvent(ctx context.Context, event Event) error {
	hooks, ok := r.eventHooks[event.Type]
	if !ok {
		return nil
	}
	for _, hook := range hooks {
		if err := hook(ctx, event); err != nil {
			return fmt.Errorf("event hook error for %s: %w", event.Type, err)
		}
	}
	return nil
}

// HandleChatMessage processes a chat message through hooks.
func (r *HookRegistry) HandleChatMessage(ctx context.Context, msg ChatMessage) (ChatMessage, error) {
	result := msg
	var err error
	for _, hook := range r.chatMessageHooks {
		result, err = hook(ctx, result)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

// HandleChatParams processes chat parameters through hooks.
func (r *HookRegistry) HandleChatParams(ctx context.Context, params ChatParams) (ChatParams, error) {
	result := params
	var err error
	for _, hook := range r.chatParamsHooks {
		result, err = hook(ctx, result)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

// HandleChatHeaders processes chat headers through hooks.
func (r *HookRegistry) HandleChatHeaders(ctx context.Context, headers ChatHeaders) (ChatHeaders, error) {
	result := headers
	var err error
	for _, hook := range r.chatHeadersHooks {
		result, err = hook(ctx, result)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}
