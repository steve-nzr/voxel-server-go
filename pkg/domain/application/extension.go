package application

const (
	ActorFactoryID          = "ActorFactory"
	ConfigProviderID        = "ConfigProvider"
	GameProfileRepositoryID = "GameProfileRepository"
)

type ActorFactoryExtension struct{}

func (e ActorFactoryExtension) ID() string {
	return ActorFactoryID
}

type ConfigProviderExtension struct{}

func (e ConfigProviderExtension) ID() string {
	return ConfigProviderID
}

type GameProfileRepositoryExtension struct{}

func (e GameProfileRepositoryExtension) ID() string {
	return GameProfileRepositoryID
}
