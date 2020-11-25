package room

import (
	"chats/app"
	rep "chats/repository"
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

type Repository struct {
	Storage *app.Storage
	Redis   *app.Redis
}

func CreateRepository(storage *app.Storage) *Repository {
	return &Repository{
		Storage: storage,
		Redis:   storage.Redis,
	}
}

func (r *Repository) CreateRoom(roomModel *Room) (uuid.UUID, *system.Error) {

	result := r.Storage.Instance.Create(roomModel)

	if result.Error != nil {
		return uuid.Nil, &system.Error{Error: result.Error}
	}
	r.redisSetRoom(roomModel)

	return roomModel.Id, nil

}

func (r *Repository) RoomSubscribeAccount(roomModel *Room, subscriber *RoomSubscriber) (uuid.UUID, *system.Error) {

	result := r.Storage.Instance.Create(subscriber)
	if result.Error != nil {
		return uuid.Nil, &system.Error{Error: result.Error}
	}
	r.redisSetRoom(roomModel)

	return subscriber.Id, nil

}

func (r *Repository) RoomUnsubscribeAccount(roomId, accountId uuid.UUID) *system.Error {

	t := time.Now()

	r.Storage.Instance.Model(&RoomSubscriber{}).
		Where("account_id = ?::uuid", accountId).
		Where("room_id = ?::uuid", roomId).
		Updates(map[string]interface{}{"unsubscribe_at": t, "updated_at": t})

	err := r.redisDeleteRooms([]uuid.UUID{roomId})
	if err != nil {
		return err
	}

	return nil

}

func (r *Repository) GetRoom(id uuid.UUID) (*Room, *system.Error) {

	room := &Room{}

	room, err := r.redisGetRoom(id)
	if err != nil {
		return nil, err
	}

	if room != nil {
		return room, nil
	} else {

		err := r.Storage.Instance.
			Preload("Subscribers").
			First(room, id).Error

		if err != nil {
			return nil, system.E(err)
		}

		r.redisSetRoom(room)

		return room, nil
	}
}

func (r *Repository) CloseRoom(roomId uuid.UUID) *system.Error {
	roomModel := &Room{}

	t := time.Now()
	roomModel.ClosedAt = &t
	roomModel.UpdatedAt = t

	r.Storage.Instance.Model(roomModel).
		Where("id = ?::uuid", roomId).
		Updates(map[string]interface{}{"closed_at": roomModel.ClosedAt, "updated_at": roomModel.UpdatedAt})
	r.redisDeleteRooms([]uuid.UUID{roomModel.Id})

	return nil
}

func (r *Repository) CloseRoomsByAccounts(accountIds []uuid.UUID) ([]uuid.UUID, *system.Error) {

	var roomIds []uuid.UUID

	if len(accountIds) == 0 {
		return roomIds, nil
	}

	closeTime := time.Now()

	r.Storage.Instance.Begin()
	rows, err := r.Storage.Instance.Raw(`
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
		r.Storage.Instance.Rollback()
		return roomIds, &system.Error{
			Error: err,
			// TODO: const
			Message: "[CloseRoomsByAccount]. Error when closing the rooms",
		}
	} else {

		r.Storage.Instance.Commit()

		for rows.Next() {
			var roomId string
			err := r.Storage.Instance.ScanRows(rows, &roomId)
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
			e := r.redisDeleteRooms(roomIds)
			if e != nil {
				return roomIds, e
			}

		}

	}

	return roomIds, nil

}

func (r *Repository) GetRoomSubscribers(roomId uuid.UUID) ([]RoomSubscriber, *system.Error) {

	room, err := r.redisGetRoom(roomId)
	if err != nil {
		return nil, err
	}

	if room != nil {
		return room.Subscribers, nil
	}

	subscribes := []RoomSubscriber{}

	r.Storage.Instance.
		Where("room_id = ?::uuid", roomId).
		Where("unsubscribe_at is null").
		Find(&subscribes)

	return subscribes, nil
}

func (db *Repository) GetRooms(criteria *GetRoomCriteria) ([]Room, *system.Error) {

	q := db.Storage.Instance.Table("rooms r")

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

	var result []Room
	for rows.Next() {
		room := &Room{}
		err := db.Storage.Instance.ScanRows(rows, room)
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

func (db *Repository) GetAccountSubscribers(accountId uuid.UUID) map[uuid.UUID]AccountSubscriber {

	result := make(map[uuid.UUID]AccountSubscriber)

	rows, err := db.Storage.Instance.Raw(`
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
		item := &AccountSubscriber{}
		db.Storage.Instance.ScanRows(rows, item)
		result[item.RoomId] = *item
	}

	return result
}

func (db *Repository) GetRoomAccountSubscribers(roomId uuid.UUID) []AccountSubscriber {

	var result []AccountSubscriber

	rows, err := db.Storage.Instance.Raw(`
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
		item := &AccountSubscriber{}
		db.Storage.Instance.ScanRows(rows, item)
		result = append(result, *item)
	}

	return result
}

func (db *Repository) GetMessageHistory(criteria *GetMessageHistoryCriteria, pagingRequest *rep.PagingRequest) ([]MessageHistoryItem,
	*rep.PagingResponse,
	[]MessageAccount,
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

	var result []MessageHistoryItem

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

	query := db.Storage.Instance.
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

	pagingResponse := &rep.PagingResponse{
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
		_ = db.Storage.Instance.ScanRows(rows, item)

		jsonParams := make(map[string]string)
		if item.Params != "" {
			err := json.Unmarshal([]byte(item.Params), &jsonParams)
			if err != nil {
				return nil, nil, nil, system.E(err)
			}
		}

		result = append(result, MessageHistoryItem{
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
			Statuses:         []MessageStatus{},
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

		statuses, err := db.Storage.Instance.Raw(`
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
			_ = db.Storage.Instance.ScanRows(statuses, item)
			for i, _ := range result {
				r := &result[i]
				if uuid.Equal(r.Id, item.MessageId) {
					r.Statuses = append(r.Statuses, MessageStatus{
						AccountId:  item.AccountId,
						Status:     item.Status,
						StatusDate: item.StatusDate,
					})
				}
			}
		}

	}

	// populate accounts
	var accountsResult []MessageAccount
	if criteria.WithAccounts {
		accounts, e := db.getMessageAccounts(roomIds)
		if e != nil {
			return nil, nil, nil, e
		}
		accountsResult = accounts
	}

	return result, pagingResponse, accountsResult, nil
}

