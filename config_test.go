package gotel_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/iamBelugax/gotel"
)

var _ = Describe("Config", func() {
	var (
		serviceVersion = "1.0.0"
		environment    = "development"
		serviceName    = "test-service"
		samplingRate   = 1.0
		endpoint       = "localhost:4317"
		batchTimeout   = time.Second * 5
		exportTimeout  = time.Second * 30
	)

	Context("DefaultConfig with no additional options", func() {
		It("should return default configuration values", func() {
			config := gotel.DefaultConfig(gotel.WithServiceInfo(serviceName, serviceVersion, environment))

			Expect(config).NotTo(BeNil())
			Expect(config.Service.Name).To(Equal(serviceName))
			Expect(config.Service.Version).To(Equal(serviceVersion))
			Expect(config.Service.Environment).To(Equal(environment))

			Expect(config.Exporter.Endpoint).To(Equal(endpoint))
			Expect(config.Exporter.BatchTimeout).To(Equal(batchTimeout))
			Expect(config.Exporter.ExportTimeout).To(Equal(exportTimeout))
			Expect(config.Exporter.Headers).To(BeEmpty())

			Expect(config.Debug).To(BeFalse())
			Expect(config.ResourceAttrs).To(BeEmpty())
			Expect(config.Tracing.SamplingRatio).To(Equal(samplingRate))
		})
	})

	Context("DefaultConfig with updated service info", func() {
		It("should correctly override service information", func() {
			newServiceVersion := "2.0.0"
			newEnvironment := "production"
			newServiceName := "updated-service-name"

			config := gotel.DefaultConfig(gotel.WithServiceInfo(newServiceName, newServiceVersion, newEnvironment))

			Expect(config.Service.Name).To(Equal(newServiceName))
			Expect(config.Service.Version).To(Equal(newServiceVersion))
			Expect(config.Service.Environment).To(Equal(newEnvironment))

			Expect(config.Exporter.Endpoint).To(Equal(endpoint))
			Expect(config.Exporter.BatchTimeout).To(Equal(batchTimeout))
			Expect(config.Exporter.ExportTimeout).To(Equal(exportTimeout))
		})
	})

	Context("DefaultConfig with updated exporter config", func() {
		It("should correctly apply exporter configuration options", func() {
			customEndpoint := "custom-endpoint:4318"
			customBatchTimeout := time.Second * 10
			customExportTimeout := time.Second * 60
			headers := map[string]string{
				"x-api-key":     "test-api-key",
				"Authorization": "Bearer token",
			}

			config := gotel.DefaultConfig(
				gotel.WithServiceInfo(serviceName, serviceVersion, environment),
				gotel.WithEndpoint(customEndpoint),
				gotel.WithBatchTimeout(customBatchTimeout),
				gotel.WithExportTimeout(customExportTimeout),
				gotel.WithHeader("x-api-key", "test-api-key"),
				gotel.WithHeaders(map[string]string{"Authorization": "Bearer token"}),
			)

			Expect(config.Exporter.Endpoint).To(Equal(customEndpoint))
			Expect(config.Exporter.BatchTimeout).To(Equal(customBatchTimeout))
			Expect(config.Exporter.ExportTimeout).To(Equal(customExportTimeout))
			Expect(config.Exporter.Headers).To(Equal(headers))
		})
	})

	Context("Configuration option combinations", func() {
		It("should handle multiple configuration options correctly", func() {
			config := gotel.DefaultConfig(
				gotel.WithServiceInfo("combination-service", "3.0.0", "staging"),
				gotel.WithSamplingRatio(0.5),
				gotel.WithDebug(true),
				gotel.WithInsecure(false),
				gotel.WithLogLevel("info"),
				gotel.WithTLSCredentials(insecure.NewCredentials()),
				gotel.WithResourceAttr("custom.key", "custom.value"),
				gotel.WithResourceAttrs(map[string]any{"platform": "kubernetes"}),
			)

			Expect(config.Service.Name).To(Equal("combination-service"))
			Expect(config.Service.Version).To(Equal("3.0.0"))
			Expect(config.Service.Environment).To(Equal("staging"))
			Expect(config.Tracing.SamplingRatio).To(Equal(0.5))
			Expect(config.Debug).To(BeTrue())
			Expect(config.ResourceAttrs["custom.key"]).To(Equal("custom.value"))
			Expect(config.Security.Insecure).To(BeFalse())
			Expect(config.Logging.Level).To(Equal("info"))
		})

		It("should clamp sampling ratio to valid range", func() {
			configHigh := gotel.DefaultConfig(gotel.WithSamplingRatio(1.5))
			Expect(configHigh.Tracing.SamplingRatio).To(Equal(1.0))

			configLow := gotel.DefaultConfig(gotel.WithSamplingRatio(-0.1))
			Expect(configLow.Tracing.SamplingRatio).To(Equal(0.0))
		})
	})
})
