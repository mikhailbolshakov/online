package database

import (
	"chats/converter"
	"chats/models"
	"chats/sdk"
	"chats/sentry"
	"fmt"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	ChatStatusOpened  = "opened"
	ChatStatusClosed  = "closed"
	CountChatsDefault = 20
)

type chatListDataItem struct {
	Id             uuid.UUID `json:"id"`
	Status         string    `json:"status"`
	ReferenceId    string    `json:"reference_id"`
	AccountId      uuid.UUID `json:"account_id"`
	Role           string    `json:"role"`
	InsertDate     time.Time `json:"insert_date"`
	LastUpdateDate time.Time `json:"last_update_date"`
}

//	TODO
func CatchPanic() (err error) {
	if r := recover(); r != nil {
		_, ok := r.(error)
		if !ok {
			err = fmt.Errorf("this %v", r)
			fmt.Println(err)
		}
	}
	return err
}

func (db *Storage) ChatCreate(chatModel *models.Chat) (uuid.UUID, *sentry.SystemError) {

	result := db.Instance.Create(chatModel)

	if result.Error != nil {
		return uuid.Nil, &sentry.SystemError{Error: result.Error}
	}

	db.Redis.SetChat(chatModel)

	return chatModel.Id, nil

}

func (db *Storage) ChatChangeStatus(chatId uuid.UUID, status string) *sentry.SystemError {
	chatModel := &models.Chat{}
	chatModel.Id = chatId
	db.Instance.Model(chatModel).Update("status", status)

	return &sentry.SystemError{Error: db.Instance.Error}
}

func (db *Storage) ChatDeactivateNotice(chatId uuid.UUID) *sentry.SystemError {
	chatMessageModel := &models.ChatMessage{}

	db.Instance.Model(chatMessageModel).
		Where("type = 'notice'").
		Where("chat_id = ?", chatId).
		Update("deleted_at", gorm.Expr("now()"))

	return &sentry.SystemError{Error: db.Instance.Error}
}

func (db *Storage) GetAccountChats(accountId uuid.UUID, limit int) ([]sdk.ChatListResponseDataItem, *sentry.SystemError) {

	if limit == 0 {
		limit = CountChatsDefault
	}

	var chats []uuid.UUID
	var result []sdk.ChatListResponseDataItem

	var sql = `
		select
			c.id,
			c.status,
			c.reference_id,
			c.created_at as insert_date,
			coalesce((select max(cm.created_at)
								from chat_messages cm
								where cm.chat_id = c.id
								and cm.deleted_at is null
			), c.created_at) as last_update_date
		from
			chat_subscribes cs
		inner join chats c on
			c.id = cs.chat_id
		where
			cs.account_id = ?::uuid
		order by cs.created_at desc
		limit ?`

	rows, err := db.Instance.
		Raw(sql, accountId, limit).
		Rows()

	defer rows.Close()
	if err != nil {
		return nil, &sentry.SystemError{Error: err}
	}

	for rows.Next() {
		chatListDataItem := chatListDataItem{}
		if err := db.Instance.ScanRows(rows, &chatListDataItem); err != nil {
			return nil, &sentry.SystemError{Error: err}
		} else {
			ChatListResponseDataItem := &sdk.ChatListResponseDataItem{
				ChatListDataItem: sdk.ChatListDataItem{
					Id:             chatListDataItem.Id,
					Status:         chatListDataItem.Status,
					ReferenceId:    chatListDataItem.ReferenceId,
					InsertDate:     chatListDataItem.InsertDate.In(db.Loc).Format(time.RFC3339),
					LastUpdateDate: chatListDataItem.LastUpdateDate.In(db.Loc).Format(time.RFC3339),
				},
			}
			result = append(result, *ChatListResponseDataItem)
			chats = append(chats, chatListDataItem.Id)
		}
	}

	if len(chats) > 0 {
		opponents, err := db.GetOpponents(chats, accountId)
		if err != nil {
			return nil, err
		}

		for i, item := range result {
			for chatId, opponent := range opponents {
				if item.Id == chatId {
					result[i].Opponent = opponent
				}
			}
		}
	}

	return result, nil
}

func (db *Storage) GetChatById(chatId uuid.UUID, accountId uuid.UUID) (*sdk.ChatListResponseDataItem, *sentry.SystemError) {

	chatsId := []uuid.UUID {chatId}
	chats, err := db.GetChatsById(chatsId, accountId)
	if err != nil {
		return nil, err
	}

	if len(chats) >= 1 {
		item := chats[0]
		return &item, nil
	} else {
		return &sdk.ChatListResponseDataItem{}, nil
	}

}

