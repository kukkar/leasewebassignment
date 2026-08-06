package handlers

import "github.com/sahil/leasewebassignment/internal/model"

type ServerResponse struct {
	Model    string  `json:"model"`
	RAM      string  `json:"ram"`
	HDD      string  `json:"hdd"`
	Location string  `json:"location"`
	Price    float64 `json:"price"`
}

type PaginationMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ServersResponse struct {
	Data []ServerResponse `json:"data"`
	Meta PaginationMeta   `json:"meta"`
}

func mapServerResponse(server model.Server) ServerResponse {
	return ServerResponse{
		Model:    server.Model,
		RAM:      server.RAM,
		HDD:      server.HDD,
		Location: server.Location,
		Price:    server.Price,
	}
}

// mapServersResponse pages the already-filtered result set: the full match
// count is reported in Meta.Total even though only one page of Data is
// returned, so a client can tell "8 total, 5 shown" apart from "5 total".
func mapServersResponse(servers []model.Server, limit, offset int) ServersResponse {
	total := len(servers)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	page := servers[start:end]

	responses := make([]ServerResponse, 0, len(page))
	for _, server := range page {
		responses = append(responses, mapServerResponse(server))
	}
	return ServersResponse{
		Data: responses,
		Meta: PaginationMeta{Total: total, Limit: limit, Offset: offset},
	}
}
