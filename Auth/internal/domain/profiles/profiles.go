package profiles

import (
	"auth/internal/domain/models"

	umv1 "github.com/chas3air/protos/gen/go/usersManager"
	"github.com/google/uuid"
)

func UsrToProtoUsr(user models.User) *umv1.User {
	return &umv1.User{
		Id:       user.Id.String(),
		Login:    user.Login,
		Password: user.Password,
		Role:     user.Role,
	}
}

func ProtoUsrToUsr(proto_usr *umv1.User) (models.User, error) {
	parsedUUID, err := uuid.Parse(proto_usr.GetId())
	if err != nil {
		return models.User{}, err
	}

	return models.User{
		Id:       parsedUUID,
		Login:    proto_usr.GetLogin(),
		Password: proto_usr.GetPassword(),
		Role:     proto_usr.GetRole(),
	}, nil
}
