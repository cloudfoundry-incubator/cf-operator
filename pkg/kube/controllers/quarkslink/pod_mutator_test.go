package quarkslink_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"gomodules.xyz/jsonpatch/v2"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"code.cloudfoundry.org/quarks-operator/pkg/kube/controllers/quarkslink"
	"code.cloudfoundry.org/quarks-operator/testing"
	"code.cloudfoundry.org/quarks-utils/pkg/config"
	"code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
	helper "code.cloudfoundry.org/quarks-utils/testing/testhelper"
)

var _ = Describe("Mount quarks link secret on entangled pods", func() {
	const (
		deploymentName = "nats-deployment"
		consumesNats   = `[{"name":"nats","type":"nats"}]`
		consumesNuts   = `[{"name":"nats","type":"nuts"}]`
	)

	var (
		client             client.Client
		ctx                context.Context
		decoder            *admission.Decoder
		entanglementSecret corev1.Secret
		env                testing.Catalog
		log                *zap.SugaredLogger
		mutator            admission.Handler
		pod                corev1.Pod
		request            admission.Request
		response           admission.Response
	)
	annotationPatch := `{"op":"add","path":"/metadata/annotations/quarks.cloudfoundry.org~1restart-on-update","value":"true"}`

	podPatch := `{"op":"add","path":"/spec/volumes","value":[{"name":"link-nats-nats","secret":{"secretName":"link-nats-nats"}}]}`
	containerPatch := `{"op":"add","path":"/spec/containers/0/volumeMounts","value":[{"mountPath":"/quarks/link/nats-deployment/nats-nats","name":"link-nats-nats","readOnly":true}]}`
	secondContainerPatch := `{"op":"add","path":"/spec/containers/1/volumeMounts","value":[{"mountPath":"/quarks/link/nats-deployment/nats-nats","name":"link-nats-nats","readOnly":true}]}`

	jsonPatches := func(operations []jsonpatch.Operation) []string {
		patches := make([]string, len(operations))
		for i, patch := range operations {
			patches[i] = patch.Json()
		}
		return patches
	}

	newAdmissionRequest := func(pod corev1.Pod) admission.Request {
		raw, _ := json.Marshal(pod)
		return admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Object: runtime.RawExtension{Raw: raw},
			},
		}
	}

	BeforeEach(func() {
		_, log = helper.NewTestLogger()
		ctx = ctxlog.NewParentContext(log)

		mutator = quarkslink.NewPodMutator(log, &config.Config{CtxTimeOut: 10 * time.Second})

		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		decoder, _ = admission.NewDecoder(scheme)
		_ = mutator.(admission.DecoderInjector).InjectDecoder(decoder)

		entanglementSecret = env.DefaultQuarksLinkSecret(deploymentName, "nats")
	})

	JustBeforeEach(func() {
		_ = mutator.(inject.Client).InjectClient(client)
		response = mutator.Handle(ctx, request)
	})

	Context("when pod has no entanglement annotation", func() {
		BeforeEach(func() {
			pod = env.DefaultPod("test-pod")
			request = newAdmissionRequest(pod)
			client = fake.NewClientBuilder().
				WithObjects(&entanglementSecret).
				Build()
		})

		It("does not apply changes", func() {
			Expect(response.AdmissionResponse.Allowed).To(BeTrue())
			Expect(response.Patches).To(BeEmpty())
		})
	})

	Context("when valid bosh entanglement exists on pod", func() {
		BeforeEach(func() {
			pod = env.AnnotatedPod("entangled-pod", map[string]string{
				quarkslink.DeploymentKey: deploymentName,
				quarkslink.ConsumesKey:   consumesNats,
			})
			pod.Spec.Containers = []corev1.Container{
				{Name: "first", Image: "busybox", Command: []string{"sleep", "3600"}},
				{Name: "second", Image: "busybox", Command: []string{"sleep", "3600"}},
			}
			request = newAdmissionRequest(pod)
		})

		Context("when entanglement secret exists", func() {
			BeforeEach(func() {
				client = fake.NewClientBuilder().
					WithObjects(&entanglementSecret).
					Build()
			})

			It("secret is mounted on all containers", func() {
				Expect(response.Allowed).To(BeTrue(), response.Result)

				Expect(response.Patches).To(HaveLen(6))
				patches := jsonPatches(response.Patches)
				Expect(patches).To(ContainElement(podPatch))
				Expect(patches).To(ContainElement(containerPatch))
				Expect(patches).To(ContainElement(secondContainerPatch))

				Expect(response.AdmissionResponse.Allowed).To(BeTrue())
			})
		})

		Context("when entanglement exists", func() {
			BeforeEach(func() {
				client = fake.NewClientBuilder().
					WithObjects(&entanglementSecret).
					Build()
			})

			It("adds a quarks restart annotation", func() {
				Expect(response.Allowed).To(BeTrue(), response.Result)

				Expect(response.Patches).To(HaveLen(6))
				patches := jsonPatches(response.Patches)
				Expect(patches).To(ContainElement(annotationPatch))

				Expect(response.AdmissionResponse.Allowed).To(BeTrue())
			})
		})

		Context("when quarks link secret doesn't exist", func() {
			BeforeEach(func() {
				client = fake.NewClientBuilder().
					Build()
			})

			It("does not mutate the pod and errors", func() {
				Expect(response.Patches).To(BeEmpty())
				Expect(response.AdmissionResponse.Allowed).To(BeFalse())
			})
		})
	})

	Context("when invalid bosh entanglement exists on pod", func() {
		BeforeEach(func() {
			pod = env.AnnotatedPod("entangled-pod", map[string]string{
				quarkslink.DeploymentKey: "nuts",
				quarkslink.ConsumesKey:   consumesNuts,
			})
			request = newAdmissionRequest(pod)
			client = fake.NewClientBuilder().
				WithObjects(&entanglementSecret).
				Build()
		})

		It("does not mutate the pod and errors", func() {
			Expect(response.Patches).To(BeEmpty())
			Expect(response.AdmissionResponse.Allowed).To(BeFalse())
		})
	})

	Context("when pod has existing volumes", func() {
		podPatch := `{"op":"add","path":"/spec/volumes/1","value":{"name":"link-nats-nats","secret":{"secretName":"link-nats-nats"}}}`
		containerPatch := `{"op":"add","path":"/spec/containers/0/volumeMounts/1","value":{"mountPath":"/quarks/link/nats-deployment/nats-nats","name":"link-nats-nats","readOnly":true}}`
		envVarsPatch := `{"op":"add","path":"/spec/containers/0/env","value":[{"name":"LINK_NATS_PASSWORD","valueFrom":{"secretKeyRef":{"key":"nats.password","name":"link-nats-nats"}}},{"name":"LINK_NATS_PORT","valueFrom":{"secretKeyRef":{"key":"nats.port","name":"link-nats-nats"}}},{"name":"LINK_NATS_USER","valueFrom":{"secretKeyRef":{"key":"nats.user","name":"link-nats-nats"}}}]}`

		BeforeEach(func() {
			pod = env.NatsPod("entangled-pod")
			pod.SetAnnotations(map[string]string{
				quarkslink.DeploymentKey: deploymentName,
				quarkslink.ConsumesKey:   consumesNats,
			})
			request = newAdmissionRequest(pod)
		})

		Context("when pod has a non-secret volume", func() {
			BeforeEach(func() {
				client = fake.NewClientBuilder().
					WithObjects(&entanglementSecret).
					Build()
			})

			It("does add the link volume and mounts it on all containers", func() {
				Expect(response.Allowed).To(BeTrue(), response.Result)
				Expect(response.Patches).To(HaveLen(4))
				patches := jsonPatches(response.Patches)
				Expect(patches).To(ContainElement(podPatch))
				Expect(patches).To(ContainElement(containerPatch))
				Expect(patches).To(ContainElement(envVarsPatch))
				Expect(response.AdmissionResponse.Allowed).To(BeTrue())
			})
		})
	})
})
