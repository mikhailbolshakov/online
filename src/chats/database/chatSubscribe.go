package database

import (
	"chats/models"
	"github.com/pkg/errors"
	"gitlab.medzdrav.ru/health-service/go-sdk"
	"time"
)

const (
	SubscribeActive   = 1
	SubscribeDeactive = 0

	UserTypeClient   = "client"
	UserTypeDoctor   = "doctor"
	UserTypeOperator = "operator"
	UserTypeBot      = "bot"
)

type SubscribeUserModel struct {
	UserId      uint   `json:"userId"`
	SubscribeId uint   `json:"subscribeId"`
	UserType    string `json:"userType"`
}

type SubscribeChatModel struct {
}

/**
Список подписок на чат: chatId -> SubscribeUserModel{subscribeId, userType}
*/
func (db *Storage) UserSubscribes(userId uint) map[uint]SubscribeUserModel {
	result := make(map[uint]SubscribeUserModel)
	subscribes := []models.ChatSubscribe{}

	fields := "cs.id, " +
		"cs.chat_id, " +
		"cs.user_id, " +
		"cs.user_type"

	query := db.Instance.
		Select(fields).
		Table("chat_subscribes cs").
		Joins("left join chats c on c.id = cs.chat_id").
		Where(
			"cs.user_id = ? and cs.active = ? and c.status = ?",
			userId,
			SubscribeActive,
			ChatStatusOpened,
		)

	rows, err := query.Rows()
	defer rows.Close()
	if err != nil {
		return nil
	}

	for rows.Next() {
		item := &models.ChatSubscribe{}
		query.ScanRows(rows, item)

		result[item.ChatId] = SubscribeUserModel{
			UserId:      item.UserId,
			SubscribeId: item.ID,
			UserType:    item.UserType,
		}
	}

	for _, item := range subscribes {
		result[item.ChatId] = SubscribeUserModel{
			UserId:      item.UserId,
			SubscribeId: item.ID,
			UserType:    item.UserType,
		}
	}

	return result
}

/**
Спиосок подписчиков на чат: userId -> SubscribeUserModel{subscribeId, userType}
*/
func (db *Storage) ChatSubscribes(chatId uint) []SubscribeUserModel {
	result := []SubscribeUserModel{}
	subscribes := []models.ChatSubscribe{}

	db.Instance.
		Where("chat_id = ?", chatId).
		Where("active = ?", SubscribeActive).
		Find(&subscribes)

	for _, item := range subscribes {
		result = append(result, SubscribeUserModel{
			UserId:      item.UserId,
			SubscribeId: item.ID,
			UserType:    item.UserType,
		})
	}

	return result
}

func (db *Storage) GetOpponents(chats []uint, userId uint, sdkConn *sdk.Sdk) (map[uint]models.ExpandedUserModel, error) {
	type Opponent struct {
		ChatId   uint   `json:"chat_id"`
		OrderId  uint   `json:"order_id"`
		UserId   uint   `json:"user_id"`
		UserType string `json:"user_type"`
	}
	result := make(map[uint]models.ExpandedUserModel)

	rows, err := db.Instance.
		Select("cs.chat_id, c.order_id, cs.user_id, cs.user_type").
		Table("chat_subscribes cs").
		Joins("left join chats c on c.id = cs.chat_id").
		Where("cs.chat_id in (?)", chats).
		Where("cs.user_id != ?", userId).
		Where("cs.user_type != ?", UserTypeBot).
		Where("active = ?", SubscribeActive).
		Rows()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		opponent := &Opponent{}
		if err := db.Instance.ScanRows(rows, opponent); err != nil {
			return nil, err
		} else {
			user := &sdk.UserModel{
				Id: opponent.UserId,
			}
			err := sdkConn.VagueUserById(user, opponent.UserType, opponent.OrderId)
			if err != nil {
				return nil, err.Error
			}

			result[opponent.ChatId] = models.ExpandedUserModel{
				UserModel: *user,
				UserType:  opponent.UserType,
			}
		}
	}

	return result, nil
}

func (db *Storage) GetChatOpponents(chatIds []uint, sdkConn *sdk.Sdk) (map[uint]models.ExpandedUserModel, error) {
	type Opponent struct {
		OrderId  uint   `json:"order_id"`
		UserId   uint   `json:"user_id"`
		UserType string `json:"user_type"`
	}
	result := make(map[uint]models.ExpandedUserModel)

	chats := db.Instance.
		Select("chats2.id").
		Table("chats chats").
		Joins("left join chats chats2 on chats2.order_id = chats.order_id").
		Where("chats.id in (?)", chatIds).SubQuery()

	rows, err := db.Instance.
		Select("c.order_id, cs.user_id, cs.user_type").
		Table("chat_subscribes cs").
		Joins("left join chats c on c.id = cs.chat_id").
		Where("cs.chat_id in ?", chats).
		Group("c.order_id, cs.user_id, cs.user_type").
		Rows()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		opponent := &Opponent{}
		if err := db.Instance.ScanRows(rows, opponent); err != nil {
			return nil, err
		} else {
			user := &sdk.UserModel{
				Id: opponent.UserId,
			}
			err := db.Redis.VagueUserById(user, opponent.UserType, opponent.OrderId, sdkConn)
			if err != nil {
				return nil, err.Error
			}

			result[opponent.UserId] = models.ExpandedUserModel{
				UserModel: *user,
				UserType:  opponent.UserType,
			}
		}
	}

	return result, nil
}

