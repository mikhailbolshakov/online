package database

import (
	"chats/models"
	"chats/system"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"math"
	"time"
)

const (
	MessageStatusRecd = "recd"
	MessageStatusRead = "read"
)

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

	t := time.Now()
	roomModel.ClosedAt = &t
	roomModel.UpdatedAt = t

	db.Instance.Model(roomModel).
		Where("id = ?::uuid", roomId).
		Updates(map[string]interface{}{"closed_at": roomModel.ClosedAt, "updated_at": roomModel.UpdatedAt})
	db.Redis.DeleteRooms([]uuid.UUID{roomModel.Id})

	return nil
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
			from room_subscribers rs
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
			from room_subscribers rs
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

func (db *Storage) GetMessageHistory(criteria *models.GetMessageHistoryCriteria, pagingRequest *models.PagingRequest) ([]models.MessageHistoryItem,
																														*models.PagingResponse,
																														[]models.MessageAccount,
																														*system.Error) {

	type item struct {
		Id               uuid.UUID `gorm:"column:id"`
		ClientMessageId  string    `gorm:"column:client_message_id"`
		ReferenceId      string    `gorm:"column:reference_id"`
		RoomId           uuid.UUID `gorm:"column:room_id"`
		Type             string    `gorm:"column:type"`
		Message          string    `gorm:"column:message"`
		FileId           string    `gorm:"column:file_id"`
		Params           string    `gorm:"column:params"`
		SenderAccountId  uuid.UUID `gorm:"column:account_id"`
		SenderExternalId string    `gorm:"column:external_id"`
	}

	// here we map incoming sort fields with real fields in the query
	var sortMap = map[string]string{
		"createdAt": "cm.created_at",
	}

	var result []models.MessageHistoryItem

	selectClause := `
			cm.id,
		  	cm.client_message_id,
		  	r.reference_id,
		  	r.id as room_id,
		  	cm."type",
		  	cm.message,
		  	cm.account_id,
			cm.file_id,
			cm.params,
		  	sender_acc.external_id
			`

	query := db.Instance.
		Table(`
			chat_messages cm 
	  			inner join rooms r on cm.room_id = r.id
	  			inner join accounts sender_acc on cm.account_id = sender_acc.id`).
		Where(`
			r.chat = 1  and
		  	r.deleted_at is null and
		  	cm.deleted_at is null`)

	if criteria.RoomId != uuid.Nil {
		query = query.Where("r.id = ?::uuid", criteria.RoomId)
	}

	if criteria.ReferenceId != "" {
		query = query.Where("r.reference_id = ?", criteria.ReferenceId)
	}

	if criteria.AccountId != uuid.Nil {
		query = query.Where(`exists(select 1 
											from room_subscribers rs
											where rs.room_id = r.id and
											rs.account_id = ?::uuid)`, criteria.AccountId)
	}

	if criteria.AccountExternalId != "" {
		query = query.Where(`exists(select 1 
												from room_subscribers rs
													inner join accounts acc_s on rs.account_id = acc_s.id 
												where rs.room_id = r.id and
												acc_s.external_id = ?)`, criteria.AccountExternalId)
	}

	if criteria.CreatedAfter != nil {
		query = query.Where(`cm.created_at >= ?`, criteria.CreatedAfter)
	}

	if criteria.CreatedBefore != nil {
		query = query.Where(`cm.created_at <= ?`, criteria.CreatedBefore)
	}

	if criteria.SentOnly {
		query = query.Where("cm.account_id = ?::uuid", criteria.AccountId)
	}

	if criteria.ReceivedOnly {
		query = query.Where("cm.account_id <> ?::uuid", criteria.AccountId)
	}

	for _, s := range pagingRequest.SortBy {
		query = query.Order(fmt.Sprintf("%s %s", sortMap[s.Field], s.Direction))
	}

	// paging
	var totalCount int64
	var offset int

	query.Count(&totalCount)

	if totalCount > int64(pagingRequest.Size) {
		offset = (pagingRequest.Index - 1) * pagingRequest.Size
	}

	pagingResponse := &models.PagingResponse{
		Total: int(math.Ceil(float64(totalCount) / float64(pagingRequest.Size))),
		Index: pagingRequest.Index,
	}

	query = query.Select(selectClause).Offset(offset).Limit(pagingRequest.Size)

	rows, err := query.Rows()
	if err != nil {
		return nil, nil, nil, system.E(err)
	}
	defer rows.Close()

	var roomIds []uuid.UUID
	for rows.Next() {
		item := &item{}
		_ = db.Instance.ScanRows(rows, item)

		jsonParams := make(map[string]string)
		if item.Params != "" {
			err := json.Unmarshal([]byte(item.Params), &jsonParams)
			if err != nil {
				return nil, nil, nil, system.E(err)
			}
		}

		result = append(result, models.MessageHistoryItem{
			Id:               item.Id,
			ClientMessageId:  item.ClientMessageId,
			ReferenceId:      item.ReferenceId,
			RoomId:           item.RoomId,
			Type:             item.Type,
			Message:          item.Message,
			FileId:           item.FileId,
			Params:           jsonParams,
			SenderAccountId:  item.SenderAccountId,
			SenderExternalId: item.SenderExternalId,
			Statuses:         []models.MessageStatus{},
		})
		roomIds = append(roomIds, item.RoomId)
	}

	// populate statuses
	if criteria.WithStatuses {

		type statusRes struct {
			MessageId  uuid.UUID `gorm:"column:message_id"`
			AccountId  uuid.UUID `gorm:"column:account_id"`
			Status     string    `gorm:"column:status"`
			StatusDate time.Time `gorm:"column:status_date"`
		}

		statuses, err := db.Instance.Raw(`
				select cms.message_id,
					   cms.account_id,
					   cms.status,
					   cms.created_at status_date
					from chat_message_statuses cms
					inner join chat_messages cm on cms.message_id = cm.id 
				where cm.room_id in (?)`, roomIds).
			Rows()
		if err != nil {
			return nil, nil, nil, system.E(err)
		}
		defer statuses.Close()

		for statuses.Next() {
			item := &statusRes{}
			_ = db.Instance.ScanRows(statuses, item)
			for i, _ := range result {
				r := &result[i]
				if uuid.Equal(r.Id, item.MessageId) {
					r.Statuses = append(r.Statuses, models.MessageStatus{
						AccountId:  item.AccountId,
						Status:     item.Status,
						StatusDate: item.StatusDate,
					})
				}
			}
		}

	}

	// populate accounts
	var accountsResult []models.MessageAccount
	if criteria.WithAccounts {
		accounts, e := db.getMessageAccounts(roomIds)
		if e != nil {
			return nil, nil, nil, e
		}
		accountsResult = accounts
	}

	return result, pagingResponse, accountsResult, nil
}

