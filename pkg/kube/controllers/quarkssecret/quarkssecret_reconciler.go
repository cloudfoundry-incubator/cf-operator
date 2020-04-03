package quarkssecret

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	certv1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"code.cloudfoundry.org/cf-operator/pkg/credsgen"
	qsv1a1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/quarkssecret/v1alpha1"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/mutate"
	"code.cloudfoundry.org/quarks-utils/pkg/config"
	"code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
	"code.cloudfoundry.org/quarks-utils/pkg/meltdown"
	"code.cloudfoundry.org/quarks-utils/pkg/names"
)

type setReferenceFunc func(owner, object metav1.Object, scheme *runtime.Scheme) error

// NewQuarksSecretReconciler returns a new ReconcileQuarksSecret
func NewQuarksSecretReconciler(ctx context.Context, config *config.Config, mgr manager.Manager, generator credsgen.Generator, srf setReferenceFunc) reconcile.Reconciler {
	return &ReconcileQuarksSecret{
		ctx:          ctx,
		config:       config,
		client:       mgr.GetClient(),
		scheme:       mgr.GetScheme(),
		generator:    generator,
		setReference: srf,
	}
}

// ReconcileQuarksSecret reconciles an QuarksSecret object
type ReconcileQuarksSecret struct {
	ctx          context.Context
	client       client.Client
	generator    credsgen.Generator
	scheme       *runtime.Scheme
	setReference setReferenceFunc
	config       *config.Config
}

type caNotReadyError struct {
	message string
}

func newCaNotReadyError(message string) *caNotReadyError {
	return &caNotReadyError{message: message}
}

// Error returns the error message
func (e *caNotReadyError) Error() string {
	return e.message
}

func isCaNotReady(o interface{}) bool {
	err := o.(error)
	err = errors.Cause(err)
	_, ok := err.(*caNotReadyError)
	return ok
}

// Reconcile reads that state of the cluster for a QuarksSecret object and makes changes based on the state read
// and what is in the QuarksSecret.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileQuarksSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	qsec := &qsv1a1.QuarksSecret{}

	// Set the ctx to be Background, as the top-level context for incoming requests.
	ctx, cancel := context.WithTimeout(r.ctx, r.config.CtxTimeOut)
	defer cancel()

	ctxlog.Infof(ctx, "Reconciling QuarksSecret %s", request.NamespacedName)
	err := r.client.Get(ctx, request.NamespacedName, qsec)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			ctxlog.Info(ctx, "Skip reconcile: quarks secret not found")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		ctxlog.Info(ctx, "Error reading the object")
		return reconcile.Result{}, errors.Wrap(err, "Error reading quarksSecret")
	}

	if meltdown.NewWindow(r.config.MeltdownDuration, qsec.Status.LastReconcile).Contains(time.Now()) {
		ctxlog.WithEvent(qsec, "Meltdown").Debugf(ctx, "Resource '%s' is in meltdown, requeue reconcile after %s", qsec.GetNamespacedName(), r.config.MeltdownRequeueAfter)
		return reconcile.Result{RequeueAfter: r.config.MeltdownRequeueAfter}, nil
	}

	// Check if allowed to generate secret, could be already done or
	// created manually by a user
	skipReconcile, err := r.skipReconcile(ctx, qsec)
	if err != nil {
		ctxlog.Errorf(ctx, "Error reading the secret: %v", err.Error())
		return reconcile.Result{}, err
	}
	if skipReconcile {
		ctxlog.WithEvent(qsec, "SkipReconcile").Infof(ctx, "Skip reconcile: quarksSecret '%s' is already generated", request.NamespacedName)
		return reconcile.Result{}, nil
	}

	// Create secret
	switch qsec.Spec.Type {
	case qsv1a1.Password:
		ctxlog.Info(ctx, "Generating password")
		err = r.createPasswordSecret(ctx, qsec)
		if err != nil {
			ctxlog.Infof(ctx, "Error generating password secret: %s", err.Error())
			return reconcile.Result{}, errors.Wrap(err, "generating password secret failed.")
		}
	case qsv1a1.RSAKey:
		ctxlog.Info(ctx, "Generating RSA Key")
		err = r.createRSASecret(ctx, qsec)
		if err != nil {
			ctxlog.Infof(ctx, "Error generating RSA key secret: %s", err.Error())
			return reconcile.Result{}, errors.Wrap(err, "generating RSA key secret failed.")
		}
	case qsv1a1.SSHKey:
		ctxlog.Info(ctx, "Generating SSH Key")
		err = r.createSSHSecret(ctx, qsec)
		if err != nil {
			ctxlog.Infof(ctx, "Error generating SSH key secret: %s", err.Error())
			return reconcile.Result{}, errors.Wrap(err, "generating SSH key secret failed.")
		}
	case qsv1a1.Certificate:
		ctxlog.Info(ctx, "Generating certificate")
		err = r.createCertificateSecret(ctx, qsec)
		if err != nil {
			if isCaNotReady(err) {
				ctxlog.Info(ctx, fmt.Sprintf("CA for secret '%s' is not ready yet: %s", request.NamespacedName, err))
				return reconcile.Result{RequeueAfter: time.Second * 5}, nil
			}
			ctxlog.Info(ctx, "Error generating certificate secret: "+err.Error())
			return reconcile.Result{}, errors.Wrap(err, "generating certificate secret.")
		}
	default:
		err = ctxlog.WithEvent(qsec, "InvalidTypeError").Errorf(ctx, "Invalid type: %s", qsec.Spec.Type)
		return reconcile.Result{}, err
	}

	r.updateStatus(ctx, qsec)
	return reconcile.Result{}, nil
}

