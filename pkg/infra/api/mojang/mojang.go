package mojang

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	appdomain "codeberg.org/ApoZero/voxel-server-go/pkg/domain/application"
	mcdomain "codeberg.org/ApoZero/voxel-server-go/pkg/domain/minecraft"
)

const (
	MojangHasJoinedEndpoint      = "hasJoined"
	MojangAPITLSHandshakeTimeout = time.Second * 5
	MojangAPITimeout             = time.Second * 10
)

var (
	MojangAPIMaxConnections = runtime.NumCPU()
)

type API struct {
	opts   *appdomain.Options_Mojang
	client *http.Client
}

func NewMojangAPI(opts *appdomain.Options_Mojang) *API {
	return &API{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        MojangAPIMaxConnections,
				MaxConnsPerHost:     MojangAPIMaxConnections,
				TLSHandshakeTimeout: MojangAPITLSHandshakeTimeout,
			},
			Timeout: MojangAPITimeout,
		},
		opts: opts,
	}
}

// GetGameProfile implements [domain.GameProfileRepository].
func (m *API) GetGameProfile(ctx context.Context, username string, hash string) (*mcdomain.GameProfile, error) {
	requestUrl, _ := url.Parse(strings.Join([]string{m.opts.SessionServerURL, MojangHasJoinedEndpoint}, "/"))
	requestUrl.RawQuery = url.Values{
		"username": {username},
		"serverId": {hash},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := m.client.Do(req)
	defer func() {
		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, appdomain.NewNotFoundError("player session")
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	joinedData := new(mcdomain.GameProfile)
	if err := json.Unmarshal(data, joinedData); err != nil {
		return nil, err
	}

	return joinedData, nil
}
