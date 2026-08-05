package handlers

import "github.com/sahil/leasewebassignment/internal/model"

type ServerResponse struct {
	Model    string  `json:"model"`
	RAM      string  `json:"ram"`
	HDD      string  `json:"hdd"`
	Location string  `json:"location"`
	Price    float64 `json:"price"`
}

type ServersResponse struct {
	Data []ServerResponse `json:"data"`
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

func mapServersResponse(servers []model.Server) ServersResponse {
	responses := make([]ServerResponse, 0, len(servers))
	for _, server := range servers {
		responses = append(responses, mapServerResponse(server))
	}
	return ServersResponse{Data: responses}
}
