package repo

import (
	"context"

	"codeberg.org/ApoZero/voxel-server-go/pkg/domain/application"
	mcdomain "codeberg.org/ApoZero/voxel-server-go/pkg/domain/minecraft"
	"codeberg.org/ApoZero/voxel-server-go/pkg/infra/api/mojang"
)

type GameProfileRepository_MojangImpl struct {
	application.GameProfileRepositoryExtension
	mojangAPI *mojang.API
}

func NewGameProfileRepository_Mojang(mojangAPI *mojang.API) *GameProfileRepository_MojangImpl {
	return &GameProfileRepository_MojangImpl{
		mojangAPI: mojangAPI,
	}
}

func (r *GameProfileRepository_MojangImpl) GetGameProfile(ctx context.Context, username string, hash string) (*mcdomain.GameProfile, error) {
	return r.mojangAPI.GetGameProfile(ctx, username, hash)
}
