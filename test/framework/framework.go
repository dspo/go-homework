package framework

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/onsi/gomega"
	"go.uber.org/zap"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	Namespace = "go-example-e2e"
)

var (
	//go:embed manifests/app.yaml
	_appSpec string

	//go:embed manifests/mysql.yaml
	_mysqlSpec string
)

var (
	_f *Framework
)

func init() {
	var err error
	_f, err = newFramework()
	if err != nil {
		log.Fatalf("failed to init framework: %v", err)
	}
}

type Framework struct {
	scaffold *KubernetesScaffold
	db       *gorm.DB
	g        *gomega.GomegaWithT
	t        *testing.T
}

func GetFramework(t *testing.T) *Framework {
	if _f == nil {
		log.Fatalln("framework is not initialized")
	}

	var f = *_f
	f.t = t
	f.g = gomega.NewGomegaWithT(t)

	return &f
}

func (f *Framework) DB() *gorm.DB {
	return f.db
}

func (f *Framework) DeployComponents() {
	f.deployDatabase()
	f.initDatabase()
	f.connectDatabase()
	f.deployApp0()
}

func (f *Framework) deployDatabase() {
	f.t.Log("it is going to deploy MySQL")
	err := k8s.KubectlApplyFromStringE(f.t, f.scaffold.kubectlOptions, _mysqlSpec)
	f.g.Ω(err).ShouldNot(gomega.HaveOccurred())

	err = f.ensureServiceWithTimeout(f.t.Context(), "mysql", f.scaffold.kubectlOptions.Namespace, 1, 60)
	f.g.Ω(err).ShouldNot(gomega.HaveOccurred())
}

func (f *Framework) initDatabase() {
	f.t.Log("it is going to init MySQL")
	db, err := sql.Open("mysql", "root:changeme@tcp(mysql:3306)/")
	f.g.Ω(err).ShouldNot(gomega.HaveOccurred())

	defer func() { _ = db.Close() }()

	err = db.Ping()
	f.g.Ω(err).ShouldNot(gomega.HaveOccurred())

	// try to drop database
	_, err = db.Exec("DROP DATABASE IF EXISTS `go_dev`")
	f.g.Ω(err).ShouldNot(gomega.HaveOccurred())

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS `go_dev`")
	f.g.Ω(err).ShouldNot(gomega.HaveOccurred())
}

func (f *Framework) connectDatabase() {
	// todo: connection database using gorm
}

func (f *Framework) deployApp0() {
	f.t.Log("it is going to deploy app0")
	err := k8s.KubectlApplyFromStringE(f.t, f.scaffold.kubectlOptions, _appSpec)
	f.g.Ω(err).ShouldNot(gomega.HaveOccurred())

	err = f.ensureServiceWithTimeout(f.t.Context(), "app0", f.scaffold.kubectlOptions.Namespace, 1, 30)
	f.g.Ω(err).ShouldNot(gomega.HaveOccurred())
}

func (f *Framework) ensureServiceWithTimeout(ctx context.Context, name, namespace string, desiredEndpoints, timeout int) error {
	backoff := wait.Backoff{
		Duration: 6 * time.Second,
		Factor:   1,
		Steps:    timeout / 6,
	}
	var lastErr error
	condFunc := func() (bool, error) {
		endpointSliceList, err := f.scaffold.clientset.DiscoveryV1().EndpointSlices(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "kubernetes.io/service-name=" + name,
		})
		if err != nil {
			lastErr = err
			log.Println("ERROR: failed to list endpoints",
				zap.String("service", name),
				zap.Error(err),
			)
			return false, nil
		}
		if len(endpointSliceList.Items) == 0 {
			log.Println("ERROR: no EndpointSlice found",
				zap.String("service", name))
			return false, nil
		}
		var es = endpointSliceList.Items[0]
		count := 0
		for _, ep := range es.Endpoints {
			if ep.Conditions.Ready != nil && *ep.Conditions.Ready {
				count += len(ep.Addresses)
			}
		}
		if count == desiredEndpoints {
			return true, nil
		}
		log.Println("INFO: endpoints count mismatch",
			zap.String("service", name),
			zap.Any("ep", es),
			zap.Int("expected", desiredEndpoints),
			zap.Int("actual", count),
		)
		lastErr = fmt.Errorf("expected endpoints: %d but seen %d", desiredEndpoints, count)
		return false, nil
	}

	err := wait.ExponentialBackoff(backoff, condFunc)
	if err != nil {
		return lastErr
	}
	return nil
}

func newFramework() (*Framework, error) {
	_f = new(Framework)

	var (
		err error
	)
	_f.scaffold, err = NewKubernetesScaffold(KubectlOptions{
		ContextName: "",
		ConfigPath:  "",
		Namespace:   Namespace,
	})
	if err != nil {
		return nil, err
	}

	return _f, nil
}
