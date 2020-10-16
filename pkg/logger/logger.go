// Copyright 2020 Nexus Operator and/or its authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Logger struct {
	logr.Logger
}

func (l *Logger) Debug(message string, keysAndValues ...interface{}) {
	l.Logger.V(1).Info(message, keysAndValues...)
}

// Warn alternative for info format with sprintf and WARN named.
func (l *Logger) Warn(message string, keysAndValues ...interface{}) {
	l.Logger.WithName("WARNING").V(0).Info(message, keysAndValues...)
}

// GetLoggerWithResource returns a custom named logger with a resource's namespace and name as fields
func GetLoggerWithResource(name string, res resource.KubernetesResource) Logger {
	// we can't use framework.Key() here as it generates an import cycle (framework needs logging)
	key := types.NamespacedName{Name: res.GetName(), Namespace: res.GetNamespace()}
	return GetLoggerWithNamespacedName(name, key)
}

func GetLoggerWithNamespacedName(name string, key types.NamespacedName) Logger {
	logger := GetLogger(name)
	logger.WithValues("nexus", key)
	return logger
}

// GetLogger returns a custom named logger
func GetLogger(name string) Logger {
	logger := zap.New(zap.UseDevMode(true))
	logger.WithName(name)
	return Logger{Logger: logger}
}
