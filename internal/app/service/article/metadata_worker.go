package article

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	gocron "github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
	"golang.org/x/net/html/charset"

	"github.com/sappy5678/DeeliAi/internal/domain/article"
)

type MetadataWorker struct {
	scheduler gocron.Scheduler
	service   *articleService
}

func NewMetadataWorker(service *articleService) *MetadataWorker {
	s, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return nil
	}
	w := &MetadataWorker{
		scheduler: s,
		service:   service,
	}
	w.scheduler.NewJob(
		gocron.DurationJob(1*time.Minute),
		gocron.NewTask(w.runMetadataFetchJob, context.Background()),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithName("MetadataWorker"),
	)

	w.scheduler.Start()

	return w
}

func (w *MetadataWorker) fetchFail(ctx context.Context, retry *article.MetadataFetchRetry, err error) {
	errorMessage := err.Error()
	retry.RetryCount++
	err = w.service.articleRepo.IncrementMetadataFetchRetryCount(ctx, retry.ID)
	if err != nil {
		w.logger(ctx).Err(err).Str("retry_id", fmt.Sprintf("%d", retry.ID)).Msg("failed to increment metadata fetch retry count")
	}
	if retry.RetryCount >= 3 {
		err = w.service.articleRepo.UpdateMetadataFetchRetryStatus(ctx, retry.ID, 2, errorMessage)
		if err != nil {
			w.logger(ctx).Err(err).Str("retry_id", fmt.Sprintf("%d", retry.ID)).Msg("failed to update metadata fetch retry status to failed")
		}
	}
}

func (w *MetadataWorker) runMetadataFetchJob(ctx context.Context) {
	w.logger(ctx).Info().Msg("running metadata fetch job")

	retries, err := w.service.articleRepo.GetPendingMetadataFetchRetries(ctx)
	if err != nil {
		w.logger(ctx).Err(err).Msg("failed to get pending metadata fetch retries")
		return
	}

	if len(retries) == 0 {
		w.logger(ctx).Info().Msg("no pending metadata fetch retries")
		return
	}

	for _, retry := range retries {
		w.logger(ctx).Info().Str("article_id", retry.ArticleID.String()).Int64("retry_id", retry.ID).Msg("attempting to fetch metadata for article")

		info, err := fetchMetadata(retry.URL)
		if err != nil {
			w.logger(ctx).Err(err).Str("url", retry.URL).Msg("failed to fetch metadata")
			w.fetchFail(ctx, retry, err)
			continue
		}

		// Metadata fetched successfully
		metadataBytes, err := json.Marshal(info)
		if err != nil {
			w.logger(ctx).Err(err).Str("url", retry.URL).Msg("failed to marshal metadata")
			w.fetchFail(ctx, retry, err)
			continue
		}

		art, err := w.service.articleRepo.GetArticleByID(ctx, retry.ArticleID)
		if err != nil {
			w.logger(ctx).Err(err).Str("article_id", retry.ArticleID.String()).Msg("failed to get article by ID")
			w.fetchFail(ctx, retry, err)
			continue
		}
				art.Title = info.Title
		art.Description = info.Description
		if len(info.Metadata["og:image"]) > 0 {
			art.ImageURL = info.Metadata["og:image"][0]
		}
		art.Metadata = metadataBytes
		err = w.service.articleRepo.UpdateArticle(ctx, art)
		if err != nil {
			w.logger(ctx).Err(err).Str("article_id", retry.ArticleID.String()).Msg("failed to update article metadata")
			w.fetchFail(ctx, retry, err)
			continue
		}
		err = w.service.articleRepo.UpdateMetadataFetchRetryStatus(ctx, retry.ID, 1, "")
		if err != nil {
			w.logger(ctx).Err(err).Str("retry_id", fmt.Sprintf("%d", retry.ID)).Msg("failed to update metadata fetch retry status to success")
		}

	}
}

// logger wrap the execution context with component info
func (s *MetadataWorker) logger(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx).With().Str("component", "metadata-worker").Logger()
	return &l
}

type BasicMeta struct {
	Title       string
	Description string
	Metadata    map[string][]string
}

// This task should put into message queue, so that it can be processed by workers concurrently
// but for simplicity, we just run it in a single worker here.
func fetchMetadata(pageURL string) (*BasicMeta, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", pageURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MetaMini/1.0)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		r = resp.Body
	}

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	// title：get <title> first，if not find try og:title
	title := strings.TrimSpace(doc.Find("title").First().Text())
	if title == "" {
		if v, ok := doc.Find(`meta[property="og:title"]`).Attr("content"); ok {
			title = strings.TrimSpace(v)
		}
	}

	// description：get meta[name=description] first，if not find try og:description
	description := ""
	if v, ok := doc.Find(`meta[name="description"]`).Attr("content"); ok {
		description = strings.TrimSpace(v)
	}
	if description == "" {
		if v, ok := doc.Find(`meta[property="og:description"]`).Attr("content"); ok {
			description = strings.TrimSpace(v)
		}
	}

	// Open Graph： get all og:*
	metadata := make(map[string][]string)
	doc.Find(`meta[property^="og:"]`).Each(func(_ int, s *goquery.Selection) {
		prop, _ := s.Attr("property")
		prop = strings.ToLower(strings.TrimSpace(prop))
		if prop == "" {
			return
		}
		if content, ok := s.Attr("content"); ok {
			val := strings.TrimSpace(content)
			if val == "" {
				return
			}
			metadata[prop] = append(metadata[prop], val)
		}
	})

	return &BasicMeta{
		Title:       title,
		Description: description,
		Metadata:    metadata,
	}, nil
}
