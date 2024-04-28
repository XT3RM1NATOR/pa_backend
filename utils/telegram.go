package utils

import (
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func GetMessageTypeAndFileID(message *tgbotapi.Message) (entity.MessageType, string) {
	if message == nil {
		return "", ""
	}

	switch {
	case message.Text != "":
		return entity.TypeText, ""
	case message.Photo != nil:
		photos := *message.Photo
		largestPhoto := photos[len(photos)-1]
		return entity.TypeImage, largestPhoto.FileID
	case message.Audio != nil:
		return entity.TypeAudio, message.Audio.FileID
	case message.Document != nil:
		return entity.TypeDocument, message.Document.FileID
	case message.Sticker != nil:
		return entity.TypeSticker, message.Sticker.FileID
	case message.Video != nil:
		return entity.TypeVideo, message.Video.FileID
	case message.Voice != nil:
		return entity.TypeVoice, message.Voice.FileID
	case message.VideoNote != nil:
		return entity.TypeVideoNote, message.VideoNote.FileID
	}

	return "", ""
}
