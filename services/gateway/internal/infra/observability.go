package infra

type ObservabilityConfig struct {
	MetricsEnabled  bool
	TracingEnabled  bool
	TracingEndpoint string
	TracingInsecure bool
}
