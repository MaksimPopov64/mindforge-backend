package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday" // Для очистки HTML
)

// Политика очистки HTML
var htmlPolicy = bluemonday.UGCPolicy()

// Структура для возврата и текста, и тегов
type ProcessedContent struct {
	OriginalURL   string `json:"original_url,omitempty"`
	Title         string `json:"title,omitempty"`
	PlainText     string `json:"plain_text"`
	HTMLContent   string `json:"html_content"`
	GeneratedTags string `json:"generated_tags"`
}

// Функция для получения HTML контента из веб-страницы
func fetchHTMLContent(urlStr string) (*ProcessedContent, error) {
	result := &ProcessedContent{OriginalURL: urlStr}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("неверный URL: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// Добавляем User-Agent для обхода блокировок
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения страницы: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("не HTML контент: %s", contentType)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга HTML: %w", err)
	}

	// Извлекаем заголовок
	title := extractTitle(doc)
	result.Title = title

	// Извлекаем основной контент
	htmlContent, plainText := extractMainContent(doc)

	// Очищаем HTML от опасных тегов
	cleanedHTML := htmlPolicy.Sanitize(htmlContent)

	// Улучшаем HTML форматирование
	cleanedHTML = formatHTML(cleanedHTML, title)

	result.HTMLContent = cleanedHTML
	result.PlainText = plainText

	return result, nil
}

// Функция извлечения заголовка
func extractTitle(doc *goquery.Document) string {
	titleSelectors := []string{
		"h1.tm-title",       // Habr
		"h1.article__title", // VC.ru
		"h1.post__title",    // VC.ru альтернативный
		"article h1",        // Общий тег
		"h1",                // Любой h1
		"title",             // Из тега title
	}

	for _, selector := range titleSelectors {
		if found := doc.Find(selector).First().Text(); found != "" {
			return strings.TrimSpace(found)
		}
	}
	return ""
}

// Функция извлечения основного контента
func extractMainContent(doc *goquery.Document) (string, string) {
	var contentSelection *goquery.Selection

	// Попробуем найти по специфичным селекторам для популярных сайтов
	contentSelectors := []string{
		"div.article-formatted-body", // Habr
		"div.content.content--full",  // VC.ru
		"div.article-content",        // TJournal
		"article",                    // Стандартный тег статьи
		"div.post-content",           // WordPress
		"main",                       // Основной контент
		"div.content",                // Общий селектор
	}

	for _, selector := range contentSelectors {
		if selection := doc.Find(selector).First(); selection.Length() > 0 {
			contentSelection = selection
			break
		}
	}

	// Если не нашли специфичный контент, ищем по эвристикам
	if contentSelection == nil {
		doc.Find("div, section, article").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			// Эвристика: блок с большим количеством текста и предложений
			if len(text) > 500 && strings.Count(text, ".") > 5 {
				if contentSelection == nil || len(text) > len(contentSelection.Text()) {
					contentSelection = s
				}
			}
		})
	}

	// Если все еще не нашли, используем body
	if contentSelection == nil {
		contentSelection = doc.Find("body")
	}

	// Удаляем ненужные элементы
	contentSelection.Find(`
		script, style, nav, iframe, 
		.comments, .tm-comments, footer, 
		.ad, .ads, .advertisement,
		.social-share, .share, 
		.header, .menu, .sidebar,
		form, button, input,
		[class*="comment"], [id*="comment"],
		[class*="ad"], [id*="ad"],
		[class*="banner"], [id*="banner"]
	`).Remove()

	// Получаем HTML и текст
	htmlContent, _ := contentSelection.Html()
	plainText := contentSelection.Text()

	// Очищаем текст
	plainText = strings.TrimSpace(plainText)
	plainText = strings.Join(strings.Fields(plainText), " ")

	// Ограничиваем длину
	if len(plainText) > 10000 {
		plainText = plainText[:10000] + "..."
	}

	return htmlContent, plainText
}

