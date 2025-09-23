package gotel_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/iamBelugax/gotel"
)

var _ = Describe("Config", func() {
	var (
		serviceVersion = "1.0.0"
		environment    = "development"
		serviceName    = "test-service"

		samplingRate = 1.0

		endpoint      = "localhost:4317"
		batchTimeout  = time.Second * 5
		exportTimeout = time.Second * 30
	)

	Context("DefaultConfig", func() {
		It("should return default config", func() {
			config := gotel.DefaultConfig(serviceName)

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
})
