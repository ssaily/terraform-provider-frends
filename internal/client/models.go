package client

// ProcessDeploymentCreate is the request body for POST /api/v1/process-deployments.
type ProcessDeploymentCreate struct {
	AgentGroupID          int64                  `json:"agentGroupId"`
	Processes             []ProcessVersionInput  `json:"processes"`
	ActivateTriggers      bool                   `json:"activateTriggers"`
	DeploymentDescription string                 `json:"deploymentDescription,omitempty"`
}

// ProcessVersionInput represents a process version in a deployment request.
type ProcessVersionInput struct {
	ProcessGUID      string            `json:"processGuid"`
	Version          int32             `json:"version"`
	ProcessVariables []ProcessVariable `json:"processVariables,omitempty"`
}

// ProcessVariable is a key/value pair for a process deployment variable.
type ProcessVariable struct {
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	IsSecret    bool   `json:"isSecret"`
	Mode        string `json:"mode,omitempty"`
	Description string `json:"description,omitempty"`
}

// ProcessDeploymentGet is the response body from GET /api/v1/process-deployments/{id}.
type ProcessDeploymentGet struct {
	DeploymentID    int64            `json:"deploymentId"`
	ProcessID       int64            `json:"processId"`
	ProcessName     string           `json:"processName"`
	ProcessGUID     string           `json:"processGuid"`
	ProcessVersion  string           `json:"processVersion"`
	Description     string           `json:"description"`
	TriggersActive  bool             `json:"triggersActive"`
	AgentGroup      AgentGroupBase   `json:"agentGroup"`
	DeployedAtUtc   string           `json:"deployedAtUtc"`
	DeployedBy      string           `json:"deployedBy"`
}

// ProcessDeploymentListResponse is the response from GET /api/v1/process-deployments.
type ProcessDeploymentListResponse struct {
	Data []ProcessDeploymentGet `json:"data"`
}

// DeploymentActivation is the request body for PUT /api/v1/process-deployments/{id}/activation.
type DeploymentActivation struct {
	Active bool `json:"active"`
}

// AgentGroupBase represents a minimal agent group reference.
type AgentGroupBase struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"displayName"`
}

// AgentGroupGet is a full agent group with agents and environment.
type AgentGroupGet struct {
	ID              int64           `json:"id"`
	DisplayName     string          `json:"displayName"`
	InternalName    string          `json:"internalName"`
	IsCrossPlatform bool            `json:"isCrossPlatform"`
	Agents          []string        `json:"agents"`
	Environment     EnvironmentBase `json:"environment"`
}

// AgentGroupListResponse wraps a list of agent groups.
type AgentGroupListResponse struct {
	Data []AgentGroupGet `json:"data"`
}

// EnvironmentBase is a minimal environment reference.
type EnvironmentBase struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"displayName"`
}

// EnvironmentGet is a full environment with agent groups.
type EnvironmentGet struct {
	ID           int64           `json:"id"`
	DisplayName  string          `json:"displayName"`
	InternalName string          `json:"internalName"`
	AgentGroups  []AgentGroupGet `json:"agentGroups"`
}

// EnvironmentListResponse wraps a list of environments.
type EnvironmentListResponse struct {
	Data []EnvironmentGet `json:"data"`
}

// EnvironmentVariableSchemaCreate is the request to create a root env var schema.
type EnvironmentVariableSchemaCreate struct {
	Name string `json:"name"`
}

// EnvironmentVariableSchemaGet is the response from GET /api/v1/environment-variables/{id}.
type EnvironmentVariableSchemaGet struct {
	ID          int64                            `json:"id"`
	Name        string                           `json:"name"`
	Type        string                           `json:"type"`
	Description interface{}                      `json:"description"`
	Values      []EnvironmentVariableValue       `json:"values"`
}

// EnvironmentVariableValue represents a value for a specific environment.
type EnvironmentVariableValue struct {
	Value       interface{}     `json:"value"`
	ModifiedUtc string          `json:"modifiedUtc"`
	Modifier    string          `json:"modifier"`
	Version     int32           `json:"version"`
	Environment EnvironmentBase `json:"environment"`
}

// EnvironmentVariableValueSet is the request body for PUT /api/v1/environment-variables/{schemaId}/values/{environmentId}.
type EnvironmentVariableValueSet struct {
	Value interface{} `json:"value"`
}

