package libs

import (
	"context"
	"fmt"
	"zumygo/helpers"
	"zumygo/config"
	"regexp"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	waTypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// Compile regex pattern once for better performance
var nonDigitRegex = regexp.MustCompile(`\D+`)

func SerializeMessage(mess *events.Message, conn *IClient) *IMessage {
	if mess == nil {
		return nil
	}
	
	var media whatsmeow.DownloadableMessage
	var text string
	var args []string
	var owner []string
	var isOwner = false
	var isMedia string
	var sender waTypes.JID

	if mess.Info.AddressingMode == "lid" {
		sender = mess.Info.SenderAlt
	} else {
		sender = mess.Info.Sender
	}

	mess.Message = helpers.ParseMessage(mess)
	body := helpers.GetTextMessage(mess)
	
	// Validate body before processing
	if body == "" {
		body = ""
	}
	
	// Safe command extraction with multi-prefix support
	parts := strings.Split(body, " ")
	var command string
	var hasPrefix bool
	var prefix string
	
	if len(parts) > 0 {
		command = strings.ToLower(parts[0])
		// Check if command has a valid prefix
		prefix, hasPrefix = ExtractPrefix(command)
		if hasPrefix {
			// Remove prefix from command
			command = strings.TrimSpace(strings.TrimPrefix(command, prefix))
		} else {
			// If no prefix found, don't treat as command
			command = ""
		}
	}
	
	if config.Config != nil {
		owner = config.Config.Owner
	}

	for _, v := range owner {
		if v != "" && strings.Contains(nonDigitRegex.ReplaceAllString(v, ""), sender.ToNonAD().User) {
			isOwner = true
			break
		}
	}

	// Safe mention removal
	if conn != nil && conn.WA != nil && conn.WA.Store != nil && conn.WA.Store.ID != nil {
		botID := "@" + conn.WA.Store.ID.ToNonAD().User
		if strings.HasPrefix(body, botID) {
			body = strings.Trim(strings.Replace(body, botID, "", 1), " ")
		}
	}

	if hasPrefix && HasCommand(command) {
		if len(parts) > 1 {
			text = strings.Join(parts[1:], ` `)
		}
		args = helpers.ArrayFilter(strings.Split(text, " "), "")
	} else {
		text = body
		args = helpers.ArrayFilter(strings.Split(body, " "), "")
	}
	
	// Command field will be set in the return statement

	quotedMsg := helpers.ParseQuotedMessage(mess.Message)

	if quotedMsg != nil {
		media = helpers.GetMediaMessage(quotedMsg)
		isMedia = helpers.GetMediaType(quotedMsg)
	} else if mess.Message != nil {
		media = helpers.GetMediaMessage(mess.Message)
		isMedia = helpers.GetMediaType(mess.Message)
	} else {
		media = nil
	}

	// Safe context info extraction
	var expiration uint32
	var quoted *waE2E.ContextInfo
	
	if mess.Message != nil {
		contextInfo := helpers.GetContextInfo(mess.Message)
		if contextInfo != nil {
			expiration = contextInfo.GetExpiration()
			quoted = contextInfo
		}
	}

	return &IMessage{
		Info:       mess.Info,
		Sender:     sender,
		IsOwner:    isOwner,
		Body:       body,
		Text:       text,
		Args:       args,
		Command:    command,
		Message:    mess.Message,
		IsMedia:    isMedia,
		Media:      media,
		Expiration: expiration,
		Quoted:     quoted,
		Client:     conn,
		Reply: func(text string, opts ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
			if conn == nil || conn.WA == nil {
				fmt.Printf("ERROR: Client is not initialized for Reply\n")
				return whatsmeow.SendResponse{}, fmt.Errorf("client is not initialized")
			}
			
			var Expiration uint32
			if mess.Message != nil {
				contextInfo := helpers.GetContextInfo(mess.Message)
				if contextInfo != nil {
					Expiration = contextInfo.GetExpiration()
				}
			}
			
			response, err := conn.SendText(mess.Info.Chat, text, &waE2E.ContextInfo{
				StanzaID:      &mess.Info.ID,
				Participant:   proto.String(mess.Info.Sender.String()),
				QuotedMessage: mess.Message,
				Expiration:    &Expiration,
			}, opts...)
			
			if err != nil {
				fmt.Printf("ERROR: Failed to send reply: %v\n", err)
			}
			
			return response, err
		},
		React: func(emoji string, opts ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
			if conn == nil || conn.WA == nil {
				return whatsmeow.SendResponse{}, fmt.Errorf("client is not initialized")
			}
			
			return conn.WA.SendMessage(context.Background(), mess.Info.Chat, conn.WA.BuildReaction(mess.Info.Chat, mess.Info.Sender, mess.Info.ID, emoji), opts...)
		},
	}
}