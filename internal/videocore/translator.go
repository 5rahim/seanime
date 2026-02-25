package videocore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"seanime/internal/mkvparser"
	"seanime/internal/util/result"
	"strings"
	"sync"
	"time"

	"github.com/5rahim/go-astisub"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

// Translator implemented by different providers
type Translator interface {
	TranslateBatch(ctx context.Context, texts []string, targetLang string) ([]string, error)
}

type TranslatorService struct {
	cache      *result.BoundedCache[string, string]
	translator Translator
	targetLang string
	vc         *VideoCore
	logSampler *zerolog.Logger
	queue      chan request
	close      chan struct{}
	closeOnce  sync.Once
}

func NewTranslatorService(vc *VideoCore, apiKey string, provider string, targetLang string) *TranslatorService {
	var t Translator
	if provider == "openai" {
		t = &OpenAITranslator{Token: apiKey, logger: vc.logger}
	} else if provider == "deepl" {
		t = &DeepLTranslator{Token: apiKey, logger: vc.logger}
	} else {
		t = NewFreeGoogleTranslator(vc.logger)
	}

	s := &TranslatorService{
		vc:         vc,
		translator: t,
		targetLang: targetLang,
		cache:      result.NewBoundedCache[string, string](10000),
		queue:      make(chan request, 1000),
		close:      make(chan struct{}),
		logSampler: new(vc.logger.Sample(&zerolog.BasicSampler{N: 500})),
	}

	go s.processQueue()

	return s
}

func (s *TranslatorService) Shutdown() {
	s.closeOnce.Do(func() {
		close(s.close)
	})
}

func (s *TranslatorService) TranslateContent(ctx context.Context, content string, format int, targetLang string) (string, error) {
	s.vc.logger.Debug().Msgf("videocore: Translating content of type %d to %s", format, targetLang)
	reader := strings.NewReader(content)
	var subs *astisub.Subtitles
	var err error

read:
	switch format {
	case mkvparser.SubtitleTypeASS:
		subs, err = astisub.ReadFromSSA(reader)
	case mkvparser.SubtitleTypeSSA:
		subs, err = astisub.ReadFromSSA(reader)
	case mkvparser.SubtitleTypeSRT:
		subs, err = astisub.ReadFromSRT(reader)
	case mkvparser.SubtitleTypeSTL:
		subs, err = astisub.ReadFromSTL(reader, astisub.STLOptions{IgnoreTimecodeStartOfProgramme: true})
	case mkvparser.SubtitleTypeTTML:
		subs, err = astisub.ReadFromTTML(reader)
	case mkvparser.SubtitleTypeWEBVTT:
		subs, err = astisub.ReadFromWebVTT(reader)
	case mkvparser.SubtitleTypeUnknown:
		detectedType := mkvparser.DetectSubtitleType(content)
		if detectedType == mkvparser.SubtitleTypeUnknown {
			s.vc.logger.Error().Msg("videocore: Failed to detect subtitle format")
			return "", fmt.Errorf("failed to detect subtitle format")
		}
		format = detectedType
		reader = strings.NewReader(content)
		goto read
	default:
		s.vc.logger.Error().Msgf("videocore: Unsupported subtitle format: %d", format)
		return "", fmt.Errorf("unsupported subtitle format: %d", format)
	}

	if err != nil {
		s.vc.logger.Error().Err(err).Msg("videocore: Failed to parse subtitles")
		return "", fmt.Errorf("parsing failed: %w", err)
	}

	type lineRef struct {
		itemIndex int
		cleaned   string
	}

	var linesToTranslate []lineRef

	// Scan items, check cache, and queue missing lines
	for i, item := range subs.Items {
		var textBuilder strings.Builder
		for _, line := range item.Lines {
			for _, lineItem := range line.Items {
				textBuilder.WriteString(lineItem.Text)
			}
			textBuilder.WriteString(" ")
		}
		fullText := strings.TrimSpace(textBuilder.String())

		if fullText == "" {
			continue
		}

		cleaned := cleanSubtitleText(fullText)
		cacheKey := generateCacheKey(cleaned, targetLang)

		if val, ok := s.cache.Get(cacheKey); ok {
			// Cache hit, update immediately
			updateItemText(item, val)
		} else {
			// Cache miss, queue it
			linesToTranslate = append(linesToTranslate, lineRef{
				itemIndex: i,
				cleaned:   cleaned,
			})
		}
	}

	// Process in batches
	batchSize := 50
	totalNeeded := len(linesToTranslate)

	for start := 0; start < totalNeeded; start += batchSize {
		end := start + batchSize
		if end > totalNeeded {
			end = totalNeeded
		}

		var batchTexts []string
		for k := start; k < end; k++ {
			batchTexts = append(batchTexts, linesToTranslate[k].cleaned)
		}

		translatedBatch, err := s.translator.TranslateBatch(ctx, batchTexts, targetLang)
		if err != nil {
			s.vc.logger.Error().Err(err).Msgf("videocore: Failed to translate batch at index %d", start)
			return "", fmt.Errorf("batch translation failed at index %d: %w", start, err)
		}

		// Map results back to original items and cache
		for k, translatedText := range translatedBatch {
			originalRef := linesToTranslate[start+k]

			cacheKey := generateCacheKey(originalRef.cleaned, targetLang)
			s.cache.Set(cacheKey, translatedText)

			updateItemText(subs.Items[originalRef.itemIndex], translatedText)
		}
	}

	s.vc.logger.Debug().Msgf("videocore: Translated %d lines", len(linesToTranslate))

	// Write output
	w := &bytes.Buffer{}
	switch format {
	case mkvparser.SubtitleTypeSSA, mkvparser.SubtitleTypeASS:
		err = subs.WriteToSSA(w)
	case mkvparser.SubtitleTypeSRT:
		err = subs.WriteToSRT(w)
	case mkvparser.SubtitleTypeSTL:
		err = subs.WriteToSTL(w)
	case mkvparser.SubtitleTypeTTML:
		err = subs.WriteToTTML(w)
	case mkvparser.SubtitleTypeWEBVTT:
		err = subs.WriteToWebVTT(w)
	default:
		err = subs.WriteToWebVTT(w)
	}

	if err != nil {
		s.vc.logger.Error().Err(err).Msg("videocore: Failed to write subtitles")
		return "", fmt.Errorf("failed to write subtitles: %w", err)
	}

	return w.String(), nil
}

