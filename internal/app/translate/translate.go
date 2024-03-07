package translate

import (
	"context"
	"fmt"

	translate "cloud.google.com/go/translate/apiv3"
	"cloud.google.com/go/translate/apiv3/translatepb"
	"github.com/dennishilgert/cloud-computing-2/pkg/logger"
)

var log = logger.NewLogger("app.translator")

type Options struct {
	ProjectId string
}

type Translator interface {
	AvailableLanguages() AvailableLanguages
	Translate(ctx context.Context, sourceLang string, targetLang string, input string) (*string, error)
	Close()
}

type translator struct {
	projectId          string
	client             *translate.TranslationClient
	availableLanguages AvailableLanguages
}

func NewTranslator(ctx context.Context, opts Options) Translator {
	log.Info("creating cloud translation api client")
	client, err := translate.NewTranslationClient(ctx)
	if err != nil {
		log.Fatalf("failed to create cloud translation api client: %v", err)
	}

	log.Info("loading available languages from cloud translation api")
	supLangReq := &translatepb.GetSupportedLanguagesRequest{
		Parent:              fmt.Sprintf("projects/%s/locations/global", opts.ProjectId),
		DisplayLanguageCode: "en",
	}
	supLangRes, err := client.GetSupportedLanguages(ctx, supLangReq)
	if err != nil {
		log.Fatalf("failed to load supported languages of cloud translation api: %v", err)
	}

	return &translator{
		projectId:          opts.ProjectId,
		client:             client,
		availableLanguages: ParseAvailableLanguages(supLangRes.GetLanguages()),
	}
}

func (t *translator) AvailableLanguages() AvailableLanguages {
	return t.availableLanguages
}

func (t *translator) Translate(ctx context.Context, sourceLang string, targetLang string, input string) (*string, error) {
	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/global", t.projectId),
		SourceLanguageCode: sourceLang,
		TargetLanguageCode: targetLang,
		MimeType:           "text/plain", // Mime types: "text/plain", "text/html"
		Contents:           []string{input},
	}
	resp, err := t.client.TranslateText(ctx, req)
	if err != nil {
		return nil, err
	}
	return &resp.GetTranslations()[0].TranslatedText, nil
}

func (t *translator) Close() {
	t.client.Close()
}
