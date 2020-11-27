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
		if err != redis.Nil {
			return nil, app.E().SetError(system.SysErr(err, system.RedisSetErrorCode, val))
		}
	}

	room := &Room{}
	if val != nil {
		err = json.Unmarshal(val, room)
		if err != nil {
			room.Id = uuid.Nil
			return nil, app.E().SetError(system.SysErr(err, system.UnmarshallingErrorCode, val))
		}
		app.L().Debugf("Room found in redis: %s", key)
	}

	return room, nil
}

func (r *Repository) redisSetRoom(room *Room) *system.Error {
	uid := room.Id.String()
	key := "room:" + uid

	marshal, _ := json.Marshal(room)

	err := r.Redis.Instance.Set(key, marshal, r.Redis.Ttl).Err()
	if err != nil {
		return app.E().SetError(system.SysErr(err, system.RedisSetErrorCode, marshal))
	}
	app.L().Debugf("Room set in redis: %s", key)

	return nil
}

func (r *Repository) redisDeleteRooms(roomIds []uuid.UUID) *system.Error {
	var keys []string
	for _, r := range roomIds {
		keys = append(keys,  "room:" + uuid.UUID.String(r))
	}
	r.Redis.Instance.Del(keys...)
	app.L().Debugf("Room delete in redis: %v", keys)
	return nil
}
