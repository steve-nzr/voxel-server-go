package domain

import (
	"context"

	mcdomain "codeberg.org/ApoZero/voxel-server-go/pkg/domain/minecraft"
)

type GameProfileRepository interface {
	// https://sessionserver.mojang.com/session/minecraft/hasJoined?username=<player_id>&serverId=<hash>
	GetGameProfile(ctx context.Context, username string, hash string) (*mcdomain.GameProfile, error)
}