// TranslateEvent handles single subtitle events from the mkv parser
func (s *TranslatorService) TranslateEvent(ctx context.Context, evt *mkvparser.SubtitleEvent, targetLang string) error {
	clean := cleanSubtitleText(evt.Text)
	if clean == "" {
		return nil
	}

	cacheKey := generateCacheKey(clean, targetLang)
	if val, ok := s.cache.Get(cacheKey); ok {
		evt.Text = val
		return nil
	}

	resCh := make(chan textResult, 1)

	select {
	case s.queue <- request{text: clean, targetLang: targetLang, resultChan: resCh}:
	case <-ctx.Done():
		return ctx.Err()
	}

	// block until the batch processor finishes (or timeout)
	select {
	case res := <-resCh:
		if res.err != nil {
			s.logSampler.Error().Err(res.err).Msg("videocore: Failed to translate subtitle event")
			return res.err
		}
		// save to cache
		s.cache.Set(cacheKey, res.text)
		evt.Text = res.text
		s.logSampler.Debug().Msgf("videocore: Translated subtitle event: %s", evt.Text)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(15 * time.Second):
		return fmt.Errorf("translation timed out")
	}
}

func (s *TranslatorService) TranslateText(ctx context.Context, text string, targetLang string) (string, error) {
	clean := cleanSubtitleText(text)
	if clean == "" {
		return "", nil
	}

	cacheKey := generateCacheKey(clean, targetLang)
	if val, ok := s.cache.Get(cacheKey); ok {
		return val, nil
	}

	resCh := make(chan textResult, 1)

	select {
	case s.queue <- request{text: clean, targetLang: targetLang, resultChan: resCh}:
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// block until the batch processor finishes (or timeout)
	select {
	case res := <-resCh:
		if res.err != nil {
			s.logSampler.Error().Err(res.err).Msg("videocore: Failed to translate text")
			return "", res.err
		}
		// save to cache
		s.cache.Set(cacheKey, res.text)
		s.logSampler.Debug().Msgf("videocore: Translated text: %s", res.text)
		return res.text, nil
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(15 * time.Second):
		return "", fmt.Errorf("translation timed out")
	}
}

//func (s *TranslatorService) TranslateEvent(ctx context.Context, evt *mkvparser.SubtitleEvent, targetLang string) error {
//	clean := cleanSubtitleText(evt.Text)
//	if clean == "" {
//		return nil
//	}
//
//	key := generateCacheKey(clean, targetLang)
//	if val, ok := s.cache.Load(key); ok {
//		evt.Text = val.(string)
//		return nil
//	}
//
//	tr, err := s.translator.TranslateBatch(ctx, []string{clean}, targetLang)
//	if err != nil || len(tr) == 0 {
//		s.logSampler.Error().Err(err).Msg("videocore: Failed to translate subtitle event")
//		return err
//	}
//
//	s.cache.Store(key, tr[0])
//	evt.Text = tr[0]
//	s.logSampler.Debug().Msgf("videocore: Translated subtitle event: %s", evt.Text)
//	return nil
//}

type request struct {
	text       string
	targetLang string
	resultChan chan textResult // The channel where we will send the answer
}

type textResult struct {
	text string
	err  error
}

func (s *TranslatorService) processQueue() {
	const maxBatchSize = 50
	const batchTimeout = 100 * time.Millisecond // Wait max to fill a batch

	// buffer holds the current batch of requests
	buffer := make([]request, 0, maxBatchSize)

	// Helper to flush the buffer
	flush := func() {
		if len(buffer) == 0 {
			return
		}
		// Copy buffer to separate slice to process async (so we don't block the queue)
		batch := make([]request, len(buffer))
		copy(batch, buffer)
		buffer = buffer[:0] // Reset buffer

		go s.executeBatch(batch)
	}

	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case req := <-s.queue:
			buffer = append(buffer, req)
			// If bus is full, leave immediately
			if len(buffer) >= maxBatchSize {
				flush()
				ticker.Reset(batchTimeout) // Reset timer since we just flushed
			}

		case <-ticker.C:
			// Timer expired, leave with whatever we have
			flush()

		case <-s.close:
			return
		}
	}
}

