package room

import (
	"chats/app"
	"chats/system"
	"encoding/json"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
)

func (r *Repository) redisGetRoom(roomId uuid.UUID) (*Room, *system.Error) {
	uid := roomId.String()
	key := "room:" + uid
	val, err := r.Redis.Instance.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		} else {
			return nil, &system.Error{
				Error:   err,
				Message: system.GetError(system.RedisSetErrorCode),
				Code:    system.RedisSetErrorCode,
				Data:    val,
			}
		}
	}

	room := &Room{}
	err = json.Unmarshal(val, room)
	if err != nil {
		room.Id = uuid.Nil
		return nil, &system.Error{
			Error:   err,
			Message: system.GetError(system.RedisSetErrorCode),
			Code:    system.RedisSetErrorCode,
			Data:    val,
		}
	}

	return room, nil

}

func (r *Repository) redisSetRoom(room *Room) *system.Error {
	uid := room.Id.String()
	key := "room:" + uid

	marshal, _ := json.Marshal(room)

	setErr := r.Redis.Instance.Set(key, marshal, r.Redis.Ttl).Err()
	if setErr != nil {
		redisError := &system.Error{
			Error:   setErr,
			Message: system.GetError(system.RedisSetErrorCode),
			Code:    system.RedisSetErrorCode,
			Data:    marshal,
		}
		app.E().SetError(redisError)

		return redisError
	}

	return nil
}

func (r *Repository) redisDeleteRooms(roomIds []uuid.UUID) *system.Error {
	var keys []string
	for _, r := range roomIds {
		keys = append(keys,  "room:" + uuid.UUID.String(r))
	}
	r.Redis.Instance.Del(keys...)
	return nil
}
