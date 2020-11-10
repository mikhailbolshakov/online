package database

import (
	"chats/models"
	"chats/sdk"
	uuid "github.com/satori/go.uuid"
	"strings"
	"time"
)

const (
	MessageTypeMessage       = "message"
	MessageTypeSystem        = "system"
	MessageTypeDocument      = "document"
	MessageTypeDoctorDetail  = "doctor_detail"
	MessageTypeClinicDetail  = "clinic_detail"
	MessageTypePayDetail     = "pay_detail"
	MessageTypeProductDetail = "product_detail"
	MessageTypeMedcardDetail = "medcard_detail"
	MessageTypeOrderDetail   = "order_detail"
	MessageTypeProductLink   = "product_link"
	MessageTypePromocodeLink = "promocode_link"
	MessageTypeNotice        = "notice"

	CountMessagesDefault = 50

	CountMessagesOld = 10
)

func ValidateType(messageType string) bool {
	var types = [...]string{
		MessageTypeMessage,
		MessageTypeSystem,
		MessageTypeDocument,
		MessageTypeDoctorDetail,
		MessageTypeClinicDetail,
		MessageTypePayDetail,
		MessageTypeProductDetail,
		MessageTypeMedcardDetail,
		MessageTypeOrderDetail,
		MessageTypeProductLink,
		MessageTypePromocodeLink,
		MessageTypeNotice,
	}

	for _, v := range types {
		if v == messageType {
			return true
		}
	}

	return false
}