// executeBatch performs the actual API call
func (s *TranslatorService) executeBatch(batch []request) {
	if len(batch) == 0 {
		return
	}

	// We assume all in this batch target the same language
	targetLang := batch[0].targetLang
	texts := make([]string, len(batch))

	for i, req := range batch {
		texts[i] = req.text
	}

	// Call the API
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	translatedTexts, err := s.translator.TranslateBatch(ctx, texts, targetLang)

	// Distribute results back to the waiters
	for i, req := range batch {
		if err != nil {
			req.resultChan <- textResult{err: err}
		} else if i < len(translatedTexts) {
			req.resultChan <- textResult{text: translatedTexts[i]}
		} else {
			req.resultChan <- textResult{err: fmt.Errorf("missing translation result")}
		}
		close(req.resultChan)
	}
}

func updateItemText(item *astisub.Item, text string) {
	item.Lines = []astisub.Line{{
		Items: []astisub.LineItem{{
			Text: text,
		}},
	}}
}

func generateCacheKey(text, lang string) string {
	hash := sha256.Sum256([]byte(text + "|" + lang))
	return hex.EncodeToString(hash[:])
}

func cleanSubtitleText(input string) string {
	// Removes ASS tags like {\an8}
	re := regexp.MustCompile(`\{.*?\}`)
	return strings.TrimSpace(re.ReplaceAllString(input, ""))
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type DeepLTranslator struct {
	Token  string
	logger *zerolog.Logger
}

type deepLRequest struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
}

type deepLResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

func (d *DeepLTranslator) TranslateBatch(ctx context.Context, texts []string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	u := "https://api-free.deepl.com/v2/translate"
	if !strings.HasSuffix(d.Token, ":fx") {
		u = "https://api.deepl.com/v2/translate"
	}

	payload := deepLRequest{Text: texts, TargetLang: targetLang}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", "DeepL-Auth-Key "+d.Token)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("deepl API error: %d", resp.StatusCode)
		}

		var result deepLResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		output := make([]string, len(result.Translations))
		for j, t := range result.Translations {
			output[j] = t.Text
		}
		return output, nil
	}

	return nil, fmt.Errorf("deepl API rate limit exceeded")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type OpenAITranslator struct {
	Token  string
	logger *zerolog.Logger
}

