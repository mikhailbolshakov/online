package database

import (
	"chats/models"
	"chats/sdk"
	"chats/system"
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

func (db *Storage) CreateRoom(roomModel *models.Room) (uuid.UUID, *system.Error) {

	result := db.Instance.Create(roomModel)

	if result.Error != nil {
		return uuid.Nil, &system.Error{Error: result.Error}
	}
	db.Redis.SetRoom(roomModel)

	return roomModel.Id, nil

}

func (db *Storage) RoomSubscribeAccount(roomModel *models.Room, subscriber *models.RoomSubscriber) (uuid.UUID, *system.Error) {

	result := db.Instance.Create(subscriber)
	if result.Error != nil {
		return uuid.Nil, &system.Error{Error: result.Error}
	}
	db.Redis.SetRoom(roomModel)

	return subscriber.Id, nil

}

func (db *Storage) RoomUnsubscribeAccount(roomId, accountId uuid.UUID) *system.Error {

	t := time.Now()

	db.Instance.Model(&models.RoomSubscriber{}).
		Where("account_id = ?::uuid", accountId).
		Where("room_id = ?::uuid", roomId).
		Updates(map[string]interface{}{"unsubscribe_at": t, "updated_at": t})

	err := db.Redis.DeleteRooms([]uuid.UUID{roomId})
	if err != nil {
		return err
	}

	return nil

}

func (db *Storage) GetRoom(id uuid.UUID) (*models.Room, *system.Error) {

	room := &models.Room{}

	room, err := db.Redis.GetRoom(id)
	if err != nil {
		return nil, err
	}

	if room != nil {
		return room, nil
	} else {

		err := db.Instance.
			Preload("Subscribers").
			First(room, id).Error

		if err != nil {
			return nil, system.E(err)
		}

		db.Redis.SetRoom(room)

		return room, nil
	}
}

func (db *Storage) CloseRoom(roomId uuid.UUID) *system.Error {
	roomModel := &models.Room{}
	roomModel.Id = roomId

	db.Instance.First(roomModel)

	t := time.Now()
	roomModel.ClosedAt = &t
	roomModel.UpdatedAt = t

	db.Instance.Model(roomModel).Updates(map[string]interface{}{"closed_at": roomModel.ClosedAt, "updated_at": roomModel.UpdatedAt})
	db.Redis.DeleteRooms([]uuid.UUID{roomModel.Id})

	return &system.Error{Error: db.Instance.Error}
}

func (db *Storage) CloseRoomsByAccounts(accountIds []uuid.UUID) ([]uuid.UUID, *system.Error) {

	var roomIds []uuid.UUID

	if len(accountIds) == 0 {
		return roomIds, nil
	}

	closeTime := time.Now()

	db.Instance.Begin()
	rows, err := db.Instance.Raw(`
			update rooms r set closed_at = ?, updated_at = ?
				where r.closed_at is null and
                      exists(select 1 
                                from room_subscribers rs
								where rs.room_id = r.id and
                                      rs.account_id in (?) and
									  rs.unsubscribe_at is null	
							)
			returning r.id
			`, closeTime, closeTime, accountIds).Rows()

	defer rows.Close()

	if err != nil {
		db.Instance.Rollback()
		return roomIds, &system.Error{
			Error: err,
			// TODO: const
			Message: "[CloseRoomsByAccount]. Error when closing the rooms",
		}
	} else {

		db.Instance.Commit()

		for rows.Next() {
			var roomId string
			err := db.Instance.ScanRows(rows, &roomId)
			if err != nil {
				return roomIds, &system.Error{
					Error:   err,
					Message: "Query error (scan rows)",
				}
			}

			if roomId != "" {
				roomIds = append(roomIds, uuid.FromStringOrNil(roomId))
			}

			// clear cache
			e := db.Redis.DeleteRooms(roomIds)
			if e != nil {
				return roomIds, e
			}

		}

	}

	return roomIds, nil

}

func (db *Storage) GetRoomSubscribers(roomId uuid.UUID) ([]models.RoomSubscriber, *system.Error) {

	room, err := db.Redis.GetRoom(roomId)
	if err != nil {
		return nil, err
	}

	if room != nil {
		return room.Subscribers, nil
	}

	subscribes := []models.RoomSubscriber{}

	db.Instance.
		Where("room_id = ?::uuid", roomId).
		Where("unsubscribe_at is null").
		Find(&subscribes)

	return subscribes, nil
}

func (db *Storage) GetRooms(criteria *models.GetRoomCriteria) ([]models.Room, *system.Error) {

	q := db.Instance.Table("rooms r")

	if criteria.RoomId != uuid.Nil {
		q = q.Where("r.id = ?::uuid", criteria.RoomId)
	}

	if criteria.ReferenceId != "" {
		q = q.Where("r.reference_id = ?", criteria.ReferenceId)
	}

	if !criteria.WithClosed {
		q = q.Where("r.closed_at is null")
	}

	if criteria.AccountId != uuid.Nil {
		q = q.Where(`exists(select 1 
										from room_subscribers rs 
										where r.id = rs.room_id and
										      rs.account_id = ?::uuid and
											  rs.unsubscribe_at is null)`, criteria.AccountId)
	}

	if criteria.ExternalAccountId != "" {
		q = q.Where(`exists(select 1 
										from room_subscribers rs
											join accounts a on rs.account_id = a.id
										where r.id = rs.room_id and
										      a.external_id = ? and 
											  rs.unsubscribe_at is null)`, criteria.ExternalAccountId)
	}

	rows, err := q.Rows()
	if err != nil {
		return nil, &system.Error{
			Error:   err,
			Message: "Query error",
		}
	}

	var result []models.Room
	for rows.Next() {
		room := &models.Room{}
		err := db.Instance.ScanRows(rows, room)
		if err != nil {
			return nil, &system.Error{
				Error:   err,
				Message: "Query error (scan rows)",
			}
		}

		if criteria.WithSubscribers {
			s, e := db.GetRoomSubscribers(room.Id)
			if e != nil {
				return nil, e
			}
			room.Subscribers = s
		}

		result = append(result, *room)
	}

	return result, nil

}

func (db *Storage) GetAccountSubscribers(accountId uuid.UUID) map[uuid.UUID]models.AccountSubscriber {

	result := make(map[uuid.UUID]models.AccountSubscriber)

	rows, err := db.Instance.Raw(`
		select r.id as room_id,
			   rs.account_id,
			   rs.id as subscriber_id,
			   rs.role	
			from room_subscribes rs
				join rooms r on r.id = rs.room_id
			where rs.account_id = ?::uuid and 
				  rs.unsubscribe_at is null and 
				  r.closed_at is null
	`, accountId).
		Rows()

	defer rows.Close()

	if err != nil {
		return result
	}

	for rows.Next() {
		item := &models.AccountSubscriber{}
		db.Instance.ScanRows(rows, item)
		result[item.RoomId] = *item
	}

	return result
}

func (db *Storage) GetRoomAccountSubscribers(roomId uuid.UUID) []models.AccountSubscriber {

	var result []models.AccountSubscriber

	rows, err := db.Instance.Raw(`
		select r.id as room_id,
			   rs.account_id,
			   rs.id as subscriber_id,
			   rs.role	
			from room_subscribes rs
				join rooms r on r.id = rs.room_id
			where r.id = ?::uuid and 
				  rs.unsubscribe_at is null and 
				  r.closed_at is null
	`, roomId).
		Rows()

	defer rows.Close()

	if err != nil {
		return result
	}

	for rows.Next() {
		item := &models.AccountSubscriber{}
		db.Instance.ScanRows(rows, item)
		result = append(result, *item)
	}

	return result
}

func (db *Storage) GetAccountChats(accountId uuid.UUID, limit int) ([]sdk.ChatListResponseDataItem, *system.Error) {

	if limit == 0 {
		limit = CountChatsDefault
	}

	var chats []uuid.UUID
	var result []sdk.ChatListResponseDataItem

	var sql = `
		select
			r.id,
			r.closed_at,
			r.reference_id,
			r.created_at as insert_date,
			coalesce((select max(cm.created_at)
								from chat_messages cm
								where cm.chat_id = r.id
								and cm.deleted_at is null
			), r.created_at) as last_update_date
		from
			room_subscribes rs
		inner join rooms r on
			r.id = rs.room_id
		where
			rs.account_id = ?::uuid and
			r.chat = 1 
		order by rs.created_at desc
		limit ?`

	rows, err := db.Instance.
		Raw(sql, accountId, limit).
		Rows()

	defer rows.Close()
	if err != nil {
		return nil, &system.Error{Error: err}
	}

	for rows.Next() {
		chatListDataItem := chatListDataItem{}
		if err := db.Instance.ScanRows(rows, &chatListDataItem); err != nil {
			return nil, &system.Error{Error: err}
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
