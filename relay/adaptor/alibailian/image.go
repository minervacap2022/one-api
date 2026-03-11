package alibailian

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// --- Request types ---

// QwenImageSyncRequest is the DashScope multimodal-generation request format
// for qwen-image-2.0 series (synchronous).
type QwenImageSyncRequest struct {
	Model      string               `json:"model"`
	Input      QwenImageSyncInput   `json:"input"`
	Parameters QwenImageSyncParams  `json:"parameters,omitempty"`
}

type QwenImageSyncInput struct {
	Messages []QwenImageSyncMessage `json:"messages"`
}

type QwenImageSyncMessage struct {
	Role    string                     `json:"role"`
	Content []QwenImageSyncContentPart `json:"content"`
}

type QwenImageSyncContentPart struct {
	Text string `json:"text,omitempty"`
}

type QwenImageSyncParams struct {
	Size           string `json:"size,omitempty"`
	N              int    `json:"n,omitempty"`
	PromptExtend   *bool  `json:"prompt_extend,omitempty"`
	Watermark      *bool  `json:"watermark,omitempty"`
	Seed           int    `json:"seed,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
}

// AsyncImageRequest is the DashScope text2image request format
// for legacy models (qwen-image-plus, qwen-image).
type AsyncImageRequest struct {
	Model string `json:"model"`
	Input struct {
		Prompt         string `json:"prompt"`
		NegativePrompt string `json:"negative_prompt,omitempty"`
	} `json:"input"`
	Parameters struct {
		Size  string `json:"size,omitempty"`
		N     int    `json:"n,omitempty"`
		Steps string `json:"steps,omitempty"`
		Scale string `json:"scale,omitempty"`
	} `json:"parameters"`
}

// --- Response types ---

// QwenImageSyncResponse is the DashScope multimodal-generation response.
type QwenImageSyncResponse struct {
	RequestId string `json:"request_id,omitempty"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Output    struct {
		Choices []struct {
			Message struct {
				Content []struct {
					Image string `json:"image,omitempty"`
				} `json:"content"`
			} `json:"message"`
		} `json:"choices,omitempty"`
	} `json:"output"`
	Usage struct {
		ImageCount int `json:"image_count,omitempty"`
	} `json:"usage"`
}

