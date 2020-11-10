package sdk

const (
	serviceName = "go-sdk"

	//	user
	userTopic         = "users.1.0"
	userIdPath        = "/user/%d"
	userPushTokenPath = "/user/%d/push-token"
	userTokenPath     = "/user/token"

	//	communication
	communicationTopic    = "communications.1.0"
	communicationPushPath = "/push"

	//	storage
	storageTopic    = "storage.1.0"
	storageFilePath = "/file"

	//	telemed
	telemedTopic              = "telemed.1.0"
	consultationWebsocketPath = "/consultation/websocket"
	consultationPath          = "/consultation/%d"

	//	doctors
	doctorsTopic              = "doctors.1.0"
	doctorsSpecializationPath = "/doctors/%d/specialization"
	doctorIdPath              = "/doctors/%d/ws"

	//	personal
	personalTopic = "personal.1.0"
	//	patient
	personalUserIdPath = "/user/%d"
	patientIdPath      = "/patients/%d/ws"
)
