package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client is an HTTP client for the Frends Platform API.
type Client struct {
	BaseURL    string
	Token      string
	httpClient *http.Client
}

// NewClient creates a new Frends API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest executes an HTTP request and decodes the JSON response into result.
// If result is nil, the response body is discarded.
// Returns (nil, nil) when status is 404 (not found).
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshalling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{Path: path}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error %d for %s %s: %s", resp.StatusCode, method, path, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decoding response from %s %s: %w", method, path, err)
		}
	}

	return nil
}

// NotFoundError is returned when the API responds with 404.
type NotFoundError struct {
	Path string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource not found: %s", e.Path)
}

// IsNotFound returns true if the error is a NotFoundError.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}

// --- Process Deployments ---

// CreateProcessDeployment creates a new process deployment.
// Uses a raw request because the API returns 201 (not 200) and we must read the body regardless.
func (c *Client) CreateProcessDeployment(ctx context.Context, req ProcessDeploymentCreate) (*ProcessDeploymentGet, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/api/v1/process-deployments", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.Token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading process deployment response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Response body may contain the created deployment object.
	var result ProcessDeploymentGet
	if len(body) > 0 {
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("decoding create response: %w", err)
		}
	}
	return &result, nil
}

// GetProcessDeployment retrieves a process deployment by ID.
func (c *Client) GetProcessDeployment(ctx context.Context, id int64) (*ProcessDeploymentGet, error) {
	var result ProcessDeploymentGet
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/process-deployments/%d", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListProcessDeployments retrieves all process deployments for an agent group.
func (c *Client) ListProcessDeployments(ctx context.Context, agentGroupID int64) ([]ProcessDeploymentGet, error) {
	var result ProcessDeploymentListResponse
	err := c.doRequest(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/process-deployments?agentGroupId=%d", agentGroupID), nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// DeleteProcessDeployment deletes a process deployment by ID.
func (c *Client) DeleteProcessDeployment(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/process-deployments/%d", id), nil, nil)
}

// SetDeploymentActivation activates or deactivates a process deployment.
func (c *Client) SetDeploymentActivation(ctx context.Context, id int64, active bool) error {
	return c.doRequest(ctx, http.MethodPut,
		fmt.Sprintf("/api/v1/process-deployments/%d/activation", id),
		DeploymentActivation{Active: active}, nil)
}

// --- Environments ---

// ListEnvironments returns all environments.
func (c *Client) ListEnvironments(ctx context.Context) ([]EnvironmentGet, error) {
	var result EnvironmentListResponse
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/environments", nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetEnvironment returns a single environment by ID.
func (c *Client) GetEnvironment(ctx context.Context, id int64) (*EnvironmentGet, error) {
	var result EnvironmentGet
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/environments/%d", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Agent Groups ---

// GetAgentGroup retrieves an agent group by ID.
func (c *Client) GetAgentGroup(ctx context.Context, id int64) (*AgentGroupGet, error) {
	var result AgentGroupGet
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/agent-groups/%d", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListAgentGroupsByEnvironment returns all agent groups for an environment.
func (c *Client) ListAgentGroupsByEnvironment(ctx context.Context, envID int64) ([]AgentGroupGet, error) {
	var result AgentGroupListResponse
	err := c.doRequest(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/environments/%d/agent-groups", envID), nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// --- Environment Variables ---

// CreateEnvironmentVariable creates a new root-level environment variable schema.
func (c *Client) CreateEnvironmentVariable(ctx context.Context, req EnvironmentVariableSchemaCreate) (*EnvironmentVariableSchemaGet, error) {
	var result EnvironmentVariableSchemaGet
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/environment-variables", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEnvironmentVariable retrieves an environment variable schema by ID.
func (c *Client) GetEnvironmentVariable(ctx context.Context, id int64) (*EnvironmentVariableSchemaGet, error) {
	var result EnvironmentVariableSchemaGet
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/environment-variables/%d", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SetEnvironmentVariableValue sets the value for a specific environment.
func (c *Client) SetEnvironmentVariableValue(ctx context.Context, schemaID, envID int64, value interface{}) error {
	return c.doRequest(ctx, http.MethodPut,
		fmt.Sprintf("/api/v1/environment-variables/%d/values/%d", schemaID, envID),
		EnvironmentVariableValueSet{Value: value}, nil)
}

// DeleteEnvironmentVariable deletes an environment variable schema.
func (c *Client) DeleteEnvironmentVariable(ctx context.Context, schemaID int64) error {
	return c.doRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/environment-variables/%d", schemaID), nil, nil)
}

// --- API Keys ---

// CreateApiKey creates a new API key.
func (c *Client) CreateApiKey(ctx context.Context, req ApiKeyCreate) (*ApiKeyGet, error) {
	var result ApiKeyGet
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/api-management/access/api-keys", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetApiKey retrieves an API key by ID.
func (c *Client) GetApiKey(ctx context.Context, id int64) (*ApiKeyGet, error) {
	var result ApiKeyGet
	err := c.doRequest(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/api-management/access/api-keys/%d", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateApiKey updates an API key.
func (c *Client) UpdateApiKey(ctx context.Context, id int64, req ApiKeyUpdate) (*ApiKeyGet, error) {
	var result ApiKeyGet
	err := c.doRequest(ctx, http.MethodPut,
		fmt.Sprintf("/api/v1/api-management/access/api-keys/%d", id), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteApiKey deletes an API key.
func (c *Client) DeleteApiKey(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/api-management/access/api-keys/%d", id), nil, nil)
}

// --- API Policies ---

// CreateApiPolicy creates a new API policy.
func (c *Client) CreateApiPolicy(ctx context.Context, req ApiPolicySave) (*ApiPolicyGet, error) {
	var result ApiPolicyGet
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/api-policies", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetApiPolicy retrieves an API policy by ID.
func (c *Client) GetApiPolicy(ctx context.Context, id int64) (*ApiPolicyGet, error) {
	var result ApiPolicyGet
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/api-policies/%d", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateApiPolicy updates an API policy.
func (c *Client) UpdateApiPolicy(ctx context.Context, id int64, req ApiPolicySave) (*ApiPolicyGet, error) {
	var result ApiPolicyGet
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/api-policies/%d", id), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteApiPolicy deletes an API policy.
func (c *Client) DeleteApiPolicy(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/api-policies/%d", id), nil, nil)
}

// --- API Specifications ---

// CreateApiSpecification creates a new API specification.
func (c *Client) CreateApiSpecification(ctx context.Context, req ApiSpecificationCreate) (*ApiSpecificationGet, error) {
	var result ApiSpecificationGet
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/api-management/api-specifications", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetApiSpecification retrieves an API specification by ID.
func (c *Client) GetApiSpecification(ctx context.Context, id int64) (*ApiSpecificationGet, error) {
	var result ApiSpecificationGet
	err := c.doRequest(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/api-management/api-specifications/%d", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateApiSpecification updates an API specification.
func (c *Client) UpdateApiSpecification(ctx context.Context, id int64, req ApiSpecificationUpdate) (*ApiSpecificationGet, error) {
	var result ApiSpecificationGet
	err := c.doRequest(ctx, http.MethodPut,
		fmt.Sprintf("/api/v1/api-management/api-specifications/%d", id), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteApiSpecification deletes an API specification.
func (c *Client) DeleteApiSpecification(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/api-management/api-specifications/%d", id), nil, nil)
}

// --- Private Applications ---

// CreatePrivateApplication creates a new private application.
func (c *Client) CreatePrivateApplication(ctx context.Context, req PrivateApplicationCreate) (*PrivateApplicationGet, error) {
	var result PrivateApplicationGet
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/private-application", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPrivateApplication retrieves a private application by ID.
func (c *Client) GetPrivateApplication(ctx context.Context, id int64) (*PrivateApplicationGet, error) {
	var result PrivateApplicationGet
	err := c.doRequest(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/private-application/%d", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdatePrivateApplication updates a private application.
func (c *Client) UpdatePrivateApplication(ctx context.Context, id int64, req PrivateApplicationUpdate) (*PrivateApplicationGet, error) {
	var result PrivateApplicationGet
	err := c.doRequest(ctx, http.MethodPut,
		fmt.Sprintf("/api/v1/private-application/%d", id), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeletePrivateApplication deletes a private application.
func (c *Client) DeletePrivateApplication(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/private-application/%d", id), nil, nil)
}

// --- Process Templates ---

// CreateProcessTemplate creates a new process template.
func (c *Client) CreateProcessTemplate(ctx context.Context, req ProcessTemplateCreate) (*ProcessTemplateGet, error) {
	var result ProcessTemplateGet
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/process-templates", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetProcessTemplate retrieves a process template by ID (GUID).
func (c *Client) GetProcessTemplate(ctx context.Context, id string) (*ProcessTemplateGet, error) {
	var result ProcessTemplateGet
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/process-templates/%s", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateProcessTemplate updates an existing process template.
func (c *Client) UpdateProcessTemplate(ctx context.Context, req ProcessTemplateUpdate) error {
	return c.doRequest(ctx, http.MethodPut, "/api/v1/process-templates", req, nil)
}

// DeleteProcessTemplate deletes a process template by ID.
func (c *Client) DeleteProcessTemplate(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/process-templates/%s", id), nil, nil)
}

// --- Processes ---

// ListProcesses returns all processes (paged).
func (c *Client) ListProcesses(ctx context.Context) ([]ProcessGet, error) {
	var result ProcessListResponse
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/processes", nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetProcess retrieves a process by ID.
func (c *Client) GetProcess(ctx context.Context, id int64) (*ProcessGet, error) {
	var result ProcessGet
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/processes/%d", id), nil, &result)
	if IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ImportProcess uploads a .frends package file to POST /api/v1/processes/batch-import.
// importConflict controls behaviour when the process GUID already exists:
// "Error", "UseExisting", "NewVersion", "NewActiveElement", "NewInactiveElement".
// Import can be slow; a dedicated http.Client with a 10-minute timeout is used.
func (c *Client) ImportProcess(ctx context.Context, filePath, importConflict string) (*ProcessImportResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening package file %q: %w", filePath, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("creating multipart file field: %w", err)
	}
	if _, err := io.Copy(fw, f); err != nil {
		return nil, fmt.Errorf("copying package file into request: %w", err)
	}
	if err := w.WriteField("importConflict", importConflict); err != nil {
		return nil, fmt.Errorf("writing importConflict field: %w", err)
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/api/v1/processes/batch-import", &buf)
	if err != nil {
		return nil, fmt.Errorf("creating import request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", w.FormDataContentType())

	longClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := longClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing import request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading import response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d importing process: %s", resp.StatusCode, string(body))
	}

	var result ProcessImportResultResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding import response: %w", err)
	}
	return &result.Data, nil
}

// ExportProcess downloads the binary package for a process version by numeric ID.
// Returns nil, nil when the process is not found.
func (c *Client) ExportProcess(ctx context.Context, id int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/v1/processes/%d/export", c.BaseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("creating export request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing export request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d exporting process %d: %s", resp.StatusCode, id, string(body))
	}
	return io.ReadAll(resp.Body)
}

// DeleteProcess deletes a process by GUID and automatically undeploys it from all agent groups.
func (c *Client) DeleteProcess(ctx context.Context, guid string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/processes/%s", guid), nil, nil)
}