func (db *Storage) GetUserType(chatId, userId uint) string {
	subscribe := &models.ChatSubscribe{}
	db.Instance.
		Where("chat_id = ?", chatId).
		Where("user_id = ?", userId).
		First(subscribe)

	return subscribe.UserType
}

func (db *Storage) SubscribeUser(chatId, userId uint, userType string) error {
	chat := &models.Chat{}
	db.Instance.First(chat, chatId)
	if chat.ID == 0 {
		return errors.New(MysqlChatNotExists)
	}

	subscribe := &models.ChatSubscribe{}
	db.Instance.
		Where("chat_id = ?", chatId).
		Where("user_id = ?", userId).
		First(subscribe)

	if subscribe.ID == 0 {
		subscribeModel := &models.ChatSubscribe{
			ChatId:   chatId,
			UserId:   userId,
			UserType: userType,
			Active:   SubscribeActive,
		}

		db.Instance.Create(subscribeModel)
		if db.Instance.Error != nil {
			return db.Instance.Error
		}
	}

	//	deactivate clients opponents before subscribe
	/*if userType != UserTypeClient {
		db.Instance.
			Model(&models.ChatSubscribe{}).
			Where("chat_id = ?", chatId).
			Where("active = ?", SubscribeActive).
			Where("user_id != ?", userId).
			Where("user_type != ?", UserTypeClient).
			UpdateColumn("active", SubscribeDeactive)
	}*/

	return nil
}

func (db *Storage) UnsubscribeUser(chatId, userId uint) error {
	subscribeModel := &models.ChatSubscribe{}

	db.Instance.
		Model(subscribeModel).
		Where("chat_id = ?", chatId).
		Where("user_id = ?", userId).
		Update("active", SubscribeDeactive)
	if db.Instance.Error != nil {
		return db.Instance.Error
	}

	return nil
}

func (db *Storage) RecdUsers(createdAt time.Time) []models.ChatSubscribe {
	subscribers := []models.ChatSubscribe{}

	db.Instance.
		Table("chat_subscribes cs").
		Joins("left join chat_message_statuses cms on cms.subscribe_id = cs.id").
		Joins("left join chat_subscribes cs2 on cs2.chat_id = cs.chat_id").
		Where("cms.created_at >= ?", createdAt).
		Where("cms.status = ?", MessageStatusRecd).
		Where("cs.user_type = ?", UserTypeClient).
		Where("cs2.user_type != ?", UserTypeClient).
		//Pluck("distinct(cs.user_id)", &users)
		Select("distinct(cs.user_id), cs2.user_type, cs.chat_id").Scan(&subscribers)

	return subscribers
}

func (db *Storage) SubscribeUserChange(params *sdk.ChatUserSubscribeChangeRequest) error {
	err := db.Check(params.Body.ChatId, params.Body.OldUserId)
	if err != nil {
		return err
	}

	subscribeModel := &models.ChatSubscribe{}
	db.Instance.Model(subscribeModel).
		Where("chat_id = ?", params.Body.ChatId).
		Where("user_id = ?", params.Body.OldUserId).
		Update("user_id", params.Body.NewUserId)

	if db.Instance.Error != nil {
		return db.Instance.Error
	}

	return nil
}

func (db *Storage) Check(chatId, userId uint) error {
	subscribe := &models.ChatSubscribe{}

	db.Instance.
		Where("chat_id = ?", chatId).
		Where("user_id = ?", userId).
		Find(subscribe)
	if subscribe.ID == 0 {
		return errors.New(MysqlChatAccessDenied)
	}

	return nil
}

func (db *Storage) LastOpponentId(userId uint) uint {
	subscribe := &models.ChatSubscribe{}

	db.Instance.Raw("select cs2.user_id "+
		"from chat_subscribes cs1 "+
		"left join chats c on c.id = cs1.chat_id "+
		"left join chat_subscribes cs2 on cs2.chat_id = cs1.chat_id "+
		"where c.status = 'opened' and cs1.user_type = 'client' and cs1.user_id = ? and cs1.active = 1 "+
		"and cs2.user_type in ('doctor', 'operator') and cs2.active = 1 "+
		"order by c.id desc limit 1", userId).Scan(subscribe)

	return subscribe.UserId
}