func (db *Repository) getMessageAccounts(roomIds []uuid.UUID) ([]MessageAccount, *system.Error) {

	var result []MessageAccount

	err := db.Storage.Instance.Raw(`
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

func (db *Repository) SetReadStatus(messageId uuid.UUID, accountId uuid.UUID) *system.Error {

	// set status for all subscribers with the session's account
	err := db.Storage.Instance.
		Model(&ChatMessageStatus{}).
		Where("message_id = ?::uuid", messageId).
		Where("account_id = ?::uuid", accountId).
		Updates(&ChatMessageStatus{
			Status: MessageStatusRead,
			BaseModel: rep.BaseModel{
				UpdatedAt: time.Now(),
			},
		}).Error
	if err != nil {
		return system.E(err)
	}

	return nil
}

func (db *Repository) CreateMessage(messageModel *ChatMessage, opponents []ChatOpponent) *system.Error {

	if len(messageModel.ClientMessageId) > 0 {
		checkMessage := &ChatMessage{}
		db.Storage.Instance.
			Where("room_id = ?::uuid", messageModel.RoomId).
			Where("client_message_id = ?", messageModel.ClientMessageId).
			First(checkMessage)

		if checkMessage.Id != uuid.Nil {
			*messageModel = *checkMessage
			return nil
		}
	}

	tx := db.Storage.Instance.Begin()
	err := tx.Create(messageModel).Error
	if err != nil {
		return &system.Error{Error: err}
	}

	for _, o := range opponents {

		status := &ChatMessageStatus{
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

func (db *Repository) RecdUsers(createdAt time.Time) []RoomSubscriber {
	subscribers := []RoomSubscriber{}

	//db.Storage.Instance.
	//	Table("room_subscribers cs").
	//	Joins("left join chat_message_statuses cms on cms.subscribe_id = cs.id").
	//	Joins("left join room_subscribers cs2 on cs2.room_id = cs.room_id").
	//	Where("cms.created_at >= ?", createdAt).
	//	Where("cms.status = ?", MessageStatusRecd).
	//	Where("cs.user_type = ?", UserTypeUser).
	//	Where("cs2.user_type != ?", UserTypeUser).
	//	//Pluck("distinct(cs.user_id)", &users)
	//	Select("distinct(cs.user_id), cs2.user_type, cs.chat_id").Scan(&subscribers)

	return subscribers
}

func (db *Repository) LastOpponentId(userId uuid.UUID) uuid.UUID {
	subscribe := &RoomSubscriber{}

	db.Storage.Instance.Raw("select cs2.user_id "+
		"from chat_subscribes cs1 "+
		"left join chats c on c.id = cs1.chat_id "+
		"left join chat_subscribes cs2 on cs2.chat_id = cs1.chat_id "+
		"where c.status = 'opened' and cs1.user_type = 'client' and cs1.user_id = ? and cs1.active = 1 "+
		"and cs2.user_type in ('doctor', 'operator') and cs2.active = 1 "+
		"order by c.id desc limit 1", userId).Scan(subscribe)

	return uuid.Nil //subscribe.AccountId
}
