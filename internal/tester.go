package internal

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/jacobweinstock/registrar"
	"github.com/rs/zerolog"

	"github.com/bmc-toolbox/bmclib/v2"
	"github.com/bmc-toolbox/bmclib/v2/providers"
)

const (
	loginTimeout  = time.Minute * 1
	logoutTimeout = time.Minute * 1
)

// Tester runs tests on a host, this struct holds config attributes for tester
type Tester struct {
	bmcHost          string
	bmcUser          string
	bmcPass          string
	bmcPort          string
	tests            []*test
	results          []Result
	disableFiltering bool
	logger           logr.Logger
}

// TestFunc should return output if any and an error to indicate test failure.
type testFunc func(context.Context, *bmclib.Client) (string, error)

// test holds attributes for a test executed by Tester
type test struct {
	Feature  string
	TestFunc testFunc
}

// NewTester returns a Tester instance with the parameters configured.
func NewTester(bmcHost, bmcUser, bmcPass, bmcPort string, disableFilter bool, logLevel string) *Tester {
	return &Tester{
		bmcHost:          bmcHost,
		bmcUser:          bmcUser,
		bmcPass:          bmcPass,
		bmcPort:          bmcPort,
		disableFiltering: disableFilter,
		logger:           NewLogger(logLevel),
	}
}

func (t *Tester) registry() map[registrar.Feature]testFunc {
	return map[registrar.Feature]testFunc{
		providers.FeaturePowerState:    t.powerState,
		providers.FeaturePowerSet:      t.powerSet,
		providers.FeatureBootDeviceSet: t.bootDeviceSet,
		providers.FeatureBmcReset:      t.bmcReset,
		providers.FeatureUserRead:      t.userRead,
	}
}

func (t *Tester) initTester(ctx context.Context, configTests *ConfigTests) error {
	// init tests to run
	runTests := make([]*test, 0, len(configTests.Features))

	registry := t.registry()

	for _, feature := range configTests.Features {
		testFunc, exists := registry[registrar.Feature(feature)]
		if !exists {
			return errors.New("unknown bmclib feature defined in test: " + feature)
		}

		runTests = append(runTests, &test{
			Feature:  feature,
			TestFunc: testFunc,
		})
	}

	t.tests = runTests

	return nil
}

// Run runs all the given tests.
func (t *Tester) Run(ctx context.Context, tests *ConfigTests) {
	if err := t.initTester(ctx, tests); err != nil {
		t.logger.Error(err, "tester init error")
	}

	client := t.newBmclibClient(tests.Provider, tests.Protocol, t.logger)

	if err := client.Open(ctx); err != nil {
		for _, test := range t.tests {
			t.results = append(t.results, Result{
				Feature:            string(test.Feature),
				Protocol:           tests.Protocol,
				Error:              err.Error(),
				ProvidersAttempted: client.GetMetadata().ProvidersAttempted,
			})
		}

		return
	}

	ctxClose, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	defer client.Close(ctxClose)

	for _, test := range t.tests {
		t.logger.V(1).Info("running test", "feature", test.Feature)

		startTime := time.Now()

		result := Result{
			Feature:  string(test.Feature),
			Protocol: tests.Protocol,
		}

		output, err := test.TestFunc(ctx, client)
		if err != nil {
			result.Error = err.Error()

			t.logger.V(1).Info("Test failed: ", test.Feature)
		} else {
			result.Succeeded = true

			t.logger.V(1).Info("Test successful: ", test.Feature)
		}

		result.SuccessfulProvider = client.GetMetadata().SuccessfulProvider
		result.Output = output
		result.Runtime = time.Duration(time.Since(startTime)).String()
		t.results = append(t.results, result)
	}
}

// TODO: get bmclib to return these
func SupportedProviders() []string {
	return []string{
		"ipmitool",
		"gofish",
	}
}

func SupportedProtocols() []string {
	return []string{
		"ipmi",
		"redfish",
	}
}

func (t *Tester) newBmclibClient(provider, protocol string, logger logr.Logger) *bmclib.Client {
	opts := []bmclib.Option{
		bmclib.WithLogger(logger),
		bmclib.WithPerProviderTimeout(loginTimeout),
	}

	// init client
	client := bmclib.NewClient(t.bmcHost, t.bmcPort, t.bmcUser, t.bmcPass, opts...)

	if !t.disableFiltering {
		drivers := registrar.Drivers{}
		drivers = append(drivers, client.Registry.Using(protocol)...)
		client.Registry.Drivers = drivers
	}

	return client
}

func (t *Tester) Results() []Result {
	return t.results
}

// DeviceResult holds the test results for a given device
type DeviceResult struct {
	Vendor  string
	Model   string
	Name    string
	BMCIP   string
	Results []Result
}

// Result is a single test result
type Result struct {
	Feature            string
	Protocol           string
	ProvidersAttempted []string
	SuccessfulProvider string
	Output             string
	Error              string
	Succeeded          bool
	Runtime            string
}

// ResultStore stores test results
type ResultStore struct {
	mu      *sync.RWMutex
	results []DeviceResult
}

func NewTestResultStore() *ResultStore {
	return &ResultStore{mu: &sync.RWMutex{}, results: []DeviceResult{}}
}

func (r *ResultStore) Save(result DeviceResult) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.results = append(r.results, result)
}

func (r *ResultStore) Read() []DeviceResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.results
}

// NewLogger returns a logr.
func NewLogger(level string) logr.Logger {
	logger := zerolog.New(os.Stdout)

	logger = logger.With().Caller().Timestamp().Logger()

	var l zerolog.Level
	switch level {
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	logger = logger.Level(l)

	return zerologr.New(&logger)
}
