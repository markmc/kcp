/*
Copyright 2021 The KCP Authors.

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

package helper

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/kcp-dev/logicalcluster/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/tenancy/v1beta1"
)

// IsValidCluster indicates whether a cluster is valid based on whether it
// adheres to logical cluster naming requirements and is rooted at root or
// system.
func IsValidCluster(cluster logicalcluster.Name) bool {
	if !cluster.IsValid() {
		return false
	}

	return cluster.HasPrefix(v1alpha1.RootCluster) || cluster.HasPrefix(logicalcluster.New("system"))
}

// QualifiedObjectName builds a fully qualified identifier for an object
// consisting of its logical cluster, namespace if applicable, and object
// metadata name.
func QualifiedObjectName(obj metav1.Object) string {
	if len(obj.GetNamespace()) > 0 {
		return fmt.Sprintf("%s|%s/%s", logicalcluster.From(obj), obj.GetNamespace(), obj.GetName())
	}
	return fmt.Sprintf("%s|%s", logicalcluster.From(obj), obj.GetName())
}

// WorkspaceLabelSelector builds a label selector for objects associated with a
// given workspace.
func WorkspaceLabelSelector(name string) string {
	return fmt.Sprintf("%s=%s", v1beta1.WorkspaceNameLabel, name)
}

// DefaultRootPathPrefix is basically constant forever, or we risk a breaking change. The
// kubectl plugin for example will use this prefix to generate the root path, and because
// we don't control kubectl plugin updates, we cannot change this prefix.
const DefaultRootPathPrefix string = "/services"

// ParseClusterURL parses a cluster workspace URL and returns both the
// base URL (i.e. with the clusters prefix removed) and the cluster name
func ParseClusterURL(host string) (*url.URL, logicalcluster.Name, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, logicalcluster.Name{}, err
	}
	ret := *u
	var clusterName logicalcluster.Name
	for _, prefix := range []string{"/clusters/", path.Join(DefaultRootPathPrefix, "workspaces") + "/"} {
		if clusterIndex := strings.Index(u.Path, prefix); clusterIndex >= 0 {
			clusterName = logicalcluster.New(strings.SplitN(ret.Path[clusterIndex+len(prefix):], "/", 2)[0])
			ret.Path = ret.Path[:clusterIndex]
			break
		}
	}
	if clusterName.Empty() || !IsValidCluster(clusterName) {
		return nil, logicalcluster.Name{}, fmt.Errorf("current cluster URL %s is not pointing to a cluster workspace", u)
	}

	return &ret, clusterName, nil
}
