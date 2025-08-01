package libs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func SerializeClient(conn *whatsmeow.Client) *IClient {
	return &IClient{
		WA: conn,
	}
}

func (conn *IClient) SendText(from types.JID, txt string, opts *waE2E.ContextInfo, optn ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
	if conn.WA == nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("client is not initialized")
	}
	
	ok, er := conn.WA.SendMessage(context.Background(), from, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(txt),
			ContextInfo: opts,
		},
	}, optn...)
	if er != nil {
		return whatsmeow.SendResponse{}, er
	}
	return ok, nil
}

func (conn *IClient) SendWithNewsLestter(from types.JID, text string, newjid string, newserver int32, name string, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	if opts == nil {
		opts = &waE2E.ContextInfo{}
	}
	
	ok, er := conn.SendText(from, text, &waE2E.ContextInfo{
		ForwardedNewsletterMessageInfo: &waE2E.ContextInfo_ForwardedNewsletterMessageInfo{
			NewsletterJID:     proto.String(newjid),
			NewsletterName:    proto.String(name),
			ServerMessageID:   proto.Int32(newserver),
			ContentType:       waE2E.ContextInfo_ForwardedNewsletterMessageInfo_UPDATE.Enum(),
			AccessibilityText: proto.String(""),
		},
		IsForwarded:   proto.Bool(true),
		StanzaID:      opts.StanzaID,
		Participant:   opts.Participant,
		QuotedMessage: opts.QuotedMessage,
	})

	if er != nil {
		return whatsmeow.SendResponse{}, er
	}
	return ok, nil
}

func (conn *IClient) SendImage(from types.JID, data []byte, caption string, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	if conn.WA == nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("client is not initialized")
	}
	
	if len(data) == 0 {
		return whatsmeow.SendResponse{}, fmt.Errorf("image data is empty")
	}
	
	uploaded, err := conn.WA.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("failed to upload image: %v", err)
	}
	
	resultImg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Caption:       proto.String(caption),
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			ContextInfo:   opts,
		},
	}
	ok, err := conn.WA.SendMessage(context.Background(), from, resultImg)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}
	return ok, nil
}

func (conn *IClient) SendVideo(from types.JID, data []byte, caption string, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	if conn.WA == nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("client is not initialized")
	}
	
	if len(data) == 0 {
		return whatsmeow.SendResponse{}, fmt.Errorf("video data is empty")
	}
	
	uploaded, err := conn.WA.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err != nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("failed to upload video: %v", err)
	}
	
	resultVideo := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Caption:       proto.String(caption),
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			ContextInfo:   opts,
		},
	}
	ok, er := conn.WA.SendMessage(context.Background(), from, resultVideo)
	if er != nil {
		return whatsmeow.SendResponse{}, er
	}
	return ok, nil
}

func (conn *IClient) SendDocument(from types.JID, data []byte, fileName string, caption string, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	if conn.WA == nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("client is not initialized")
	}
	
	if len(data) == 0 {
		return whatsmeow.SendResponse{}, fmt.Errorf("document data is empty")
	}
	
	if fileName == "" {
		fileName = "document"
	}
	
	uploaded, err := conn.WA.Upload(context.Background(), data, whatsmeow.MediaDocument)
	if err != nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("failed to upload document: %v", err)
	}
	
	resultDoc := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileName:      proto.String(fileName),
			Caption:       proto.String(caption),
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			ContextInfo:   opts,
		},
	}
	ok, er := conn.WA.SendMessage(context.Background(), from, resultDoc)
	if er != nil {
		return whatsmeow.SendResponse{}, er
	}
	return ok, nil
}

func (conn *IClient) SendAudio(from types.JID, data []byte, fileName string, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	if conn.WA == nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("client is not initialized")
	}
	
	if len(data) == 0 {
		return whatsmeow.SendResponse{}, fmt.Errorf("audio data is empty")
	}
	
	uploaded, err := conn.WA.Upload(context.Background(), data, whatsmeow.MediaAudio)
	if err != nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("failed to upload audio: %v", err)
	}
	
	resultAudio := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String("audio/mpeg"),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			ContextInfo:   opts,
		},
	}
	
	ok, err := conn.WA.SendMessage(context.Background(), from, resultAudio)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}
	return ok, nil
}

func (conn *IClient) DeleteMsg(from types.JID, id string, me bool) error {
	if conn.WA == nil {
		return fmt.Errorf("client is not initialized")
	}
	
	if id == "" {
		return fmt.Errorf("message ID is required")
	}
	
	_, err := conn.WA.SendMessage(context.Background(), from, &waE2E.Message{
		ProtocolMessage: &waE2E.ProtocolMessage{
			Type: waE2E.ProtocolMessage_REVOKE.Enum(),
			Key: &waCommon.MessageKey{
				FromMe: proto.Bool(me),
				ID:     proto.String(id),
			},
		},
	})
	return err
}

func (conn *IClient) ParseJID(arg string) (types.JID, bool) {
	if arg == "" {
		return types.JID{}, false
	}
	
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			return recipient, false
		} else if recipient.User == "" {
			return recipient, false
		}
		return recipient, true
	}
}

