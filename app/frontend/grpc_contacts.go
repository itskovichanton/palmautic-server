package frontend

type ContactGrpcHandler struct {
	UnimplementedContactsServer
	PalmGrpcControllerImpl

	CreateOrUpdateContactAction *CreateOrUpdateContactAction
	DeleteContactAction         *DeleteContactAction
	SearchContactAction         *SearchContactAction
}