func (r *ReconcileQuarksSecret) updateStatus(ctx context.Context, qsec *qsv1a1.QuarksSecret) {
	qsec.Status.Generated = true

	now := metav1.Now()
	qsec.Status.LastReconcile = &now
	err := r.client.Status().Update(ctx, qsec)
	if err != nil {
		ctxlog.Errorf(ctx, "could not create or update QuarksSecret status '%s': %v", qsec.GetNamespacedName(), err)
	}
}

func (r *ReconcileQuarksSecret) createPasswordSecret(ctx context.Context, qsec *qsv1a1.QuarksSecret) error {
	request := credsgen.PasswordGenerationRequest{}
	password := r.generator.GeneratePassword(qsec.GetName(), request)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      qsec.Spec.SecretName,
			Namespace: qsec.GetNamespace(),
		},
		StringData: map[string]string{
			"password": password,
		},
	}

	return r.createSecret(ctx, qsec, secret)
}

func (r *ReconcileQuarksSecret) createRSASecret(ctx context.Context, qsec *qsv1a1.QuarksSecret) error {
	key, err := r.generator.GenerateRSAKey(qsec.GetName())
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      qsec.Spec.SecretName,
			Namespace: qsec.GetNamespace(),
		},
		StringData: map[string]string{
			"private_key": string(key.PrivateKey),
			"public_key":  string(key.PublicKey),
		},
	}

	return r.createSecret(ctx, qsec, secret)
}

func (r *ReconcileQuarksSecret) createSSHSecret(ctx context.Context, qsec *qsv1a1.QuarksSecret) error {
	key, err := r.generator.GenerateSSHKey(qsec.GetName())
	if err != nil {
		return err
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      qsec.Spec.SecretName,
			Namespace: qsec.GetNamespace(),
		},
		StringData: map[string]string{
			"private_key":            string(key.PrivateKey),
			"public_key":             string(key.PublicKey),
			"public_key_fingerprint": key.Fingerprint,
		},
	}

	return r.createSecret(ctx, qsec, secret)
}