// EnvironmentVariableDescriptionUpdate is for PATCH /api/v1/environment-variables/{schemaId}.
type EnvironmentVariableDescriptionUpdate struct {
	Description string `json:"description"`
}

// EnvironmentVariableSchemaPagedResponse wraps a paged list of env var schemas.
type EnvironmentVariableSchemaPagedResponse struct {
	Data []EnvironmentVariableSchemaGet `json:"data"`
}

// ApiKeyCreate is the request body for POST /api/v1/api-management/access/api-keys.
type ApiKeyCreate struct {
	Name          string `json:"name,omitempty"`
	EnvironmentID int64  `json:"environmentId"`
}

// ApiKeyUpdate is the request body for PUT /api/v1/api-management/access/api-keys/{id}.
type ApiKeyUpdate struct {
	Name string `json:"name,omitempty"`
}

// ApiKeyGet is the response from GET /api/v1/api-management/access/api-keys/{id}.
type ApiKeyGet struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Value       string          `json:"value"`
	Modified    string          `json:"modified"`
	Modifier    string          `json:"modifier"`
	Environment EnvironmentBase `json:"environment"`
}

// ApiKeyListResponse wraps a list of API keys.
type ApiKeyListResponse struct {
	Data []ApiKeyGet `json:"data"`
}

// ApiPolicySave is the request body for POST/PUT /api/v1/api-policies.
type ApiPolicySave struct {
	Name                  string                          `json:"name"`
	Description           string                          `json:"description,omitempty"`
	Tags                  []string                        `json:"tags,omitempty"`
	AllowPublicAccess     bool                            `json:"allowPublicAccess"`
	ApiKeyName            string                          `json:"apiKeyName,omitempty"`
	ApiKeyLocation        string                          `json:"apiKeyLocation,omitempty"`
	TargetEndpoints       []ApiPolicyTargetEndpointSave   `json:"targetEndpoints"`
	Identities            []ApiPolicyIdentitySave         `json:"identities,omitempty"`
	ApiKeyGroups          []ApiPolicyApiKeyGroupSave      `json:"apiKeyGroups,omitempty"`
	RequestLoggingOptions []ApiPolicyRequestLoggingSave   `json:"requestLoggingOptions,omitempty"`
}

// ApiPolicyTargetEndpointSave defines an API endpoint target.
type ApiPolicyTargetEndpointSave struct {
	URL    string `json:"url"`
	Method string `json:"method,omitempty"`
}

// ApiPolicyIdentitySave defines an identity for an API policy.
type ApiPolicyIdentitySave struct {
	Name  string                      `json:"name"`
	Rules []ApiPolicyIdentityRuleSave `json:"rules,omitempty"`
}

// ApiPolicyIdentityRuleSave defines a rule for an identity.
type ApiPolicyIdentityRuleSave struct {
	ClaimType  string `json:"claimType"`
	ClaimValue string `json:"claimValue"`
	MatchType  string `json:"matchType"`
}

// ApiPolicyApiKeyGroupSave defines an API key group for an API policy.
type ApiPolicyApiKeyGroupSave struct {
	Name    string                              `json:"name"`
	ApiKeys []ApiPolicyApiKeySave               `json:"apiKeys,omitempty"`
}

// ApiPolicyApiKeySave defines an API key within a group.
type ApiPolicyApiKeySave struct {
	Value string `json:"value"`
}

// ApiPolicyRequestLoggingSave defines request logging options.
type ApiPolicyRequestLoggingSave struct {
	DeploymentID int64  `json:"deploymentId"`
	IpLogging    string `json:"ipLogging,omitempty"`
}

// ApiPolicyGet is the response from GET /api/v1/api-policies/{id}.
type ApiPolicyGet struct {
	ID                    int64                          `json:"id"`
	Name                  string                         `json:"name"`
	Description           string                         `json:"description"`
	Tags                  []string                       `json:"tags"`
	AllowPublicAccess     bool                           `json:"allowPublicAccess"`
	ApiKeyName            string                         `json:"apiKeyName"`
	ApiKeyLocation        string                         `json:"apiKeyLocation"`
	TargetEndpoints       []ApiPolicyTargetEndpointView  `json:"targetEndpoints"`
	Identities            []ApiPolicyIdentityView        `json:"identities"`
}

