package database

import (
	"chats/models"
	uuid "github.com/satori/go.uuid"
)

const (
	MessageStatusRecd = "recd"
	MessageStatusRead = "read"
)

func (db *Storage) NewStatus(messageStatusModel *models.ChatMessageStatus) error {
	messageStatusModel.Status = MessageStatusRecd
	db.Instance.Create(messageStatusModel)
	if db.Instance.Error != nil {
		return db.Instance.Error
	}

	return nil
}

// Возвращает статус сообщения
func (db *Storage) getMessageStatus(messageId uuid.UUID, status string) (*models.ChatMessageStatus, error) {
	statusModel := &models.ChatMessageStatus{}

	where := "message_id = ? and status = ?"

	err := db.Instance.Find(statusModel, where, messageId, status).Error

	return statusModel, err
}

// Меняет статус сообщения на "прочитанно"
func (db *Storage) SetReadStatus(messageId uuid.UUID) error {
	statusModel, err := db.getMessageStatus(messageId, MessageStatusRecd)

	if err == nil {
		err = db.Instance.Model(statusModel).Update("status", MessageStatusRead).Error
	}

	return err
}