func (r *ReconcileQuarksSecret) createCertificateSecret(ctx context.Context, qsec *qsv1a1.QuarksSecret) error {
	serviceIPForEKSWorkaround := ""

	for _, serviceRef := range qsec.Spec.Request.CertificateRequest.ServiceRef {
		service := &corev1.Service{}

		err := r.client.Get(ctx, types.NamespacedName{Namespace: qsec.Namespace, Name: serviceRef.Name}, service)

		if err != nil {
			return errors.Wrapf(err, "Failed to get service reference '%s' for QuarksSecret '%s'", serviceRef.Name, qsec.GetNamespacedName())
		}

		if serviceIPForEKSWorkaround == "" {
			serviceIPForEKSWorkaround = service.Spec.ClusterIP
		}

		qsec.Spec.Request.CertificateRequest.AlternativeNames = append(append(
			qsec.Spec.Request.CertificateRequest.AlternativeNames,
			service.Name,
			service.Name+"."+service.Namespace,
			"*."+service.Name,
			"*."+service.Name+"."+service.Namespace,
			service.Spec.ClusterIP,
			service.Spec.LoadBalancerIP,
			service.Spec.ExternalName,
		), service.Spec.ExternalIPs...)
	}

	if len(qsec.Spec.Request.CertificateRequest.SignerType) == 0 {
		qsec.Spec.Request.CertificateRequest.SignerType = qsv1a1.LocalSigner
	}

	generationRequest, err := r.generateCertificateGenerationRequest(ctx, qsec.Namespace, qsec.Spec.Request.CertificateRequest)
	if err != nil {
		return errors.Wrap(err, "generating certificate generation request")
	}

	switch qsec.Spec.Request.CertificateRequest.SignerType {
	case qsv1a1.ClusterSigner:
		if qsec.Spec.Request.CertificateRequest.ActivateEKSWorkaroundForSAN {
			if serviceIPForEKSWorkaround == "" {
				return errors.Errorf("can't activate EKS workaround for QuarksSecret '%s'; couldn't find a ClusterIP for any service reference", qsec.GetNamespacedName())
			}

			ctxlog.Infof(ctx, "Activating EKS workaround for QuarksSecret '%s'. Using IP '%s' as a common name. See 'https://github.com/awslabs/amazon-eks-ami/issues/341' for more details.", qsec.GetNamespacedName(), serviceIPForEKSWorkaround)

			generationRequest.CommonName = serviceIPForEKSWorkaround
		}

		ctxlog.Info(ctx, "Generating certificate signing request and its key")
		csr, key, err := r.generator.GenerateCertificateSigningRequest(generationRequest)
		if err != nil {
			return err
		}

		// private key Secret which will be merged to certificate Secret later
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      names.CsrPrivateKeySecretName(names.CSRName(qsec.Namespace, qsec.Name)),
				Namespace: qsec.GetNamespace(),
			},
			StringData: map[string]string{
				"private_key": string(key),
				"is_ca":       strconv.FormatBool(qsec.Spec.Request.CertificateRequest.IsCA),
			},
		}

		err = r.createSecret(ctx, qsec, secret)
		if err != nil {
			return err
		}

		return r.createCertificateSigningRequest(ctx, qsec, csr)
	case qsv1a1.LocalSigner:
		// Generate certificate
		cert, err := r.generator.GenerateCertificate(qsec.GetName(), generationRequest)
		if err != nil {
			return err
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      qsec.Spec.SecretName,
				Namespace: qsec.GetNamespace(),
			},
			StringData: map[string]string{
				"certificate": string(cert.Certificate),
				"private_key": string(cert.PrivateKey),
				"is_ca":       strconv.FormatBool(qsec.Spec.Request.CertificateRequest.IsCA),
			},
		}

		if len(generationRequest.CA.Certificate) > 0 {
			secret.StringData["ca"] = string(generationRequest.CA.Certificate)
		}

		return r.createSecret(ctx, qsec, secret)
	default:
		return fmt.Errorf("unrecognized signer type: %s", qsec.Spec.Request.CertificateRequest.SignerType)
	}
}

// Skip reconcile when
// * secret is already generated according to qsecs status field
// * secret exists, but was not generated (user created secret)
func (r *ReconcileQuarksSecret) skipReconcile(ctx context.Context, qsec *qsv1a1.QuarksSecret) (bool, error) {
	if qsec.Status.Generated {
		return true, nil
	}

	secretName := qsec.Spec.SecretName

	existingSecret := &corev1.Secret{}
	err := r.client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: qsec.GetNamespace()}, existingSecret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "could not get secret")
	}

	secretLabels := existingSecret.GetLabels()
	if secretLabels == nil {
		secretLabels = map[string]string{}
	}

	// skip the user generated secret
	if secretLabels[qsv1a1.LabelKind] != qsv1a1.GeneratedSecretKind {
		return true, nil
	}

	return false, nil
}

// createSecret applies common properties(labels and ownerReferences) to the secret and creates it
func (r *ReconcileQuarksSecret) createSecret(ctx context.Context, qsec *qsv1a1.QuarksSecret, secret *corev1.Secret) error {
	ctxlog.Debugf(ctx, "Creating secret '%s', owned by quarks secret '%s'", secret.Name, qsec.GetNamespacedName())

	secretLabels := secret.GetLabels()
	if secretLabels == nil {
		secretLabels = map[string]string{}
	}

	secretLabels[qsv1a1.LabelKind] = qsv1a1.GeneratedSecretKind

	secret.SetLabels(secretLabels)

	if err := r.setReference(qsec, secret, r.scheme); err != nil {
		return errors.Wrapf(err, "error setting owner for secret '%s' to QuarksSecret '%s'", secret.GetName(), qsec.GetNamespacedName())
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.client, secret, mutate.SecretMutateFn(secret))
	if err != nil {
		return errors.Wrapf(err, "could not create or update secret '%s/%s'", qsec.Namespace, secret.GetName())
	}

	if op != "unchanged" {
		ctxlog.Debugf(ctx, "Secret '%s' has been %s", secret.Name, op)
	}

	return nil
}