// AsyncTaskResponse is the DashScope async task response (submit + poll).
type AsyncTaskResponse struct {
	RequestId string `json:"request_id,omitempty"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Output    struct {
		TaskId     string `json:"task_id,omitempty"`
		TaskStatus string `json:"task_status,omitempty"`
		Code       string `json:"code,omitempty"`
		Message    string `json:"message,omitempty"`
		Results    []struct {
			B64Image string `json:"b64_image,omitempty"`
			Url      string `json:"url,omitempty"`
		} `json:"results,omitempty"`
	} `json:"output"`
}

// OpenAI-compatible response types (avoid importing openai package)

type imageData struct {
	Url           string `json:"url,omitempty"`
	B64Json       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type imageResponse struct {
	Created int64       `json:"created"`
	Data    []imageData `json:"data"`
}

// --- Conversion ---

// ConvertImageRequest converts an OpenAI ImageRequest to the appropriate
// DashScope format based on the model.
func ConvertImageRequest(request model.ImageRequest) any {
	if IsQwenImageSyncModel(request.Model) {
		return convertSyncRequest(request)
	}
	return convertAsyncRequest(request)
}

func convertSyncRequest(request model.ImageRequest) *QwenImageSyncRequest {
	size := strings.Replace(request.Size, "x", "*", -1)
	if size == "" {
		size = "1024*1024"
	}
	n := request.N
	if n == 0 {
		n = 1
	}
	watermark := false

	return &QwenImageSyncRequest{
		Model: request.Model,
		Input: QwenImageSyncInput{
			Messages: []QwenImageSyncMessage{
				{
					Role: "user",
					Content: []QwenImageSyncContentPart{
						{Text: request.Prompt},
					},
				},
			},
		},
		Parameters: QwenImageSyncParams{
			Size:      size,
			N:         n,
			Watermark: &watermark,
		},
	}
}

func convertAsyncRequest(request model.ImageRequest) *AsyncImageRequest {
	var req AsyncImageRequest
	req.Model = request.Model
	req.Input.Prompt = request.Prompt
	req.Parameters.Size = strings.Replace(request.Size, "x", "*", -1)
	req.Parameters.N = request.N
	return &req
}

// --- Response handling ---

// ImageHandler routes to the correct handler based on model type.
func ImageHandler(c *gin.Context, resp *http.Response, meta *meta.Meta) (*model.ErrorWithStatusCode, *model.Usage) {
	if IsQwenImageSyncModel(meta.ActualModelName) {
		return syncImageHandler(c, resp)
	}
	return asyncImageHandler(c, resp)
}

func syncImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	responseFormat := c.GetString("response_format")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return wrapError(err, "read_response_body_failed"), nil
	}
	_ = resp.Body.Close()

	var qwenResp QwenImageSyncResponse
	if err := json.Unmarshal(body, &qwenResp); err != nil {
		return wrapError(err, "unmarshal_response_body_failed"), nil
	}

	if qwenResp.Code != "" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message:  qwenResp.Message,
				Type:     "ali_error",
				Code:     qwenResp.Code,
				RawError: errors.New(qwenResp.Message),
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	imgResp := imageResponse{Created: helper.GetTimestamp()}
	for _, choice := range qwenResp.Output.Choices {
		for _, content := range choice.Message.Content {
			if content.Image == "" {
				continue
			}
			d := imageData{Url: content.Image}
			if responseFormat == "b64_json" {
				raw, dlErr := downloadImage(content.Image)
				if dlErr != nil {
					continue
				}
				d.B64Json = base64.StdEncoding.EncodeToString(raw)
				d.Url = ""
			}
			imgResp.Data = append(imgResp.Data, d)
		}
	}

	return writeJSON(c, resp.StatusCode, imgResp)
}

func asyncImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	apiKey := c.Request.Header.Get("Authorization")
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")
	responseFormat := c.GetString("response_format")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return wrapError(err, "read_response_body_failed"), nil
	}
	_ = resp.Body.Close()

	var taskResp AsyncTaskResponse
	if err := json.Unmarshal(body, &taskResp); err != nil {
		return wrapError(err, "unmarshal_response_body_failed"), nil
	}

	if taskResp.Message != "" {
		return wrapError(errors.Errorf("ali async task failed: %s", taskResp.Message), "ali_async_task_failed"), nil
	}

	// Poll for completion
	result, err := asyncTaskWait(taskResp.Output.TaskId, apiKey)
	if err != nil {
		return wrapError(err, "ali_async_task_wait_failed"), nil
	}

	if result.Output.TaskStatus != "SUCCEEDED" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message:  result.Output.Message,
				Type:     "ali_error",
				Code:     result.Output.Code,
				RawError: errors.New(result.Output.Message),
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	imgResp := imageResponse{Created: helper.GetTimestamp()}
	for _, r := range result.Output.Results {
		d := imageData{Url: r.Url, B64Json: r.B64Image}
		if responseFormat == "b64_json" && d.B64Json == "" && d.Url != "" {
			raw, dlErr := downloadImage(d.Url)
			if dlErr != nil {
				continue
			}
			d.B64Json = base64.StdEncoding.EncodeToString(raw)
		}
		imgResp.Data = append(imgResp.Data, d)
	}

	return writeJSON(c, resp.StatusCode, imgResp)
}

// --- Async polling ---

func asyncTaskWait(taskID, apiKey string) (*AsyncTaskResponse, error) {
	const maxSteps = 20
	const waitSeconds = 2

	for step := 0; step < maxSteps; step++ {
		resp, err := pollTask(taskID, apiKey)
		if err != nil {
			return nil, err
		}

		switch resp.Output.TaskStatus {
		case "SUCCEEDED", "FAILED", "CANCELED", "UNKNOWN", "":
			return resp, nil
		}

		time.Sleep(waitSeconds * time.Second)
	}

	return nil, errors.Errorf("ali async task wait timeout for task %s", taskID)
}

func pollTask(taskID, apiKey string) (*AsyncTaskResponse, error) {
	url := fmt.Sprintf("https://dashscope.aliyuncs.com/api/v1/tasks/%s", taskID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result AsyncTaskResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Helpers ---

func downloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "download image from %s", url)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func wrapError(err error, code string) *model.ErrorWithStatusCode {
	return &model.ErrorWithStatusCode{
		Error: model.Error{
			Message:  err.Error(),
			Type:     "ali_error",
			Code:     code,
			RawError: err,
		},
		StatusCode: http.StatusInternalServerError,
	}
}

func writeJSON(c *gin.Context, statusCode int, v any) (*model.ErrorWithStatusCode, *model.Usage) {
	data, err := json.Marshal(v)
	if err != nil {
		return wrapError(err, "marshal_response_body_failed"), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(statusCode)
	_, _ = c.Writer.Write(data)
	return nil, nil
}