func (conn *IClient) FetchGroupAdmin(Jid types.JID) ([]string, error) {
	if conn.WA == nil {
		return nil, fmt.Errorf("client is not initialized")
	}
	
	var Admin []string
	resp, err := conn.WA.GetGroupInfo(Jid)
	if err != nil {
		return Admin, err
	} else {
		for _, group := range resp.Participants {
			if group.IsAdmin || group.IsSuperAdmin {
				Admin = append(Admin, group.JID.String())
			}
		}
	}
	return Admin, err
}

func (conn *IClient) SendSticker(jid types.JID, data []byte, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	if conn.WA == nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("client is not initialized")
	}
	
	if len(data) == 0 {
		return whatsmeow.SendResponse{}, fmt.Errorf("sticker data is empty")
	}
	
	uploaded, err := conn.WA.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("failed to upload sticker: %v", err)
	}

	ok, er := conn.WA.SendMessage(context.Background(), jid, &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			ContextInfo:   opts,
		},
	})

	if er != nil {
		return whatsmeow.SendResponse{}, er
	}

	return ok, nil
}

// MediaItem represents a single media item in an album
type MediaItem struct {
	Data     []byte
	Type     string // "image", "video", "document"
	Caption  string
	FileName string // for documents
}

// SendMediaAlbum sends multiple media items as an album
func (conn *IClient) SendMediaAlbum(from types.JID, mediaItems []MediaItem, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	if conn.WA == nil {
		return whatsmeow.SendResponse{}, fmt.Errorf("client is not initialized")
	}
	
	if len(mediaItems) == 0 {
		return whatsmeow.SendResponse{}, fmt.Errorf("no media items provided")
	}
	
	if len(mediaItems) == 1 {
		// If only one item, send as single media
		item := mediaItems[0]
		switch item.Type {
		case "image":
			return conn.SendImage(from, item.Data, item.Caption, opts)
		case "video":
			return conn.SendVideo(from, item.Data, item.Caption, opts)
		case "document":
			return conn.SendDocument(from, item.Data, item.FileName, item.Caption, opts)
		default:
			return whatsmeow.SendResponse{}, fmt.Errorf("unsupported media type: %s", item.Type)
		}
	}
	
	// For multiple items, we need to send them as separate messages but grouped
	// WhatsApp doesn't have native album support like Whiskeysockets, but we can group them
	var responses []whatsmeow.SendResponse
	
	for i, item := range mediaItems {
		var response whatsmeow.SendResponse
		var err error
		
		// Add context info to group messages
		contextInfo := &waE2E.ContextInfo{}
		if opts != nil {
			contextInfo = opts
		}
		
		switch item.Type {
		case "image":
			response, err = conn.SendImage(from, item.Data, item.Caption, contextInfo)
		case "video":
			response, err = conn.SendVideo(from, item.Data, item.Caption, contextInfo)
		case "document":
			response, err = conn.SendDocument(from, item.Data, item.FileName, item.Caption, contextInfo)
		default:
			return whatsmeow.SendResponse{}, fmt.Errorf("unsupported media type: %s", item.Type)
		}
		
		if err != nil {
			return whatsmeow.SendResponse{}, fmt.Errorf("failed to send media item %d: %v", i+1, err)
		}
		
		responses = append(responses, response)
	}
	
	// Return the first response (they should all be successful)
	return responses[0], nil
}

// SendImageAlbum sends multiple images as an album
func (conn *IClient) SendImageAlbum(from types.JID, images [][]byte, captions []string, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	if len(images) != len(captions) {
		return whatsmeow.SendResponse{}, fmt.Errorf("number of images and captions must match")
	}
	
	var mediaItems []MediaItem
	for i, image := range images {
		caption := ""
		if i < len(captions) {
			caption = captions[i]
		}
		
		mediaItems = append(mediaItems, MediaItem{
			Data:    image,
			Type:    "image",
			Caption: caption,
		})
	}
	
	return conn.SendMediaAlbum(from, mediaItems, opts)
}

// SendVideoAlbum sends multiple videos as an album
func (conn *IClient) SendVideoAlbum(from types.JID, videos [][]byte, captions []string, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	if len(videos) != len(captions) {
		return whatsmeow.SendResponse{}, fmt.Errorf("number of videos and captions must match")
	}
	
	var mediaItems []MediaItem
	for i, video := range videos {
		caption := ""
		if i < len(captions) {
			caption = captions[i]
		}
		
		mediaItems = append(mediaItems, MediaItem{
			Data:    video,
			Type:    "video",
			Caption: caption,
		})
	}
	
	return conn.SendMediaAlbum(from, mediaItems, opts)
}

// SendMixedAlbum sends a mixed album with different types of media
func (conn *IClient) SendMixedAlbum(from types.JID, mediaItems []MediaItem, opts *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	return conn.SendMediaAlbum(from, mediaItems, opts)
}

func (conn *IClient) GetBytes(url string) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("URL is required")
	}
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return bytes, nil
}
