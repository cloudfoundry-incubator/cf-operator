package controllers_test

import (
	"context"
	"encoding/base64"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"code.cloudfoundry.org/cf-operator/pkg/credsgen"
	gfakes "code.cloudfoundry.org/cf-operator/pkg/credsgen/fakes"
	"code.cloudfoundry.org/cf-operator/pkg/kube/controllers"
	cfakes "code.cloudfoundry.org/cf-operator/pkg/kube/controllers/fakes"
	ocontext "code.cloudfoundry.org/cf-operator/pkg/kube/util/context"
	"code.cloudfoundry.org/cf-operator/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Controllers", func() {
	Describe("AddToScheme", func() {
		It("registers our schemes with the operator", func() {
			scheme := scheme.Scheme
			controllers.AddToScheme(scheme)
			kinds := []string{}
			for k, _ := range scheme.AllKnownTypes() {
				kinds = append(kinds, k.Kind)
			}
			Expect(kinds).To(ContainElement("BOSHDeployment"))
			Expect(kinds).To(ContainElement("ExtendedSecret"))
			Expect(kinds).To(ContainElement("ExtendedStatefulSet"))
		})
	})

	// "AddToManager" tested via integration tests

	Describe("AddHooks", func() {
		var (
			manager    *cfakes.FakeManager
			client     *cfakes.FakeClient
			log        *zap.SugaredLogger
			ctrsConfig *ocontext.Config
			generator  *gfakes.FakeGenerator
			env        testing.Catalog
		)

		BeforeEach(func() {
			controllers.AddToScheme(scheme.Scheme)
			client = &cfakes.FakeClient{}
			restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})
			restMapper.Add(schema.GroupVersionKind{Group: "", Kind: "Pod", Version: "v1"}, meta.RESTScopeNamespace)

			manager = &cfakes.FakeManager{}
			manager.GetSchemeReturns(scheme.Scheme)
			manager.GetClientReturns(client)
			manager.GetRESTMapperReturns(restMapper)

			core, _ := observer.New(zapcore.InfoLevel)
			log = zap.New(core).Sugar()
			generator = &gfakes.FakeGenerator{}
			generator.GenerateCertificateReturns(credsgen.Certificate{Certificate: []byte("thecert")}, nil)

			ctrsConfig = env.DefaultContextConfig()
		})

		It("sets the operator namespace label", func() {
			client.UpdateCalls(func(_ context.Context, object runtime.Object) error {
				ns := object.(*unstructured.Unstructured)
				labels := ns.GetLabels()

				Expect(labels["cf-operator-ns"]).To(Equal(ctrsConfig.Namespace))

				return nil
			})

			err := controllers.AddHooks(log, ctrsConfig, manager, generator)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("if there is no cert secret yet", func() {
			It("generates and persists the certificates on disk and in a secret", func() {
				Expect(afero.Exists(ctrsConfig.Fs, "/tmp/cf-operator-certs/key.pem")).To(BeFalse())

				err := controllers.AddHooks(log, ctrsConfig, manager, generator)
				Expect(err).ToNot(HaveOccurred())

				Expect(afero.Exists(ctrsConfig.Fs, "/tmp/cf-operator-certs/key.pem")).To(BeTrue())
				Expect(generator.GenerateCertificateCallCount()).To(Equal(2)) // Generate CA and certificate
				Expect(client.CreateCallCount()).To(Equal(2))                 // Persist secret and the webhook config
			})
		})

		Context("if there is a persisted cert secret already", func() {
			BeforeEach(func() {
				secret := &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name":      "cf-operator-webhook-server-cert",
							"namespace": ctrsConfig.Namespace,
						},
						"data": map[string]interface{}{
							"certificate":    base64.StdEncoding.EncodeToString([]byte("the-cert")),
							"private_key":    base64.StdEncoding.EncodeToString([]byte("the-key")),
							"ca_certificate": base64.StdEncoding.EncodeToString([]byte("the-ca-cert")),
							"ca_private_key": base64.StdEncoding.EncodeToString([]byte("the-ca-key")),
						},
					},
				}
				client.GetCalls(func(context context.Context, nn types.NamespacedName, object runtime.Object) error {
					switch object.(type) {
					case *unstructured.Unstructured:
						secret.DeepCopyInto(object.(*unstructured.Unstructured))
						return nil
					}
					return apierrors.NewNotFound(schema.GroupResource{}, nn.Name)
				})
			})

			It("does not overwrite the existing secret", func() {
				err := controllers.AddHooks(log, ctrsConfig, manager, generator)
				Expect(err).ToNot(HaveOccurred())
				Expect(client.CreateCallCount()).To(Equal(1)) // webhook config
			})

			It("generates the webhook configuration", func() {
				client.CreateCalls(func(context context.Context, object runtime.Object) error {
					config := object.(*admissionregistrationv1beta1.MutatingWebhookConfiguration)
					Expect(config.Name).To(Equal("cf-operator-mutating-hook-" + ctrsConfig.Namespace))
					Expect(len(config.Webhooks)).To(Equal(1))

					wh := config.Webhooks[0]
					Expect(wh.Name).To(Equal("mutatepods.example.com"))
					Expect(*wh.ClientConfig.URL).To(Equal("https://foo.com:1234/mutate-pods"))
					Expect(wh.ClientConfig.CABundle).To(ContainSubstring("the-ca-cert"))
					Expect(*wh.FailurePolicy).To(Equal(admissionregistrationv1beta1.Fail))
					return nil
				})
				err := controllers.AddHooks(log, ctrsConfig, manager, generator)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