// ApiPolicyTargetEndpointView is a read response for a target endpoint.
type ApiPolicyTargetEndpointView struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}

// ApiPolicyIdentityView is a read response for an identity.
type ApiPolicyIdentityView struct {
	Name string `json:"name"`
}

// ApiPolicyListResponse wraps a list of API policies.
type ApiPolicyListResponse struct {
	Data []ApiPolicyGet `json:"data"`
}

// ApiSpecificationCreate is the request to create an API specification.
type ApiSpecificationCreate struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ApiSpecificationUpdate is the request to update an API specification.
type ApiSpecificationUpdate struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ApiSpecificationGet is the response from GET /api/v1/api-management/api-specifications/{id}.
type ApiSpecificationGet struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	ActiveVersion int32  `json:"activeVersion"`
	Description   string `json:"description"`
}

// ApiSpecificationPagedResponse wraps paged API specifications.
type ApiSpecificationPagedResponse struct {
	Data []ApiSpecificationGet `json:"data"`
}

// PrivateApplicationCreate is the request to create a private application.
type PrivateApplicationCreate struct {
	Name                    string                 `json:"name"`
	Description             string                 `json:"description,omitempty"`
	DefaultTokenLifetimeDays int32                 `json:"defaultTokenLifetimeDays"`
	CustomTokenClaims        map[string]interface{} `json:"customTokenClaims"`
	Tags                    []string               `json:"tags,omitempty"`
}

// PrivateApplicationUpdate is the request to update a private application.
type PrivateApplicationUpdate struct {
	Name                    string                 `json:"name"`
	Description             string                 `json:"description,omitempty"`
	DefaultTokenLifetimeDays int32                 `json:"defaultTokenLifetimeDays"`
	CustomTokenClaims        map[string]interface{} `json:"customTokenClaims"`
	Tags                    []string               `json:"tags,omitempty"`
}

// PrivateApplicationGet is the response from GET /api/v1/private-application/{id}.
type PrivateApplicationGet struct {
	ID                       int64                  `json:"id"`
	Name                     string                 `json:"name"`
	Description              string                 `json:"description"`
	DefaultTokenLifetimeDays int32                  `json:"defaultTokenLifetimeDays"`
	CustomTokenClaims        map[string]interface{} `json:"customTokenClaims"`
	Tags                     []string               `json:"tags"`
	Modifier                 string                 `json:"modifier"`
	ModifiedUtc              string                 `json:"modifiedUtc"`
	HasTokens                bool                   `json:"hasTokens"`
	HasActiveTokens          bool                   `json:"hasActiveTokens"`
}

// PrivateApplicationListResponse wraps a list of private applications.
type PrivateApplicationListResponse struct {
	Data []PrivateApplicationGet `json:"data"`
}

// ProcessTemplateCreate is the request body for creating a process template.
type ProcessTemplateCreate struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ProcessTemplateUpdate is the request body for updating a process template.
type ProcessTemplateUpdate struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ProcessTemplateGet is the response from GET /api/v1/process-templates/{id}.
type ProcessTemplateGet struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     int32  `json:"version"`
}

// ProcessTemplateListResponse wraps paged process templates.
type ProcessTemplateListResponse struct {
	Data []ProcessTemplateGet `json:"data"`
}

// ProcessGet is the response from GET /api/v1/processes/{id}.
type ProcessGet struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	UniqueIdentifier string `json:"uniqueIdentifier"`
	Version          int32  `json:"version"`
	Description      string `json:"description"`
	IsDeleted        bool   `json:"isDeleted"`
	IsDraft          bool   `json:"isDraft"`
}

// ProcessListResponse wraps paged process list.
type ProcessListResponse struct {
	Data []ProcessGet `json:"data"`
}

// ProcessImportResult is the data payload returned by POST /api/v1/processes/batch-import.
type ProcessImportResult struct {
	Name              string `json:"name"`
	ElementIdentifier string `json:"elementIdentifier"` // GUID, stable across versions
	ID                int64  `json:"id"`                // numeric version ID
	ResourceLocation  string `json:"resourceLocation"`
	Version           int32  `json:"version"`
}

// ProcessImportResultResponse wraps the import result envelope.
type ProcessImportResultResponse struct {
	Data ProcessImportResult `json:"data"`
}
