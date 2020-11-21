package database

import (
	//"chats/converter"
	"chats/models"
	"chats/sdk"
	"chats/system"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	UserTypeUser = "user"
	UserTypeAnonymousUser = "anonymous_user"
	UserTypeBot  = "bot"
)


func (db *Storage) GetOpponents(chats []uuid.UUID, accountId uuid.UUID) (map[uuid.UUID]models.ExpandedAccountModel, *system.Error) {
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
		Where("unsubscribe_at is null").
		Rows()

	if err != nil {
		return nil, &system.Error{
			Error: err,
		}
	}

	for rows.Next() {
		opponent := &Opponent{}
		if err := db.Instance.ScanRows(rows, opponent); err != nil {
			return nil, &system.Error{Error: err}
		} else {

			account, err := db.GetAccount(opponent.AccountId, "")
			if err != nil {
				return nil, err
			}

			result[opponent.ChatId] = models.ExpandedAccountModel{
				Account: sdk.Account{
					Id:         account.Id,
					Account:    account.Account,
					Type:       account.Type,
					ExternalId: account.ExternalId,
					FirstName:  account.FirstName,
					MiddleName: account.MiddleName,
					LastName:   account.LastName,
					Email:      account.Email,
					Phone:      account.Phone,
					AvatarUrl:  account.AvatarUrl,
				},
				Role:    opponent.Role,
			}
		}
	}

	return result, nil
}

func (db *Storage) GetChatOpponents(chatIds []uuid.UUID, sdkConn *sdk.Sdk) (map[uuid.UUID]models.ExpandedAccountModel, *system.Error) {
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
		return nil, &system.Error{Error: err}
	}

	for rows.Next() {
		opponent := &Opponent{}
		if err := db.Instance.ScanRows(rows, opponent); err != nil {
			return nil, &system.Error{Error: err}
		} else {

			account, err := db.GetAccount(opponent.AccountId, "")
			if err != nil {
				return nil, err
			}

			result[opponent.AccountId] = models.ExpandedAccountModel{
				Account: sdk.Account{
					Id:         account.Id,
					Account:    account.Account,
					Type:       account.Type,
					ExternalId: account.ExternalId,
					FirstName:  account.FirstName,
					MiddleName: account.MiddleName,
					LastName:   account.LastName,
					Email:      account.Email,
					Phone:      account.Phone,
					AvatarUrl:  account.AvatarUrl,
				},
				Role:    opponent.Role,
			}
		}
	}

	return result, nil
}

func (db *Storage) GetAccountRole(chatId, accountId uuid.UUID) string {
	subscribe := &models.RoomSubscriber{}
	db.Instance.
		Where("chat_id = ?", chatId).
		Where("user_id = ?", accountId).
		First(subscribe)

	return subscribe.Role
}

func (db *Storage) RecdUsers(createdAt time.Time) []models.RoomSubscriber {
	subscribers := []models.RoomSubscriber{}

	db.Instance.
		Table("chat_subscribes cs").
		Joins("left join chat_message_statuses cms on cms.subscribe_id = cs.id").
		Joins("left join chat_subscribes cs2 on cs2.chat_id = cs.chat_id").
		Where("cms.created_at >= ?", createdAt).
		Where("cms.status = ?", MessageStatusRecd).
		Where("cs.user_type = ?", UserTypeUser).
		Where("cs2.user_type != ?", UserTypeUser).
		//Pluck("distinct(cs.user_id)", &users)
		Select("distinct(cs.user_id), cs2.user_type, cs.chat_id").Scan(&subscribers)

	return subscribers
}

func (db *Storage) SubscribeUserChange(params *sdk.ChatAccountSubscribeChangeRequest) error {
	err := db.Check(params.Body.ChatId, params.Body.OldAccountId)
	if err != nil {
		return err
	}

	subscribeModel := &models.RoomSubscriber{}
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
	subscribe := &models.RoomSubscriber{}

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
	subscribe := &models.RoomSubscriber{}

	db.Instance.Raw("select cs2.user_id "+
		"from chat_subscribes cs1 "+
		"left join chats c on c.id = cs1.chat_id "+
		"left join chat_subscribes cs2 on cs2.chat_id = cs1.chat_id "+
		"where c.status = 'opened' and cs1.user_type = 'client' and cs1.user_id = ? and cs1.active = 1 "+
		"and cs2.user_type in ('doctor', 'operator') and cs2.active = 1 "+
		"order by c.id desc limit 1", userId).Scan(subscribe)

	return uuid.Nil //subscribe.AccountId
}
