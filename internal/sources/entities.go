package sources

import "time"

type ApplicationTypeIDs struct {
	Data  []ApplicationTypes `json:"data"`
	Meta  Meta               `json:"meta"`
	Links Links              `json:"links"`
}

type Meta struct {
	Count  int `json:"count"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type Links struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

type ApplicationTypes struct {
	ID                           string        `json:"id"`
	CreatedAt                    time.Time     `json:"created_at"`
	UpdatedAt                    time.Time     `json:"updated_at"`
	Name                         string        `json:"name"`
	DisplayName                  string        `json:"display_name"`
	DependentApplications        []interface{} `json:"dependent_applications"`
	SupportedSourceTypes         []string      `json:"supported_source_types"`
	SupportedAuthenticationTypes struct {
		Quay      []string `json:"quay"`
		Github    []string `json:"github"`
		Gitlab    []string `json:"gitlab"`
		Bitbucket []string `json:"bitbucket"`
		Dockerhub []string `json:"dockerhub"`
	} `json:"supported_authentication_types,omitempty"`
	SupportedAuthenticationTypes0 struct {
		Azure  []string `json:"azure"`
		Amazon []string `json:"amazon"`
	} `json:"supported_authentication_types,omitempty"`
	SupportedAuthenticationTypes1 struct {
		Ibm                       []string `json:"ibm"`
		Azure                     []string `json:"azure"`
		Amazon                    []string `json:"amazon"`
		Google                    []string `json:"google"`
		Openshift                 []string `json:"openshift"`
		OracleCloudInfrastructure []string `json:"oracle-cloud-infrastructure"`
	} `json:"supported_authentication_types,omitempty"`
	SupportedAuthenticationTypes2 struct {
		Satellite []string `json:"satellite"`
	} `json:"supported_authentication_types,omitempty"`
	SupportedAuthenticationTypes3 struct {
		Azure  []string `json:"azure"`
		Amazon []string `json:"amazon"`
		Google []string `json:"google"`
	} `json:"supported_authentication_types,omitempty"`
}

type Application struct {
	ID                      string `json:"id"`
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
	AvailabilityStatus      string `json:"availability_status"`
	LastCheckedAt           string `json:"last_checked_at"`
	AvailabilityStatusError string `json:"availability_status_error"`
	SourceID                string `json:"source_id"`
	ApplicationTypeID       string `json:"application_type_id"`
}

type Authentication struct {
	ID                 string `json:"id"`
	Authtype           string `json:"authtype"`
	Username           string `json:"username"`
	AvailabilityStatus string `json:"availability_status"`
	ResourceType       string `json:"resource_type"`
	ResourceID         string `json:"resource_id"`
}