type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (o *OpenAITranslator) TranslateBatch(ctx context.Context, texts []string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	u := "https://api.openai.com/v1/chat/completions"

	systemPrompt := fmt.Sprintf("You are an anime subtitle translator. Translate the following JSON array of strings to %s. Return ONLY a JSON array of strings without explanations. Maintain the order exactly.", strings.ToUpper(targetLang))
	userContent, _ := json.Marshal(texts)

	payload := openAIRequest{
		Model: "gpt-4o-mini",
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: string(userContent)},
		},
		Temperature: 0.3,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", "Bearer "+o.Token)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("openai API error: %d", resp.StatusCode)
		}

		var result openAIResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		if len(result.Choices) == 0 {
			return nil, fmt.Errorf("openai returned no choices")
		}

		// Parse the JSON array string back into a slice
		var translatedTexts []string
		content := strings.TrimSpace(result.Choices[0].Message.Content)

		// Clean up potential markdown formatting like ```json ... ```
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")

		if err := json.Unmarshal([]byte(content), &translatedTexts); err != nil {
			return nil, fmt.Errorf("failed to parse openai response: %w", err)
		}

		if len(translatedTexts) != len(texts) {
			return nil, fmt.Errorf("openai returned mismatching count: got %d, expected %d", len(translatedTexts), len(texts))
		}

		return translatedTexts, nil
	}

	return nil, fmt.Errorf("openai API rate limit exceeded")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FreeGoogleTranslator struct {
	limiter *rate.Limiter
	client  *http.Client
	logger  *zerolog.Logger
}

func NewFreeGoogleTranslator(logger *zerolog.Logger) *FreeGoogleTranslator {
	return &FreeGoogleTranslator{
		limiter: rate.NewLimiter(rate.Every(500*time.Millisecond), 50),
		client:  &http.Client{Timeout: 10 * time.Second},
		logger:  logger,
	}
}

func (g *FreeGoogleTranslator) TranslateBatch(ctx context.Context, texts []string, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	results := make([]string, len(texts))
	var wg sync.WaitGroup
	var errMutex sync.Mutex
	var firstErr error

	g.logger.Debug().Msgf("videocore: (google) Translating %d lines", len(texts))

	for i, text := range texts {
		wg.Add(1)

		go func(idx int, txt string) {
			defer wg.Done()

			if err := g.limiter.Wait(ctx); err != nil {
				return
			}

			// Add a tiny random jitter
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

			translated, err := g.translateSingle(ctx, txt, targetLang)
			if err != nil {
				errMutex.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errMutex.Unlock()
				return
			}
			results[idx] = translated
		}(i, text)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

func (g *FreeGoogleTranslator) translateSingle(ctx context.Context, text, targetLang string) (string, error) {
	endpoint := "https://translate.googleapis.com/translate_a/single"

	params := url.Values{}
	params.Add("client", "gtx")
	params.Add("sl", "auto")                      // Source Language: Auto
	params.Add("tl", strings.ToUpper(targetLang)) // Target Language
	params.Add("dt", "t")                         // Data Type: Translation
	params.Add("q", text)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return "", fmt.Errorf("google rate limited")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google api error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// The response is a messy JSON array of arrays: [[["Hola","Hello",null,null,1]],...]
	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result) > 0 {
		// The first element is an array of sentences
		if sentences, ok := result[0].([]interface{}); ok {
			var sb strings.Builder
			for _, s := range sentences {
				// Each sentence is an array where index 0 is the translated text
				if parts, ok := s.([]interface{}); ok && len(parts) > 0 {
					if translatedPart, ok := parts[0].(string); ok {
						sb.WriteString(translatedPart)
					}
				}
			}
			return sb.String(), nil
		}
	}

	return "", fmt.Errorf("failed to parse google response")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// TranslateContent translates the file content based on saved settings
func (vc *VideoCore) TranslateContent(ctx context.Context, content string, format int) string {
	if vc.translatorService == nil {
		return content
	}
	translated, err := vc.translatorService.TranslateContent(ctx, content, format, vc.translatorService.targetLang)
	if err != nil {
		vc.logger.Error().Err(err).Msg("videocore: Failed to translate content")
		return content
	}

	return translated
}

// TranslateEvent translates the subtitle event based on saved settings
func (vc *VideoCore) TranslateEvent(ctx context.Context, event *mkvparser.SubtitleEvent) {
	if vc.translatorService == nil {
		return
	}
	err := vc.translatorService.TranslateEvent(ctx, event, vc.translatorService.targetLang)
	if err != nil {
		return
	}
}

// TranslateText translates the text based on saved settings
func (vc *VideoCore) TranslateText(ctx context.Context, text string) string {
	if vc.translatorService == nil {
		return text
	}
	ret, err := vc.translatorService.TranslateText(ctx, text, vc.translatorService.targetLang)
	if err != nil {
		return text
	}
	return ret
}

func (vc *VideoCore) GetTranslationTargetLanguage() string {
	if vc.translatorService == nil {
		return ""
	}
	return vc.translatorService.targetLang
}
