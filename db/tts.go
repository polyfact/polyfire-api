package db

type TTSVoice struct {
	ID              string `json:"id"`
	Slug            string `json:"slug"`
	Provider        string `json:"provider"`
	ProviderVoiceID string `json:"provider_voice_id"`
}

func (TTSVoice) TableName() string {
	return "tts_voices"
}

func GetTTSVoice(slug string) (TTSVoice, error) {
	var voice TTSVoice
	err := DB.First(&voice, "slug = ?", slug).Error
	if err != nil {
		return voice, err
	}

	return voice, nil
}