func (db *Storage) getMessageAccounts(roomIds []uuid.UUID) ([]models.MessageAccount, *system.Error) {

	var result []models.MessageAccount

	err := db.Instance.Raw(`
			 select a.* 
			   from room_subscribers rs 
			   inner join accounts a on rs.account_id = a.id 
			 where rs.system_account <> 1 and 
				   rs.deleted_at is null and
					rs.room_id in (?)
		`, roomIds).Scan(&result).Error
	if err != nil {
		return nil, system.E(err)
	}

	return result, nil
}

func (db *Storage) SetReadStatus(messageId uuid.UUID, accountId uuid.UUID) *system.Error {

	// set status for all subscribers with the session's account
	err := db.Instance.
		Model(&models.ChatMessageStatus{}).
		Where("message_id = ?::uuid", messageId).
		Where("account_id = ?::uuid", accountId).
		Updates(&models.ChatMessageStatus{
			Status: MessageStatusRead,
			BaseModel: models.BaseModel{
				UpdatedAt: time.Now(),
			},
		}).Error
	if err != nil {
		return system.E(err)
	}

	return nil
}

func (db *Storage) CreateMessage(messageModel *models.ChatMessage, opponents []models.ChatOpponent) *system.Error {

	if len(messageModel.ClientMessageId) > 0 {
		checkMessage := &models.ChatMessage{}
		db.Instance.
			Where("room_id = ?::uuid", messageModel.RoomId).
			Where("client_message_id = ?", messageModel.ClientMessageId).
			First(checkMessage)

		if checkMessage.Id != uuid.Nil {
			*messageModel = *checkMessage
			return nil
		}
	}

	tx := db.Instance.Begin()
	err := tx.Create(messageModel).Error
	if err != nil {
		return &system.Error{Error: err}
	}

	for _, o := range opponents {

		status := &models.ChatMessageStatus{
			Id:          system.Uuid(),
			AccountId:   o.AccountId,
			MessageId:   messageModel.Id,
			SubscribeId: o.SubscriberId,
			Status:      MessageStatusRecd,
		}

		err := tx.Create(status).Error
		if err != nil {
			tx.Rollback()
			return system.E(err)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return &system.Error{Error: err}
	}

	return nil
}