func (db *Storage) NewMessageTransact(messageModel *models.ChatMessage, params map[string]string, opponentsId []uuid.UUID) error {
	db.Instance.Exec("set transaction isolation level serializable") //	TODO

	if len(messageModel.ClientMessageId) > 0 {
		checkMessage := &models.ChatMessage{}
		db.Instance.
			Where("chat_id = ?", messageModel.ChatId).
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
		return err
	}

	err = db.SetParams(messageModel.Id, params)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, opponentId := range opponentsId {
		status := &models.ChatMessageStatus{
			MessageId:   messageModel.Id,
			SubscribeId: opponentId,
		}

		err := db.NewStatus(status)
		if err != nil {
			tx.Rollback()
			return nil
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}

	return nil
}

func (db *Storage) GetMessagesRecent(params *models.ChatMessageHistory) ([]sdk.ChatMessagesResponseDataItem, error) {
	subscribe := &models.ChatSubscribe{}
	if params.Admin {
		db.Instance.
			Where("chat_id = ?", params.ChatId).
			Find(subscribe)
	} else {
		err := db.Check(params.ChatId, params.AccountId)
		if err != nil {
			return nil, err
		}
	}

	status := db.Instance.
		Select("cms.status as status").
		Table("chat_message_statuses cms").
		Where("cms.message_id = cm.id").
		Where("cms.status = ?", MessageStatusRead).
		Limit(1).
		SubQuery()

	chats := db.Instance.
		Select("chats2.id").
		Table("chats chats").
		Joins("left join chats chats2 on chats2.order_id = chats.order_id").
		Where("chats.id = ?", params.ChatId).SubQuery()

	fields := "cm.id, " +
		"cm.client_message_id, " +
		"cm.chat_id, " +
		"IFNULL(?, ?) as status, " +
		"cm.type, " +
		"cm.message text, " +
		"cm.file_id, " +
		"cm.created_at insert_date, " +
		"cs.user_type sender, " +
		"cs.user_id"

	query := db.Instance.
		Select(fields, status, MessageStatusRecd).
		Table("chat_messages cm").
		Joins("left join chat_subscribes cs on cs.id = cm.subscribe_id").
		Where("cm.chat_id in ?", chats).
		Where("cm.deleted_at IS null").
		Where("cm.id > ?", params.MessageId).
		Order("cm.id asc")

	rows, err := query.Rows()
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var messages []sdk.ChatMessagesResponseDataItem
	messageIds := []uuid.UUID{}
	for rows.Next() {
		var message sdk.ChatMessagesResponseDataItemDb
		query.ScanRows(rows, &message)
		tmp := sdk.ChatMessagesResponseDataItem{
			ChatMessagesResponseDataItemTmp: message.ChatMessagesResponseDataItemTmp,
			InsertDate:                      message.InsertDate.In(db.Loc).Format(time.RFC3339),
		}
		messages = append(messages, tmp)
		messageIds = append(messageIds, message.Id)
	}

	mp, err := db.GetParams(messageIds)
	if err != nil {
		return nil, err
	}

	if len(mp) > 0 {
		for i, message := range messages {
			first := true
			for _, item := range mp {
				if message.Id == item.MessageId {
					if first {
						messages[i].Params = make(map[string]string)
						first = false
					}
					messages[i].Params[item.Key] = item.Value
				}
			}
		}
	}

	return messages, nil
}

func (db *Storage) GetMessagesHistory(params *models.ChatMessageHistory) ([]sdk.ChatMessagesResponseDataItem, error) {
	if params.Count == 0 {
		params.Count = CountMessagesDefault
	}

	if params.Admin {
		subscribe := &models.ChatSubscribe{}
		db.Instance.
			Where("chat_id = ?", params.ChatId).
			Find(subscribe)
	} else {
		err := db.Check(params.ChatId, params.AccountId)
		if err != nil {
			return nil, err
		}
	}

	loadPrevMessages := false

	if params.OnlyOneChat == false && params.MessageId == uuid.Nil && params.Search == "" && params.Date == "" {
		firstMessageChats := db.Instance.
			Select("chats2.id").
			Table("chats chats").
			Joins("left join chats chats2 on chats2.order_id = chats.order_id").
			Where("chats.id = ?", params.ChatId).
			SubQuery()

		firstMessage := &models.FirstMessage{}
		db.Instance.
			Select("cm.id").
			Table("chat_messages cm").
			Where("cm.chat_id in ?", firstMessageChats).
			Order("cm.id asc").
			Limit(1).
			Find(firstMessage)

		params.MessageId = firstMessage.Id
		params.NewMessages = true
		loadPrevMessages = true
	}

	chatsQuery := db.Instance.
		Select("chats2.id").
		Table("chats chats").
		Joins("left join chats chats2 on chats2.order_id = chats.order_id")

	searchChats := db.Instance.
		Select("cs.chat_id").
		Table("chat_subscribes cs")

	if params.UserType == "operator" {
		//пациент из запрашиваемого чата
		client := db.Instance.
			Select("ocs.user_id").
			Table("chat_subscribes ocs").
			Where("ocs.chat_id = ?", params.ChatId).
			Where("ocs.user_type = 'client'").
			SubQuery()

		//все чаты пациента
		userChats := db.Instance.
			Select("ucs.chat_id").
			Table("chat_subscribes ucs").
			Where("ucs.user_id in ?", client).
			SubQuery()

		//все чаты мк и пациента
		searchChats = searchChats.
			Where("cs.user_id = ?", params.AccountId).
			Where("cs.chat_id in ?", userChats)

		chatsQuery = searchChats
	} else if params.UserType == "patient" {
		//все чаты пациента
		searchChats = searchChats.Where("cs.user_id = ?", params.AccountId)
		chatsQuery = chatsQuery.Where("chats.id in ?", searchChats.SubQuery())
	} else {
		chatsQuery = chatsQuery.Where("chats.id = ?", params.ChatId)
	}

	status := db.Instance.
		Select("cms.status as status").
		Table("chat_message_statuses cms").
		Where("cms.message_id = cm.id").
		Where("cms.status = ?", MessageStatusRead).
		Limit(1).
		SubQuery()

	fields := "cm.id, " +
		"cm.client_message_id, " +
		"cm.chat_id, " +
		"IFNULL(?, ?) as status, " +
		"cm.type, " +
		"cm.message text, " +
		"cm.file_id, " +
		"cm.created_at insert_date, " +
		"cs.user_type sender, " +
		"cs.user_id"

	queryPrevMessages := db.Instance
	query := db.Instance

	if params.MessageId != uuid.Nil {

		ordering := "<"
		if params.NewMessages {
			ordering = ">="
		}

		query = db.Instance.
			Select(fields, status, MessageStatusRecd).
			Table("chat_messages cm").
			Joins("left join chat_subscribes cs on cs.id = cm.subscribe_id").
			Where("cm.chat_id in ?", chatsQuery.SubQuery()).
			Where("cm.deleted_at IS null").
			Where("cm.id "+ordering+" ?", params.MessageId).
			Order("cm.id desc")

		if loadPrevMessages {
			orderingPrev := ">="
			if params.NewMessages {
				orderingPrev = "<"
			}
			queryPrevMessages = db.Instance.
				Select(fields, status, MessageStatusRecd).
				Table("chat_messages cm").
				Joins("left join chat_subscribes cs on cs.id = cm.subscribe_id").
				Where("cm.chat_id in ?", chatsQuery.SubQuery()).
				Where("cm.deleted_at IS null").
				Where("cm.id "+orderingPrev+" ?", params.MessageId).
				Order("cm.id desc").
				Limit(CountMessagesOld)
		}
	} else {
		query = db.Instance.
			Select(fields, status, MessageStatusRecd).
			Table("chat_messages cm").
			Joins("left join chat_subscribes cs on cs.id = cm.subscribe_id").
			Where("cm.chat_id in ?", chatsQuery.SubQuery()).
			Where("cm.deleted_at IS null")
	}

	if params.NewMessages != false {
		var totalCount uint16
		var offset uint16
		totalCount = 0
		offset = 0

		query.Count(&totalCount)

		if totalCount > params.Count {
			offset = totalCount - params.Count
		}

		query = query.Offset(offset)
	}

	query = query.Limit(params.Count)
	
	if params.Search != "" {
		params.Search = strings.Replace(params.Search, ",", "", -1)
		searchMessages := strings.Split(params.Search, " ")

		for i := range searchMessages {
			v := strings.TrimSpace(searchMessages[i])

			if v != "" {
				query = query.Where("cm.message LIKE ?", "%%"+v+"%%")
			}
		}
	}

	if params.Date != "" {
		date, err := time.Parse("2006-01-02", params.Date) //todo format in const

		if err != nil {
			return nil, err
		}

		year, month, day := date.Date()
		begin := time.Date(year, month, day, 0, 0, 0, 0, date.Location())
		end := time.Date(year, month, day, 23, 59, 59, 1e9-1, date.Location())

		query = query.
			Where("cm.created_at BETWEEN ? AND ?", begin, end)

		if params.Search == "" {
			query = query.Limit(1)
		}
	}

	if params.Search == "" && params.Date == "" {
		query = query.Order("cm.id desc")
	} else {
		query = query.Order("cm.id")
	}

	rows, err := query.Rows()
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var messages []sdk.ChatMessagesResponseDataItem
	var messageIds []uuid.UUID
	for rows.Next() {
		var message sdk.ChatMessagesResponseDataItemDb
		query.ScanRows(rows, &message)
		tmp := sdk.ChatMessagesResponseDataItem{
			ChatMessagesResponseDataItemTmp: message.ChatMessagesResponseDataItemTmp,
			InsertDate:                      message.InsertDate.In(db.Loc).Format(time.RFC3339),
		}
		messages = append([]sdk.ChatMessagesResponseDataItem{tmp}, messages...)
		messageIds = append(messageIds, message.Id)
	}

	if loadPrevMessages {
		rows, err := queryPrevMessages.Rows()
		defer rows.Close()
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var message sdk.ChatMessagesResponseDataItemDb
			queryPrevMessages.ScanRows(rows, &message)
			tmp := sdk.ChatMessagesResponseDataItem{
				ChatMessagesResponseDataItemTmp: message.ChatMessagesResponseDataItemTmp,
				InsertDate:                      message.InsertDate.In(db.Loc).Format(time.RFC3339),
			}
			messages = append([]sdk.ChatMessagesResponseDataItem{tmp}, messages...)
			messageIds = append(messageIds, message.Id)
		}
	}

	mp, err := db.GetParams(messageIds)
	if err != nil {
		return nil, err
	}

	if len(mp) > 0 {
		for i, message := range messages {
			first := true
			for _, item := range mp {
				if message.Id == item.MessageId {
					if first {
						messages[i].Params = make(map[string]string)
						first = false
					}
					messages[i].Params[item.Key] = item.Value
				}
			}
		}
	}

	return messages, nil
}
