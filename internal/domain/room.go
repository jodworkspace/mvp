package domain

type Room struct {
	RoomID   string `json:"room_id" db:"room_id"`
	RoomName string `json:"room_name" db:"room_name"`
	OwnerID  string `json:"owner_id" db:"owner_id"`
}
