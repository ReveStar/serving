/*
Copyright 2018 The Knative Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"

	. "knative.dev/pkg/configmap/testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]string
		want    *Config
		wantErr bool
	}{{
		name:  "default",
		input: map[string]string{},
		want:  defaultConfig(),
	}, {
		name: "minimum",
		input: map[string]string{
			"max-scale-up-rate":                       "1.001",
			"container-concurrency-target-percentage": "0.5",
			"container-concurrency-target-default":    "10.0",
			"target-burst-capacity":                   "0",
			"stable-window":                           "5m",
			"tick-interval":                           "2s",
			"panic-window-percentage":                 "10",
			"panic-threshold-percentage":              "200",
			"activator-capacity":                      "1",
		},
		want: func() *Config {
			c := defaultConfig()
			c.ContainerConcurrencyTargetFraction = 0.5
			c.ContainerConcurrencyTargetDefault = 10
			c.MaxScaleUpRate = 1.001
			c.TargetBurstCapacity = 0
			c.StableWindow = 5 * time.Minute
			c.ActivatorCapacity = 1
			return c
		}(),
	}, {
		name: "concurrencty target percentage as percent",
		input: map[string]string{
			"container-concurrency-target-percentage": "55",
		},
		want: func() *Config {
			c := defaultConfig()
			c.ContainerConcurrencyTargetFraction = 0.55
			return c
		}(),
	}, {
		name: "with -1 tbc",
		input: map[string]string{
			"target-burst-capacity": "-1",
		},
		want: func() *Config {
			c := defaultConfig()
			c.TargetBurstCapacity = -1
			return c
		}(),
	}, {
		name: "with default toggles set",
		input: map[string]string{
			"enable-scale-to-zero":                    "true",
			"enable-graceful-scaledown":               "false",
			"max-scale-down-rate":                     "3.0",
			"max-scale-up-rate":                       "1.01",
			"container-concurrency-target-percentage": "0.71",
			"container-concurrency-target-default":    "10.5",
			"requests-per-second-target-default":      "10.11",
			"target-burst-capacity":                   "12345",
			"stable-window":                           "5m",
			"tick-interval":                           "2s",
			"panic-window-percentage":                 "10",
			"panic-threshold-percentage":              "200",
			"pod-autoscaler-class":                    "some.class",
			"activator-capacity":                      "905",
			"scale-to-zero-pod-retention-period":      "2m3s",
		},
		want: func() *Config {
			c := defaultConfig()
			c.TargetBurstCapacity = 12345
			c.ContainerConcurrencyTargetDefault = 10.5
			c.ContainerConcurrencyTargetFraction = 0.71
			c.RPSTargetDefault = 10.11
			c.MaxScaleDownRate = 3
			c.MaxScaleUpRate = 1.01
			c.StableWindow = 5 * time.Minute
			c.ActivatorCapacity = 905
			c.PodAutoscalerClass = "some.class"
			c.ScaleToZeroPodRetentionPeriod = 2*time.Minute + 3*time.Second
			return c
		}(),
	}, {
		name: "with toggles on strange casing",
		input: map[string]string{
			"enable-scale-to-zero":      "TRUE",
			"enable-graceful-scaledown": "FALSE",
		},
		want: defaultConfig(),
	}, {
		name: "with toggles explicitly flipped",
		input: map[string]string{
			"enable-scale-to-zero":      "false",
			"enable-graceful-scaledown": "true",
		},
		want: func() *Config {
			c := defaultConfig()
			c.EnableScaleToZero = false
			c.EnableGracefulScaledown = true
			return c
		}(),
	}, {
		name: "with explicit grace period",
		input: map[string]string{
			"enable-scale-to-zero":       "false",
			"scale-to-zero-grace-period": "33s",
		},
		want: func() *Config {
			c := defaultConfig()
			c.EnableScaleToZero = false
			c.ScaleToZeroGracePeriod = 33 * time.Second
			return c
		}(),
	}, {
		name: "malformed float",
		input: map[string]string{
			"max-scale-up-rate": "not a float",
		},
		wantErr: true,
	}, {
		name: "invalid pod retention period",
		input: map[string]string{
			"scale-to-zero-pod-retention-period": "-4m11s",
		},
		wantErr: true,
	}, {
		name: "malformed duration",
		input: map[string]string{
			"stable-window": "not a duration",
		},
		wantErr: true,
	}, {
		name: "invalid target burst capacity",
		input: map[string]string{
			"target-burst-capacity": "-11",
		},
		wantErr: true,
	}, {
		name: "invalid target %, too small",
		input: map[string]string{
			"container-concurrency-target-percentage": "-42",
		},
		wantErr: true,
	}, {
		name: "invalid target %, too big",
		input: map[string]string{
			"container-concurrency-target-percentage": "142.4",
		},
		wantErr: true,
	}, {
		name: "invalid RPS target, too small",
		input: map[string]string{
			"requests-per-second-target-default": "-5.25",
		},
		wantErr: true,
	}, {
		name: "max scale up rate 1.0",
		input: map[string]string{
			"max-scale-up-rate": "1",
		},
		wantErr: true,
	}, {
		name: "max down down rate negative",
		input: map[string]string{
			"max-scale-down-rate": "-55",
		},
		wantErr: true,
	}, {
		name: "max down down rate 1.0",
		input: map[string]string{
			"max-scale-down-rate": "1",
		},
		wantErr: true,
	}, {
		name: "stable window too small",
		input: map[string]string{
			"stable-window": "1s",
		},
		wantErr: true,
	}, {
		name: "stable not seconds",
		input: map[string]string{
			"stable-window": "61984ms",
		},
		wantErr: true,
	}, {
		name: "activator-capacity invalid",
		input: map[string]string{
			"activator-capacity": "0.95",
		},
		wantErr: true,
	}, {
		name: "panic window percentage too small",
		input: map[string]string{
			"stable-window":           "12s",
			"panic-window-percentage": "5", // 0.6s < BucketSize
		},
		wantErr: true,
	}, {
		name: "panic window percentage too big",
		input: map[string]string{
			"stable-window":           "12s",
			"panic-window":            "3s",
			"panic-window-percentage": "110",
		},
		wantErr: true,
	}, {
		name: "TU*CC < 0.01",
		input: map[string]string{
			"container-concurrency-target-percentage": "1",
			"container-concurrency-target-default":    "0.001",
		},
		wantErr: true,
	}, {
		name: "grace window too small",
		input: map[string]string{
			"stable-window":              "12s",
			"scale-to-zero-grace-period": "4s",
		},
		wantErr: true,
	}, {
		name: "with prohibited default initial scale",
		input: map[string]string{
			"allow-zero-initial-scale": "false",
			"initial-scale":            "0",
		},
		wantErr: true,
	}, {
		name: "with negative default initial scale",
		input: map[string]string{
			"allow-zero-initial-scale": "false",
			"initial-scale":            "-1",
		},
		wantErr: true,
	}, {
		name: "with non-parseable default initial scale",
		input: map[string]string{
			"allow-zero-initial-scale": "false",
			"initial-scale":            "invalid",
		},
		wantErr: true,
	}, {
		name: "with valid default initial scale",
		input: map[string]string{
			"allow-zero-initial-scale": "true",
			"initial-scale":            "0",
		},
		want: func() *Config {
			c := defaultConfig()
			c.AllowZeroInitialScale = true
			c.InitialScale = 0
			return c
		}(),
	}, {
		name: "with non-parseable allow-zero-initial-scale",
		input: map[string]string{
			"allow-zero-initial-scale": "invalid",
		},
		want: func() *Config {
			c := defaultConfig()
			c.AllowZeroInitialScale = false
			return c
		}(),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewConfigFromConfigMap(&corev1.ConfigMap{
				Data: test.input,
			})
			t.Log("Error =", err)
			if (err != nil) != test.wantErr {
				t.Errorf("NewConfig() = %v, want %v", err, test.wantErr)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("NewConfig (-want, +got) = %v", diff)
			}
		})
	}
}

func TestOurConfig(t *testing.T) {
	cm, example := ConfigMapsFromTestFile(t, ConfigName)
	if _, err := NewConfigFromConfigMap(cm); err != nil {
		t.Errorf("NewConfigFromConfigMap(actual) = %v", err)
	}
	if cm, err := NewConfigFromConfigMap(example); err != nil {
		t.Errorf("NewConfigFromConfigMap(example) = %v", err)
	} else if got, want := cm, defaultConfig(); !cmp.Equal(want, got) {
		t.Errorf("ExampleConfig is not equal to defaults (-want, +got) = %s",
			cmp.Diff(want, got))
	}

}
