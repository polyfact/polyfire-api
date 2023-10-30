package providers

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/polyfire/api/llm/providers/options"
	replicate "github.com/polyfire/api/llm/providers/replicate"
	"github.com/polyfire/api/utils"
)

type ReplicateProvider struct {
	Model           string
	ReplicateApiKey string
	IsCustomApiKey  bool
}

func NewReplicateProvider(ctx context.Context, model string) ReplicateProvider {
	var apiKey string

	customToken, ok := ctx.Value(utils.ContextKeyReplicateToken).(string)
	if ok {
		apiKey = customToken
	} else {
		apiKey = os.Getenv("REPLICATE_API_KEY")
	}

	return ReplicateProvider{Model: model, ReplicateApiKey: apiKey, IsCustomApiKey: ok}
}

func (m ReplicateProvider) GetCreditsPerSecond() float64 {
	switch m.Model {
	case "llama-2-70b-chat":
		return 14000.0
	case "replit-code-v1-3b":
		return 11500.0
	case "wizard-mega-13b-awq":
		return 7250.0
	default:
		fmt.Printf("Invalid model: %v\n", m.Model)
		return 0.0
	}
}

func (m ReplicateProvider) GetVersion() (string, bool, error) {
	switch m.Model {
	case "llama-2-70b-chat":
		return "02e509c789964a7ea8736978a43525956ef40397be9033abf9fd2badfe68c9e3", true, nil
	case "replit-code-v1-3b":
		return "b84f4c074b807211cd75e3e8b1589b6399052125b4c27106e43d47189e8415ad", true, nil
	case "wizard-mega-13b-awq":
		return "a4be2a7c75e51c53b22167d44de3333436f1aa9253a201d2619cf74286478599", false, nil
	default:
		return "", false, errors.New("Invalid model")
	}
}

func (m ReplicateProvider) Generate(
	task string,
	c options.ProviderCallback,
	opts *options.ProviderOptions,
) chan options.Result {
	version, stream, err := m.GetVersion()
	if err != nil {
		fmt.Printf("Error getting version: %v\n", err)
		return nil
	}

	replicateProvider := replicate.ReplicateProvider{
		Model:            m.Model,
		ReplicateApiKey:  m.ReplicateApiKey,
		IsCustomApiKey:   m.IsCustomApiKey,
		Version:          version,
		CreditsPerSecond: m.GetCreditsPerSecond(),
	}

	var chan_res chan options.Result
	if stream {
		chan_res = replicateProvider.Stream(task, c, opts)
	} else {
		chan_res = replicateProvider.NoStream(task, c, opts)
	}

	return chan_res
}

func (m ReplicateProvider) UserAllowed(_user_id string) bool {
	return true
}

func (m ReplicateProvider) Name() string {
	return "replicate"
}

func (m ReplicateProvider) ProviderModel() (string, string) {
	return "replicate", m.Model
}

func (m ReplicateProvider) DoesFollowRateLimit() bool {
	return !m.IsCustomApiKey
}