// Функция форматирования HTML
func formatHTML(html string, title string) string {
	// Удаляем пустые теги
	html = removeEmptyTags(html)

	// Оборачиваем в контейнер с базовыми стилями
	formattedHTML := fmt.Sprintf(`
		<div class="parsed-content" style="
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			line-height: 1.7;
			color: #333;
			max-width: 100%%;
			overflow-wrap: break-word;
		">
			%s
		</div>
	`, html)

	return formattedHTML
}

// Функция удаления пустых тегов
func removeEmptyTags(html string) string {
	patterns := []string{
		`<p>\s*</p>`,
		`<div>\s*</div>`,
		`<span>\s*</span>`,
		`<strong>\s*</strong>`,
		`<em>\s*</em>`,
		`<h[1-6]>\s*</h[1-6]>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		html = re.ReplaceAllString(html, "")
	}

	return html
}

func ProcessNoteContent(input string) (*ProcessedContent, error) {
	result := &ProcessedContent{}

	// Проверяем, является ли ввод URL
	if strings.HasPrefix(strings.ToLower(input), "http://") ||
		strings.HasPrefix(strings.ToLower(input), "https://") {

		// Это URL - парсим страницу
		processed, err := fetchHTMLContent(input)
		if err != nil {
			log.Printf("⚠️ Ошибка парсинга HTML: %v", err)
			// Если не удалось, используем как обычный текст
			result.PlainText = fmt.Sprintf("Ссылка: %s\n\nНе удалось автоматически получить контент страницы.", input)
			result.HTMLContent = fmt.Sprintf(`<div class="parsed-content"><p>%s</p></div>`, result.PlainText)
		} else {
			result = processed
		}
	} else {
		// Это обычный текст
		result.PlainText = input
		result.HTMLContent = fmt.Sprintf(`<div class="parsed-content"><p>%s</p></div>`, input)
	}

	// Генерируем теги на основе чистого текста
	if strings.TrimSpace(result.PlainText) == "" {
		return result, fmt.Errorf("текст для анализа пуст")
	}

	tags, err := AskOllamaForTags(result.PlainText)
	if err != nil {
		log.Printf("⚠️ Ошибка генерации тегов: %v", err)
		tags = "без тегов"
	}

	result.GeneratedTags = tags

	return result, nil
}

// Обновленная функция AskOllamaForTags (оставляем как есть, но можно использовать для чистого текста)
func AskOllamaForTags(noteText string) (string, error) {
	if strings.TrimSpace(noteText) == "" {
		return "", fmt.Errorf("текст заметки пуст")
	}

	// 1. Формируем промпт для генерации тегов
	prompt := `Проанализируй текст заметки и верни 3-5 ключевых тегов на русском. 
Теги должны быть короткими словами или фразами. 
Верни ТОЛЬКО теги, разделенные запятыми, без пояснений, точек и нумерации.

Текст заметки: ` + noteText

	// 2. Структура запроса к API /api/generate
	requestData := map[string]interface{}{
		"model":  "llama3.2",
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.3,
			"num_predict": 50,
		},
	}

	// 3. Кодируем запрос в JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("ошибка кодирования запроса: %w", err)
	}

	// 4. Отправляем POST-запрос к локальному серверу Ollama
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка HTTP-запроса к Ollama: %w", err)
	}
	defer resp.Body.Close()

	// 5. Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// 6. Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama вернул ошибку HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 7. Парсим JSON ответа
	var ollamaResp struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("ошибка парсинга ответа: %w", err)
	}
	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ошибка в ответе ollama: %s", ollamaResp.Error)
	}

	// 8. Очищаем ответ
	cleanTags := strings.TrimSpace(ollamaResp.Response)
	cleanTags = strings.TrimSuffix(cleanTags, ".")
	cleanTags = strings.ReplaceAll(cleanTags, "\n", ", ")
	cleanTags = strings.Join(strings.Fields(cleanTags), " ")

	return cleanTags, nil
}
