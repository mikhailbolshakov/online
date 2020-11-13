package database

import (
	"chats/converter"
	"chats/models"
	"chats/sdk"
	"chats/sentry"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
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

type SubscribeAccountModel struct {
	AccountId   uuid.UUID `json:"accountId"`
	SubscribeId uuid.UUID `json:"subscribeId"`
	Role        string    `json:"role"`
}

type SubscribeChatModel struct {
}

/**
Список подписок на чат: chatId -> SubscribeAccountModel{subscribeId, userType}
*/
func (db *Storage) AccountSubscribes(accountId uuid.UUID) map[uuid.UUID]SubscribeAccountModel {
	result := make(map[uuid.UUID]SubscribeAccountModel)
	subscribes := []models.ChatSubscribe{}

	fields := "cs.id, " +
		"cs.chat_id, " +
		"cs.account_id"

	query := db.Instance.
		Select(fields).
		Table("chat_subscribes cs").
		Joins("left join chats c on c.id = cs.chat_id").
		Where(
			"cs.account_id = ? and cs.active = ? and c.status = ?",
			accountId,
			SubscribeActive,
			ChatStatusOpened,
		)

	rows, err := query.Rows()
	defer rows.Close()
	if err != nil {
		return result
	}

	for rows.Next() {
		item := &models.ChatSubscribe{}
		query.ScanRows(rows, item)

		result[item.ChatId] = SubscribeAccountModel{

			AccountId:   item.AccountId,
			SubscribeId: item.Id,
			Role:        item.Role,
		}
	}

	for _, item := range subscribes {
		result[item.ChatId] = SubscribeAccountModel{
			AccountId:   item.AccountId,
			SubscribeId: item.Id,
			Role:        item.Role,
		}
	}

	return result
}

/**
Спиосок подписчиков на чат: userId -> SubscribeAccountModel{subscribeId, userType}
*/
func (db *Storage) ChatSubscribes(chatId uuid.UUID) []SubscribeAccountModel {
	result := []SubscribeAccountModel{}
	subscribes := []models.ChatSubscribe{}

	db.Instance.
		Where("chat_id = ?", chatId).
		Where("active = ?", SubscribeActive).
		Find(&subscribes)

	for _, item := range subscribes {
		result = append(result, SubscribeAccountModel{
			AccountId:   item.AccountId,
			SubscribeId: item.Id,
			Role:        item.Role,
		})
	}

	return result
}

func (db *Storage) GetOpponents(chats []uuid.UUID, accountId uuid.UUID) (map[uuid.UUID]models.ExpandedAccountModel, *sentry.SystemError) {
	type Opponent struct {
		ChatId      uuid.UUID `json:"chat_id"`
		ReferenceId string    `json:"reference_id"`
		AccountId   uuid.UUID `json:"account_id"`
		Role        string    `json:"role"`
	}
	result := make(map[uuid.UUID]models.ExpandedAccountModel)

	rows, err := db.Instance.
		Select("cs.chat_id, c.reference_id, cs.account_id, cs.role").
		Table("chat_subscribes cs").
		Joins("left join chats c on c.id = cs.chat_id").
		Where("cs.chat_id in (?)", chats).
		Where("cs.account_id != ?", accountId).
		Where("cs.role != ?", UserTypeBot).
		Where("active = ?", SubscribeActive).
		Rows()

	if err != nil {
		return nil, &sentry.SystemError{
			Error: err,
		}
	}

	for rows.Next() {
		opponent := &Opponent{}
		if err := db.Instance.ScanRows(rows, opponent); err != nil {
			return nil, &sentry.SystemError{Error: err}
		} else {

			accountModel, err := db.GetAccount(opponent.AccountId, "")
			if err != nil {
				return nil, err
			}

			result[opponent.ChatId] = models.ExpandedAccountModel{
				Account: *converter.ConvertAccountFromModel(accountModel),
				Role:    opponent.Role,
			}
		}
	}

	return result, nil
}

func (db *Storage) GetChatOpponents(chatIds []uuid.UUID, sdkConn *sdk.Sdk) (map[uuid.UUID]models.ExpandedAccountModel, *sentry.SystemError) {
	type Opponent struct {
		ReferenceId uuid.UUID `json:"reference_id"`
		AccountId   uuid.UUID `json:"account_id"`
		Role        string    `json:"role"`
	}
	result := make(map[uuid.UUID]models.ExpandedAccountModel)

	chats := db.Instance.
		Select("chats2.id").
		Table("chats chats").
		Joins("left join chats chats2 on chats2.order_id = chats.order_id").
		Where("chats.id in (?)", chatIds)

	rows, err := db.Instance.
		Select("c.order_id, cs.user_id, cs.user_type").
		Table("chat_subscribes cs").
		Joins("left join chats c on c.id = cs.chat_id").
		Where("cs.chat_id in ?", chats).
		Group("c.order_id, cs.user_id, cs.user_type").
		Rows()

	if err != nil {
		return nil, &sentry.SystemError{Error: err}
	}

	for rows.Next() {
		opponent := &Opponent{}
		if err := db.Instance.ScanRows(rows, opponent); err != nil {
			return nil, &sentry.SystemError{Error: err}
		} else {

			accountModel, err := db.GetAccount(opponent.AccountId, "")
			if err != nil {
				return nil, err
			}

			result[opponent.AccountId] = models.ExpandedAccountModel{
				Account: *converter.ConvertAccountFromModel(accountModel),
				Role:    opponent.Role,
			}
		}
	}

	return result, nil
}

func (db *Storage) GetAccountRole(chatId, accountId uuid.UUID) string {
	subscribe := &models.ChatSubscribe{}
	db.Instance.
		Where("chat_id = ?", chatId).
		Where("user_id = ?", accountId).
		First(subscribe)

	return subscribe.Role
}

func (db *Storage) SubscribeAccount(subscribeModel *models.ChatSubscribe) (uuid.UUID, error) {

	// we don't have to do additional check to avoid round-trip
	// we'd better to have unique constraint to check uniqueness of (user_id, chat_id)
	// after all, it supposed to be quite rare case
	/*
		subscribe := &models.ChatSubscribe{}
		db.Instance.
			Where("chat_id = ?", chatId).
			Where("user_id = ?", userId).
			First(subscribe)

		if subscribe.ID == 0 {
			subscribeModel := &models.ChatSubscribe{
				ChatId:   chatId,
				AccountId:   userId,
				Role: userType,
				Active:   SubscribeActive,
			}

			db.Instance.Create(subscribeModel)
			if db.Instance.Error != nil {
				return db.Instance.Error
			}
		}
	*/

	db.Instance.Create(subscribeModel)
	if db.Instance.Error != nil {
		return uuid.Nil, db.Instance.Error
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

	return subscribeModel.Id, nil
}

func (db *Storage) UnsubscribeUser(chatId, accountId uuid.UUID) error {
	subscribeModel := &models.ChatSubscribe{}

	db.Instance.
		Model(subscribeModel).
		Where("chat_id = ?", chatId).
		Where("account_id = ?", accountId).
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

func (db *Storage) SubscribeUserChange(params *sdk.ChatAccountSubscribeChangeRequest) error {
	err := db.Check(params.Body.ChatId, params.Body.OldAccountId)
	if err != nil {
		return err
	}

	subscribeModel := &models.ChatSubscribe{}
	db.Instance.Model(subscribeModel).
		Where("chat_id = ?", params.Body.ChatId).
		Where("user_id = ?", params.Body.OldAccountId).
		Update("user_id", params.Body.NewAccountId)

	if db.Instance.Error != nil {
		return db.Instance.Error
	}

	return nil
}

func (db *Storage) Check(chatId, accountId uuid.UUID) error {
	subscribe := &models.ChatSubscribe{}

	db.Instance.
		Where("chat_id = ?", chatId).
		Where("account_id = ?", accountId).
		Find(subscribe)
	if subscribe.Id == uuid.Nil {
		return errors.New(MysqlChatAccessDenied)
	}

	return nil
}

func (db *Storage) LastOpponentId(userId uuid.UUID) uuid.UUID {
	subscribe := &models.ChatSubscribe{}

	db.Instance.Raw("select cs2.user_id "+
		"from chat_subscribes cs1 "+
		"left join chats c on c.id = cs1.chat_id "+
		"left join chat_subscribes cs2 on cs2.chat_id = cs1.chat_id "+
		"where c.status = 'opened' and cs1.user_type = 'client' and cs1.user_id = ? and cs1.active = 1 "+
		"and cs2.user_type in ('doctor', 'operator') and cs2.active = 1 "+
		"order by c.id desc limit 1", userId).Scan(subscribe)

	return uuid.Nil //subscribe.AccountId
}
