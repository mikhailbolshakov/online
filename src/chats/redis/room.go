package redis

import (
	"chats/models"
	"chats/system"
	"encoding/json"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
)

func (r *Redis) GetRoom(roomId uuid.UUID) (*models.Room, *system.Error) {
	uid := roomId.String()
	key := "room:" + uid
	val, err := r.Instance.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		} else {
			return nil, &system.Error{
				Error:   err,
				Message: RedisSetError,
				Code:    RedisSetErrorCode,
				Data:    val,
			}
		}
	}

	room := &models.Room{}
	err = json.Unmarshal(val, room)
	if err != nil {
		room.Id = uuid.Nil
		return nil, &system.Error{
			Error:   err,
			Message: RedisSetError,
			Code:    RedisSetErrorCode,
			Data:    val,
		}
	}

	return room, nil

}

func (r *Redis) SetRoom(room *models.Room) *system.Error {
	uid := room.Id.String()
	key := "room:" + uid

	marshal, _ := json.Marshal(room)

	setErr := r.Instance.Set(key, marshal, r.Ttl).Err()
	if setErr != nil {
		redisError := &system.Error{
			Error:   setErr,
			Message: RedisSetError,
			Code:    RedisSetErrorCode,
			Data:    marshal,
		}
		system.ErrHandler.SetError(redisError)

		return redisError
	}

	return nil
}

func (r *Redis) DeleteRooms(roomIds []uuid.UUID) *system.Error {
	var keys []string
	for _, r := range roomIds {
		keys = append(keys,  "room:" + uuid.UUID.String(r))
	}
	r.Instance.Del(keys...)
	return nil
}