// generateCertificateGenerationRequest generates CertificateGenerationRequest for certificate
func (r *ReconcileQuarksSecret) generateCertificateGenerationRequest(ctx context.Context, namespace string, certificateRequest qsv1a1.CertificateRequest) (credsgen.CertificateGenerationRequest, error) {
	var request credsgen.CertificateGenerationRequest
	switch certificateRequest.SignerType {
	case qsv1a1.ClusterSigner:
		// Generate cluster-signed CA certificate
		request = credsgen.CertificateGenerationRequest{
			CommonName:       certificateRequest.CommonName,
			AlternativeNames: certificateRequest.AlternativeNames,
		}
	case qsv1a1.LocalSigner:
		// Generate local-issued CA certificate
		request = credsgen.CertificateGenerationRequest{
			IsCA:             certificateRequest.IsCA,
			CommonName:       certificateRequest.CommonName,
			AlternativeNames: certificateRequest.AlternativeNames,
		}

		if len(certificateRequest.CARef.Name) > 0 {
			// Get CA certificate
			caSecret := &corev1.Secret{}
			caNamespacedName := types.NamespacedName{
				Namespace: namespace,
				Name:      certificateRequest.CARef.Name,
			}
			err := r.client.Get(ctx, caNamespacedName, caSecret)
			if err != nil {
				if apierrors.IsNotFound(err) {
					return request, newCaNotReadyError("CA secret not found")
				}
				return request, errors.Wrap(err, "getting CA secret")
			}
			ca := caSecret.Data[certificateRequest.CARef.Key]

			// Get CA key
			if certificateRequest.CAKeyRef.Name != certificateRequest.CARef.Name {
				caSecret = &corev1.Secret{}
				caNamespacedName = types.NamespacedName{
					Namespace: namespace,
					Name:      certificateRequest.CAKeyRef.Name,
				}
				err = r.client.Get(ctx, caNamespacedName, caSecret)
				if err != nil {
					if apierrors.IsNotFound(err) {
						return request, newCaNotReadyError("CA key secret not found")
					}
					return request, errors.Wrap(err, "getting CA Key secret")
				}
			}
			key := caSecret.Data[certificateRequest.CAKeyRef.Key]
			request.CA = credsgen.Certificate{
				IsCA:        true,
				PrivateKey:  key,
				Certificate: ca,
			}
		}
	default:
		return request, fmt.Errorf("unrecognized signer type: %s", certificateRequest.SignerType)
	}

	return request, nil
}

// createCertificateSigningRequest creates CertificateSigningRequest Object
func (r *ReconcileQuarksSecret) createCertificateSigningRequest(ctx context.Context, qsec *qsv1a1.QuarksSecret, csr []byte) error {
	csrName := names.CSRName(qsec.Namespace, qsec.Name)
	ctxlog.Debugf(ctx, "Creating certificatesigningrequest '%s'", csrName)

	annotations := qsec.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[qsv1a1.AnnotationCertSecretName] = qsec.Spec.SecretName
	annotations[qsv1a1.AnnotationQSecNamespace] = qsec.Namespace
	annotations[qsv1a1.AnnotationQSecName] = qsec.Name

	csrObj := &certv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:        csrName,
			Labels:      qsec.Labels,
			Annotations: annotations,
		},
		Spec: certv1.CertificateSigningRequestSpec{
			Request: csr,
			Usages:  qsec.Spec.Request.CertificateRequest.Usages,
		},
	}

	oldCsrObj := &certv1.CertificateSigningRequest{}

	// CSR spec is immutable after the request is created
	err := r.client.Get(ctx, types.NamespacedName{Name: csrObj.Name}, oldCsrObj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = r.client.Create(ctx, csrObj)
			if err != nil {
				return errors.Wrapf(err, "could not create certificatesigningrequest '%s'", csrObj.Name)
			}
			return nil
		}
		return errors.Wrapf(err, "could not get certificatesigningrequest '%s'", csrObj.Name)
	}

	ctxlog.Infof(ctx, "Ignoring immutable CSR '%s'", csrObj.Name)
	return nil
}
