package response

type Endpoint struct {
	ID           string `json:"id"`
	HttpMethod   string `gorm:"size:10;not null"`
	HttpEndpoint string `gorm:"not null"`
}

type Endpoints []Endpoint

type EndpointsResponse struct {
	Endpoints
}
