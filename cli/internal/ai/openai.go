package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cli/internal/logger"
	"cli/internal/store"
)

type OpenAIClient struct {
	apiKey     string
	endpoint   string
	apiVersion string
	deployment string
	httpClient *http.Client
}

type OpenAIRequest struct {
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type AIRecommendation struct {
	Type           string  `json:"type"`
	DDL            string  `json:"ddl"`
	Rationale      string  `json:"rationale"`
	Confidence     float64 `json:"confidence"`
	ImpactEstimate string  `json:"impact_estimate"`
	RiskLevel      string  `json:"risk_level"`
	RewriteSQL     string  `json:"rewrite_sql,omitempty"`
}

type AIRecommendationResponse struct {
	Recommendations []AIRecommendation `json:"recommendations"`
	Analysis        string             `json:"analysis"`
}

func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiVersion := os.Getenv("AZURE_OPENAI_API_VERSION")
	deployment := os.Getenv("AZURE_OPENAI_CHAT_DEPLOYMENT_NAME")

	if apiKey == "" || endpoint == "" || apiVersion == "" || deployment == "" {
		return nil, fmt.Errorf("missing Azure OpenAI environment variables")
	}

	logger.LogInfof("Initializing Azure OpenAI client with endpoint: %s, deployment: %s", endpoint, deployment)

	return &OpenAIClient{
		apiKey:     apiKey,
		endpoint:   endpoint,
		apiVersion: apiVersion,
		deployment: deployment,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *OpenAIClient) GenerateRecommendations(query store.QueryStats, tables []store.TableInfo, indexes []store.IndexInfo) ([]store.Recommendation, error) {
	logger.LogInfof("Generating AI recommendations for query with %d calls, %.2fms avg time", query.Calls, query.MeanExecTime)

	prompt := c.buildRecommendationPrompt(query, tables, indexes)

	response, err := c.callOpenAI(prompt)
	if err != nil {
		logger.LogErrorf("Failed to call OpenAI API: %v", err)
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	recommendations, err := c.parseRecommendations(response)
	if err != nil {
		logger.LogErrorf("Failed to parse AI recommendations: %v", err)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	logger.LogInfof("Generated %d AI-powered recommendations", len(recommendations))
	return recommendations, nil
}

func (c *OpenAIClient) buildRecommendationPrompt(query store.QueryStats, tables []store.TableInfo, indexes []store.IndexInfo) string {
	// Convert data to JSON for structured input
	tablesJSON, _ := json.MarshalIndent(tables, "", "  ")
	indexesJSON, _ := json.MarshalIndent(indexes, "", "  ")

	prompt := fmt.Sprintf(`You are an expert PostgreSQL performance analyst. Analyze the following slow query and database metadata to provide actionable optimization recommendations.

QUERY PERFORMANCE DATA:
- SQL: %s
- Calls: %d
- Mean Execution Time: %.2f ms
- Total Time: %.2f ms
- Rows Returned: %d
- Shared Blocks Hit: %d
- Shared Blocks Read: %d

DATABASE TABLES:
%s

DATABASE INDEXES:
%s

ANALYSIS REQUIREMENTS:
1. Identify specific performance bottlenecks in this query
2. Suggest concrete optimizations with DDL statements
3. Provide confidence scores (0.0-1.0) based on data evidence
4. Estimate performance impact in plain English
5. Assess risk level (low/medium/high) for each recommendation

RECOMMENDATION TYPES TO CONSIDER:
- missing_index: Single column indexes for WHERE/ORDER BY clauses
- composite_index: Multi-column indexes for complex WHERE clauses and JOINs
- covering_index: Include columns to avoid table lookups
- join_index: Indexes to optimize JOIN performance
- correlated_subquery: Query rewrite suggestions (JOIN/EXISTS alternatives)
- redundant_index: Identify unused or duplicate indexes
- cardinality_issue: Statistics or data distribution problems
- query_rewrite: Alternative SQL formulations

RESPONSE FORMAT (JSON only, no markdown):
{
  "recommendations": [
    {
      "type": "missing_index",
      "ddl": "CREATE INDEX idx_table_column ON table_name (column_name);",
      "rationale": "Detailed explanation of why this helps performance",
      "confidence": 0.85,
      "impact_estimate": "Expected 50-80%% performance improvement",
      "risk_level": "low",
      "rewrite_sql": "Alternative SQL if applicable"
    }
  ],
  "analysis": "Overall performance analysis summary"
}

Provide only valid JSON response. Focus on actionable, high-impact recommendations based on the actual query patterns and database structure.`,
		query.Query, query.Calls, query.MeanExecTime, query.TotalTime, query.Rows, query.SharedBlksHit, query.SharedBlksRead,
		string(tablesJSON), string(indexesJSON))

	return prompt
}

func (c *OpenAIClient) callOpenAI(prompt string) (string, error) {
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		c.endpoint, c.deployment, c.apiVersion)

	reqBody := OpenAIRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are an expert PostgreSQL performance analyst. Respond only with valid JSON.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.1, // Low temperature for consistent, factual responses
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.LogDebugf("Sending request to Azure OpenAI: %s", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.LogErrorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal OpenAI response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	logger.LogInfof("OpenAI API call successful. Tokens used: %d", openAIResp.Usage.TotalTokens)
	return openAIResp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) parseRecommendations(response string) ([]store.Recommendation, error) {
	var aiResp AIRecommendationResponse
	if err := json.Unmarshal([]byte(response), &aiResp); err != nil {
		logger.LogErrorf("Failed to parse AI response JSON: %v", err)
		logger.LogDebugf("Raw AI response: %s", response)
		return nil, fmt.Errorf("invalid JSON response from AI: %w", err)
	}

	recommendations := make([]store.Recommendation, 0, len(aiResp.Recommendations))

	for _, aiRec := range aiResp.Recommendations {
		rec := store.Recommendation{
			Type:           aiRec.Type,
			DDL:            aiRec.DDL,
			RewriteSQL:     aiRec.RewriteSQL,
			Rationale:      aiRec.Rationale,
			Confidence:     aiRec.Confidence,
			ImpactEstimate: aiRec.ImpactEstimate,
			RiskLevel:      aiRec.RiskLevel,
			CreatedAt:      time.Now(),
		}

		// Validate confidence score
		if rec.Confidence < 0.0 || rec.Confidence > 1.0 {
			logger.LogDebugf("Adjusting invalid confidence score: %.2f -> 0.5", rec.Confidence)
			rec.Confidence = 0.5
		}

		// Validate risk level
		if rec.RiskLevel != "low" && rec.RiskLevel != "medium" && rec.RiskLevel != "high" {
			logger.LogDebugf("Adjusting invalid risk level: %s -> medium", rec.RiskLevel)
			rec.RiskLevel = "medium"
		}

		recommendations = append(recommendations, rec)
	}

	return recommendations, nil
}