func (db *Storage) GetChatsById(chatsId []uuid.UUID, accountId uuid.UUID) ([]sdk.ChatListResponseDataItem, *sentry.SystemError) {
	result := []sdk.ChatListResponseDataItem{}

	items := &[]chatListDataItem{}

	var sql = `
		select
			c.id,
			c.status,
			c.reference_id,
			c.created_at as insert_date,
			coalesce((select max(cm.created_at)
								from chat_messages cm
								where cm.chat_id = c.id
								and cm.deleted_at is null
			), c.created_at) as last_update_date
		from
			chat_subscribes cs
		inner join chats c on
			c.id = cs.chat_id
		where
			cs.chat_id in (?)
			and cs.account_id = ?::uuid
			and cs.active = ?`

	err := db.Instance.
		Raw(sql, chatsId, accountId, SubscribeActive).
		Find(items).
		Error

	if err != nil {
		return nil, &sentry.SystemError{Error: err}
	}

	for _, chatListDataItem := range *items {
		item := &sdk.ChatListResponseDataItem{
			ChatListDataItem: sdk.ChatListDataItem{
				Id:             chatListDataItem.Id,
				Status:         chatListDataItem.Status,
				ReferenceId:    chatListDataItem.ReferenceId,
				InsertDate:     chatListDataItem.InsertDate.In(db.Loc).Format(time.RFC3339),
				LastUpdateDate: chatListDataItem.LastUpdateDate.In(db.Loc).Format(time.RFC3339),
			},
		}

		result = append(result, *item)
	}

	opponents, sentryErr := db.GetOpponents(chatsId, accountId)
	if sentryErr != nil {
		return nil, sentryErr
	}

	for i, item := range result {
		for chatId, opponent := range opponents {
			if item.Id == chatId && len(opponents) > 0 {
				result[i].Opponent = opponent
			}
		}
	}

	return result, nil
}

func (db *Storage) GetChatsByReference(data []sdk.ReferenceChatRequestBodyItem) ([]sdk.ChatListResponseDataItem, *sentry.SystemError) {
	result := []sdk.ChatListResponseDataItem{}
	chatListDataItem, err := db.GetChatsByReferenceItems(data)
	if err != nil {
		return nil, err
	}

	for _, chatListDataItem := range chatListDataItem {
		item := &sdk.ChatListResponseDataItem{
			ChatListDataItem: sdk.ChatListDataItem{
				Id:             chatListDataItem.Id,
				Status:         chatListDataItem.Status,
				ReferenceId:    chatListDataItem.ReferenceId,
				InsertDate:     chatListDataItem.InsertDate.In(db.Loc).Format(time.RFC3339),
				LastUpdateDate: chatListDataItem.LastUpdateDate.In(db.Loc).Format(time.RFC3339),
			},
		}

		opponents := make(map[uuid.UUID]models.ExpandedAccountModel)

		account, err := db.GetAccount(chatListDataItem.AccountId, "")
		if err != nil {
			return nil, err
		}

		opponents[chatListDataItem.AccountId] = models.ExpandedAccountModel{
			*converter.ConvertAccountFromModel(account),
			chatListDataItem.Role,
		}

		item.Opponent = opponents

		result = append(result, *item)
	}

	return result, nil
}

func (db *Storage) GetChatsByReferenceItems(data []sdk.ReferenceChatRequestBodyItem) ([]chatListDataItem, *sentry.SystemError) {
	chatListDataItem := []chatListDataItem{}

	where := ""
	for i, item := range data {
		if i > 0 {
			where += " or "
		}

		where += "(c.reference_id = " + item.ReferenceId +
			" and cs.account_id = " + uuid.UUID.String(item.OpponentId) + ")"
	}

	query := "select distinct " +
		"c.id, c.order_id,  c.status, c.order_id, c.created_at as insert_date, cs.user_id, cs.user_type, " +
		"ifnull((select created_at from chat_messages where chat_id = c.id and deleted_at IS null order by id desc limit 1), c.created_at) " +
		"as last_update_date " +
		"from chat_subscribes cs " +
		"inner join chats c on c.id = cs.chat_id " +
		"where " + where + " " +
		"order by c.id desc"

	err := db.Instance.
		Raw(query).
		Scan(&chatListDataItem).
		Error
	if err != nil {
		return nil, &sentry.SystemError{Error: err}
	}

	return chatListDataItem, nil
}

func (db *Storage) Chat(id uuid.UUID) *models.Chat {
	chat := &models.Chat{}
	chat.Id = id

	db.Redis.Chat(chat)
	if chat.Id != uuid.Nil {
		return chat
	} else {
		db.Instance.First(chat, id)
		db.Redis.SetChat(chat)
		return chat
	}
}
