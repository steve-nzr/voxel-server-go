package minecraft

// GameProfile is returned by Mojang's Session Server.
// Link : https://sessionserver.mojang.com/session/minecraft/hasJoined?username=<player_id>&serverId=<hash>
type GameProfile struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Properties []GameProfile_Property `json:"properties"`
}

type GameProfile_Property struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Signature string `json:"signature,omitempty"`
}

// StatusResponse is what is expected by the Minecraft client.
// Source: https://minecraft.wiki/w/Java_Edition_protocol/Server_List_Ping#Status_Response
type StatusResponse struct {
	Version     StatusResponse_Version     `json:"version"`
	Players     StatusResponse_Players     `json:"players"`
	Description StatusResponse_Description `json:"description"`
}

type StatusResponse_Version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type StatusResponse_Players struct {
	Max    int                     `json:"max"`
	Online int                     `json:"online"`
	Sample []StatusResponse_Player `json:"sample,omitempty"`
}

type StatusResponse_Player struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type StatusResponse_Description struct {
	Text string `json:"text"`
}
