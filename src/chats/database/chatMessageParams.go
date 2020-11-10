package database

import (
	"chats/models"
	uuid "github.com/satori/go.uuid"
)

type Params struct {
	MessageId uint   `json:"message_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

func (db *Storage) SetParams(messageId uuid.UUID, params map[string]string) error {
	if len(params) > 0 {
		db.Instance.Exec("set transaction isolation level serializable")
		tx := db.Instance.Begin()
		for key, value := range params {
			paramsModel := &models.ChatMessageParams{
				MessageId: messageId,
				Key:       key,
				Value:     value,
			}
			tx.Create(paramsModel)

			if tx.Error != nil {
				tx.Rollback()
				return tx.Error
			}
		}

		tx.Commit()
	}
	return nil
}

func (db *Storage) GetParams(messages []uuid.UUID) (result []models.ChatMessageParams, err error) {
	result = []models.ChatMessageParams{}

	if len(messages) > 0 {
		err = db.Instance.Where("message_id in (?)", messages).Find(&result).Error
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (db *Storage) GetParamsMap(messageId uuid.UUID) (result map[string]string, err error) {
	result = map[string]string{}
	messages := []uuid.UUID{messageId}

	get, err := db.GetParams(messages)
	if err != nil {
		return nil, err
	}

	for _, item := range get {
		result[item.Key] = item.Value
	}

	return result, nil
}
