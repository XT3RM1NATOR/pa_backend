package utils

import (
	"github.com/Point-AI/backend/internal/messenger/domain/entity"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gotd/td/tg"
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

func GetMessageTypeAndFileIDFromTelegramAccount(message *tg.Message) (entity.MessageType, int64) {
	if message == nil {
		return "", -1
	}

	switch {
	case message.Message != "":
		return entity.TypeText, -1

	case message.Media != nil:
		switch media := message.Media.(type) {
		case *tg.MessageMediaPhoto:
			if photo, ok := media.Photo.(*tg.Photo); ok && len(photo.Sizes) > 0 {
				return entity.TypeImage, photo.ID
			}

		case *tg.MessageMediaDocument:
			if doc, ok := media.Document.(*tg.Document); ok {
				switch doc.MimeType {
				case "audio/mpeg":
					return entity.TypeAudio, doc.ID
				case "application/pdf":
					return entity.TypeDocument, doc.ID
				case "image/gif":
					return entity.TypeGif, doc.ID
				case "video/mp4":
					return entity.TypeVideo, doc.ID
				case "audio/ogg":
					return entity.TypeVoice, doc.ID
				}
			}
		}
	}

	return "", -1
}

//func DownloadFileByID(client *telegram.Client, fileID int64, accessHash int64, fileRef []byte) ([]byte, error) {
//	ctx := context.Background()
//
//	inputFileLocation := &tg.InputDocumentFileLocation{
//		ID:            fileID,
//		AccessHash:    accessHash,
//		FileReference: fileRef,
//	}
//
//	result, err := client.API().File
//	if err != nil {
//		return nil, err
//	}
//
//	return result.Bytes, nil
//}
