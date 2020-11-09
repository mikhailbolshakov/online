package database

import (
	"chats/models"
	"fmt"
	"github.com/jinzhu/gorm"
	"gitlab.medzdrav.ru/health-service/go-sdk"
	"strconv"
	"time"
)

const (
	ChatStatusOpened  = "opened"
	ChatStatusClosed  = "closed"
	CountChatsDefault = 20
)

type chatListDataItem struct {
	Id             uint      `json:"id"`
	Status         string    `json:"status"`
	OrderId        uint      `json:"order_id"`
	UserId         uint      `json:"user_id"`
	UserType       string    `json:"user_type"`
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

func (db *Storage) ChatCreate(orderId uint) (uint, error) {
	chatModel := &models.Chat{OrderId: orderId}
	db.Instance.Exec("set transaction isolation level serializable") //	TODO
	tx := db.Instance.Begin()
	err := tx.Create(chatModel).Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return chatModel.ID, nil
}

func (db *Storage) ChatChangeStatus(chatId uint, status string) error {
	chatModel := &models.Chat{}
	chatModel.ID = chatId
	db.Instance.Model(chatModel).Update("status", status)

	return db.Instance.Error
}

func (db *Storage) ChatDeactivateNotice(chatId uint) error {
	chatMessageModel := &models.ChatMessage{}

	db.Instance.Model(chatMessageModel).
		Where("type = 'notice'").
		Where("chat_id = ?", chatId).
		Update("deleted_at", gorm.Expr("now()"))

	return db.Instance.Error
}

func (db *Storage) GetUserChats(userId uint, limit uint16, sdkConn *sdk.Sdk) ([]sdk.ChatListResponseDataItem, error) {
	if limit == 0 {
		limit = CountChatsDefault
	}

	chats := []uint{}
	result := []sdk.ChatListResponseDataItem{}
	lastUpdateDate := db.Instance.
		Select("created_at").
		Table("chat_messages cm").
		Where("cm.chat_id = c.id").
		Where("cm.deleted_at IS null").
		Order("id desc").
		Limit("1").SubQuery()

	fields := "c.id, " +
		"c.status, " +
		"c.order_id, " +
		"c.created_at as insert_date, " +
		"IFNULL(?, c.created_at) as last_update_date"

	query := db.Instance.
		Select(fields, lastUpdateDate).
		Table("chat_subscribes cs").
		Joins("inner join chats c on c.id = cs.chat_id").
		Where("cs.user_id = ?", userId).
		Limit(limit).
		Order("cs.chat_id desc")

	rows, err := query.Rows()
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		chatListDataItem := chatListDataItem{}
		if err := db.Instance.ScanRows(rows, &chatListDataItem); err != nil {
			return nil, err
		} else {
			ChatListResponseDataItem := &sdk.ChatListResponseDataItem{
				ChatListDataItem: sdk.ChatListDataItem{
					Id:             chatListDataItem.Id,
					Status:         chatListDataItem.Status,
					OrderId:        chatListDataItem.OrderId,
					InsertDate:     chatListDataItem.InsertDate.In(db.Loc).Format(time.RFC3339),
					LastUpdateDate: chatListDataItem.LastUpdateDate.In(db.Loc).Format(time.RFC3339),
				},
			}
			result = append(result, *ChatListResponseDataItem)
			chats = append(chats, chatListDataItem.Id)
		}
	}

	if len(chats) > 0 {
		opponents, err := db.GetOpponents(chats, userId, sdkConn)
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

func (db *Storage) GetChatById(chatId uint, userId uint, sdkConn *sdk.Sdk) (*sdk.ChatListResponseDataItem, error) {
	chatListDataItem := &chatListDataItem{}

	fields := "c.id, " +
		"c.status, " +
		"c.order_id, " +
		"c.created_at as insert_date, " +
		"IFNULL(?, c.created_at) as last_update_date"

	lastUpdateDate := db.Instance.
		Select("created_at").
		Table("chat_messages cm").
		Where("cm.chat_id = c.id").
		Where("cm.deleted_at IS null").
		Order("cm.id desc").
		Limit("1").SubQuery()

	err := db.Instance.
		Select(fields, lastUpdateDate).
		Table("chat_subscribes cs").
		Joins("inner join chats c on c.id = cs.chat_id").
		Where("cs.chat_id = ?", chatId).
		Where("cs.user_id = ?", userId).
		Where("cs.active = ?", SubscribeActive).
		Find(chatListDataItem).
		Error
	if err != nil {
		return nil, err
	}

	result := &sdk.ChatListResponseDataItem{
		ChatListDataItem: sdk.ChatListDataItem{
			Id:             chatListDataItem.Id,
			Status:         chatListDataItem.Status,
			OrderId:        chatListDataItem.OrderId,
			InsertDate:     chatListDataItem.InsertDate.In(db.Loc).Format(time.RFC3339),
			LastUpdateDate: chatListDataItem.LastUpdateDate.In(db.Loc).Format(time.RFC3339),
		},
	}

	chats := []uint{chatId}
	opponents, err := db.GetOpponents(chats, userId, sdkConn)
	if err != nil {
		return nil, err
	}

	if len(opponents) > 0 {
		result.Opponent = opponents[chatId]
	}

	return result, nil
}

func (db *Storage) GetChatsById(chatsId []uint, userId uint, sdkConn *sdk.Sdk) ([]sdk.ChatListResponseDataItem, error) {
	result := []sdk.ChatListResponseDataItem{}

	items := &[]chatListDataItem{}

	fields := "c.id, " +
		"c.status, " +
		"c.order_id, " +
		"c.created_at as insert_date, " +
		"IFNULL(?, c.created_at) as last_update_date"

	lastUpdateDate := db.Instance.
		Select("created_at").
		Table("chat_messages cm").
		Where("cm.chat_id = c.id").
		Where("cm.deleted_at IS null").
		Order("cm.id desc").
		Limit("1").SubQuery()

	err := db.Instance.
		Select(fields, lastUpdateDate).
		Table("chat_subscribes cs").
		Joins("inner join chats c on c.id = cs.chat_id").
		Where("cs.chat_id in (?)", chatsId).
		Where("cs.user_id = ?", userId).
		Where("cs.active = ?", SubscribeActive).
		Find(items).
		Error
	if err != nil {
		return nil, err
	}

	for _, chatListDataItem := range *items {
		item := &sdk.ChatListResponseDataItem{
			ChatListDataItem: sdk.ChatListDataItem{
				Id:             chatListDataItem.Id,
				Status:         chatListDataItem.Status,
				OrderId:        chatListDataItem.OrderId,
				InsertDate:     chatListDataItem.InsertDate.In(db.Loc).Format(time.RFC3339),
				LastUpdateDate: chatListDataItem.LastUpdateDate.In(db.Loc).Format(time.RFC3339),
			},
		}

		result = append(result, *item)
	}

	opponents, err := db.GetOpponents(chatsId, userId, sdkConn)
	if err != nil {
		return nil, err
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

func (db *Storage) GetChatsByOrder(data []sdk.OrderChatRequestBodyItem, sdkConn *sdk.Sdk) ([]sdk.ChatListResponseDataItem, error) {
	result := []sdk.ChatListResponseDataItem{}
	chatListDataItem, err := db.GetChatsByOrderItems(data)
	if err != nil {
		return nil, err
	}

	for _, chatListDataItem := range chatListDataItem {
		item := &sdk.ChatListResponseDataItem{
			ChatListDataItem: sdk.ChatListDataItem{
				Id:             chatListDataItem.Id,
				Status:         chatListDataItem.Status,
				OrderId:        chatListDataItem.OrderId,
				InsertDate:     chatListDataItem.InsertDate.In(db.Loc).Format(time.RFC3339),
				LastUpdateDate: chatListDataItem.LastUpdateDate.In(db.Loc).Format(time.RFC3339),
			},
		}

		opponents := make(map[uint]models.ExpandedUserModel)
		user := &sdk.UserModel{Id: chatListDataItem.UserId}
		err := sdkConn.VagueUserById(user, chatListDataItem.UserType, chatListDataItem.OrderId)
		if err != nil {
			return nil, err.Error
		}

		opponents[chatListDataItem.UserId] = models.ExpandedUserModel{
			*user,
			chatListDataItem.UserType,
		}

		item.Opponent = opponents

		result = append(result, *item)
	}

	return result, nil
}

func (db *Storage) GetChatsByOrderItems(data []sdk.OrderChatRequestBodyItem) ([]chatListDataItem, error) {
	chatListDataItem := []chatListDataItem{}

	where := ""
	for i, item := range data {
		if i > 0 {
			where += " or "
		}
		where += "(c.order_id = " + strconv.FormatUint(uint64(item.OrderId), 10) +
			" and cs.user_id = " + strconv.FormatUint(uint64(item.OpponentId), 10) + ")"
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
		return nil, err
	}

	return chatListDataItem, nil
}

func (db *Storage) Chat(id uint) *models.Chat {
	chat := &models.Chat{}
	chat.ID = id

	db.Redis.Chat(chat)
	if chat.ID > 0 {
		return chat
	} else {
		db.Instance.First(chat, id)
		db.Redis.SetChat(chat)

		return chat
	}
}
