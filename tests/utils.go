package tests

import "os"

// PromGatewayHost 获取Prometheus网关地址
func PromGatewayHost() string {
	h := os.Getenv("PROM_GATEWAY_HOST")
	if h == "" {
		h = "http://127.0.0.1:9091"
	}
	return h
}

// PromQueryHost 获取Prometheus查询地址
func PromQueryHost() string {
	h := os.Getenv("PROM_QUERY_HOST")
	if h == "" {
		h = "http://127.0.0.1:9091"
	}
	return h
}
