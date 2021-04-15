package preflight

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/replicatedhq/kots/kotskinds/client/kotsclientset/scheme"
	"github.com/replicatedhq/kots/pkg/k8sutil"
	kotstypes "github.com/replicatedhq/kots/pkg/kotsadm/types"
	downstream "github.com/replicatedhq/kots/pkg/kotsadmdownstream"
	"github.com/replicatedhq/kots/pkg/kotsutil"
	"github.com/replicatedhq/kots/pkg/logger"
	"github.com/replicatedhq/kots/pkg/registry"
	registrytypes "github.com/replicatedhq/kots/pkg/registry/types"
	"github.com/replicatedhq/kots/pkg/render"
	"github.com/replicatedhq/kots/pkg/render/helper"
	"github.com/replicatedhq/kots/pkg/reporting"
	"github.com/replicatedhq/kots/pkg/store"
	"github.com/replicatedhq/kots/pkg/version"
	troubleshootv1beta2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	troubleshootpreflight "github.com/replicatedhq/troubleshoot/pkg/preflight"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

const (
	SpecDataKey = "preflight-spec"
)

func Run(appID string, appSlug string, sequence int64, isAirgap bool, archiveDir string) error {
	renderedKotsKinds, err := kotsutil.LoadKotsKindsFromPath(archiveDir)
	if err != nil {
		return errors.Wrap(err, "failed to load rendered kots kinds")
	}

	status, err := downstream.GetDownstreamVersionStatus(appID, sequence)
	if err != nil {
		return errors.Wrapf(err, "failed to check downstream version %d status", sequence)
	}

	// preflights should not run until config is finished
	if status == "pending_config" {
		logger.Debug("not running preflights for app that is pending required configuration",
			zap.String("appID", appID),
			zap.Int64("sequence", sequence))
		return nil
	}

	if renderedKotsKinds.Preflight != nil {
		status, err := downstream.GetDownstreamVersionStatus(appID, sequence)
		if err != nil {
			return errors.Wrap(err, "failed to get version status")
		}

		if status != "deployed" {
			if err := downstream.SetDownstreamVersionPendingPreflight(appID, sequence); err != nil {
				return errors.Wrapf(err, "failed to set downstream version %d pending preflight", sequence)
			}
		}

		ignoreRBAC, err := downstream.GetIgnoreRBACErrors(appID, sequence)
		if err != nil {
			return errors.Wrap(err, "failed to get ignore rbac flag")
		}

		// render the preflight file
		// we need to convert to bytes first, so that we can reuse the renderfile function
		renderedMarshalledPreflights, err := renderedKotsKinds.Marshal("troubleshoot.replicated.com", "v1beta1", "Preflight")
		if err != nil {
			return errors.Wrap(err, "failed to marshal rendered preflight")
		}

		registrySettings, err := store.GetStore().GetRegistryDetailsForApp(appID)
		if err != nil {
			return errors.Wrap(err, "failed to get registry settings for app")
		}

		renderedPreflight, err := render.RenderFile(renderedKotsKinds, registrySettings, appSlug, sequence, isAirgap, []byte(renderedMarshalledPreflights))
		if err != nil {
			return errors.Wrap(err, "failed to render preflights")
		}
		p, err := kotsutil.LoadPreflightFromContents(renderedPreflight)
		if err != nil {
			return errors.Wrap(err, "failed to load rendered preflight")
		}

		injectDefaultPreflights(p, renderedKotsKinds, registrySettings)

		collectors, err := registry.UpdateCollectorSpecsWithRegistryData(p.Spec.Collectors, registrySettings, renderedKotsKinds.Installation.Spec.KnownImages, renderedKotsKinds.License)
		if err != nil {
			return errors.Wrap(err, "failed to rewrite images in preflight")
		}
		p.Spec.Collectors = collectors

		go func() {
			logger.Debug("preflight checks beginning")
			uploadPreflightResults, err := execute(appID, sequence, p, ignoreRBAC)
			if err != nil {
				err = errors.Wrap(err, "failed to run preflight checks")
				logger.Error(err)
				return
			}
			logger.Debug("preflight checks completed")

			err = maybeDeployFirstVersion(appID, sequence, uploadPreflightResults)
			if err != nil {
				err = errors.Wrap(err, "failed to deploy first version")
				logger.Error(err)
				return
			}

			if err := reporting.SendPreflightInfo(appID, int(sequence), false, false); err != nil {
				logger.Debugf("failed to send preflights data to replicated app: %v", err)
				return
			}
		}()
	} else if sequence == 0 {
		err := maybeDeployFirstVersion(appID, sequence, &troubleshootpreflight.UploadPreflightResults{})
		if err != nil {
			return errors.Wrap(err, "failed to deploy first version")
		}
	} else {
		status, err := downstream.GetDownstreamVersionStatus(appID, sequence)
		if err != nil {
			return errors.Wrap(err, "failed to get version status")
		}
		if status != "deployed" {
			if err := downstream.SetDownstreamVersionReady(appID, sequence); err != nil {
				return errors.Wrap(err, "failed to set downstream version ready")
			}
		}
	}

	return nil
}

// maybeDeployFirstVersion will deploy the first version if
// 1. preflight checks pass
// 2. we have not already deployed it
func maybeDeployFirstVersion(appID string, sequence int64, preflightResults *troubleshootpreflight.UploadPreflightResults) error {
	if sequence != 0 {
		return nil
	}

	app, err := store.GetStore().GetApp(appID)
	if err != nil {
		return errors.Wrap(err, "failed to get app")
	}

	// do not revert to first version
	if app.CurrentSequence != 0 {
		return nil
	}

	preflightState := getPreflightState(preflightResults)
	if preflightState != "pass" {
		return nil
	}

	logger.Debug("automatically deploying first app version")

	// note: this may attempt to re-deploy the first version but the operator will take care of
	// comparing the version to current

	return version.DeployVersion(appID, sequence)
}

func getPreflightState(preflightResults *troubleshootpreflight.UploadPreflightResults) string {
	if len(preflightResults.Errors) > 0 {
		return "fail"
	}

	if len(preflightResults.Results) == 0 {
		return "pass"
	}

	state := "pass"
	for _, result := range preflightResults.Results {
		if result.IsFail {
			return "fail"
		} else if result.IsWarn {
			state = "warn"
		}
	}

	return state
}

func GetSpecSecretName(appSlug string) string {
	return fmt.Sprintf("kotsadm-%s-preflight", appSlug)
}

func GetSpecURI(appSlug string) string {
	return fmt.Sprintf("secret/%s/%s", os.Getenv("POD_NAMESPACE"), GetSpecSecretName(appSlug))
}

func GetPreflightCommand(appSlug string) []string {
	comamnd := []string{
		"curl https://krew.sh/preflight | bash",
		fmt.Sprintf("kubectl preflight %s\n", GetSpecURI(appSlug)),
	}

	return comamnd
}

func CreateRenderedSpec(appID string, sequence int64, origin string, inCluster bool, kotsKinds *kotsutil.KotsKinds) error {
	builtPreflight := kotsKinds.Preflight.DeepCopy()
	if builtPreflight == nil {
		builtPreflight = &troubleshootv1beta2.Preflight{
			TypeMeta: v1.TypeMeta{
				Kind:       "Preflight",
				APIVersion: "troubleshoot.sh/v1beta2",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "default-preflight",
			},
		}
	}

	app, err := store.GetStore().GetApp(appID)
	if err != nil {
		return errors.Wrap(err, "failed to get app")
	}

	registrySettings, err := store.GetStore().GetRegistryDetailsForApp(appID)
	if err != nil {
		return errors.Wrap(err, "failed to get registry settings for app")
	}

	injectDefaultPreflights(builtPreflight, kotsKinds, registrySettings)

	collectors, err := registry.UpdateCollectorSpecsWithRegistryData(builtPreflight.Spec.Collectors, registrySettings, kotsKinds.Installation.Spec.KnownImages, kotsKinds.License)
	if err != nil {
		return errors.Wrap(err, "failed to rewrite images in preflight")
	}
	builtPreflight.Spec.Collectors = collectors

	baseURL := os.Getenv("API_ADVERTISE_ENDPOINT")
	if inCluster {
		baseURL = os.Getenv("API_ENDPOINT")
	} else if origin != "" {
		baseURL = origin
	}
	builtPreflight.Spec.UploadResultsTo = fmt.Sprintf("%s/api/v1/preflight/app/%s/sequence/%d", baseURL, app.Slug, sequence)

	s := serializer.NewYAMLSerializer(serializer.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	var b bytes.Buffer
	if err := s.Encode(builtPreflight, &b); err != nil {
		return errors.Wrap(err, "failed to encode preflight")
	}

	templatedSpec := b.Bytes()

	renderedSpec, err := helper.RenderAppFile(app, &sequence, templatedSpec, kotsKinds)
	if err != nil {
		return errors.Wrap(err, "failed render preflight spec")
	}

	clientset, err := k8sutil.GetClientset()
	if err != nil {
		return errors.Wrap(err, "failed to get k8s clientset")
	}

	secretName := GetSpecSecretName(app.Slug)

	existingSecret, err := clientset.CoreV1().Secrets(os.Getenv("POD_NAMESPACE")).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil && !kuberneteserrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to read preflight secret")
	} else if kuberneteserrors.IsNotFound(err) {
		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: os.Getenv("POD_NAMESPACE"),
				Labels:    kotstypes.GetKotsadmLabels(),
			},
			Data: map[string][]byte{
				SpecDataKey: renderedSpec,
			},
		}

		_, err = clientset.CoreV1().Secrets(os.Getenv("POD_NAMESPACE")).Create(context.TODO(), secret, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create preflight secret")
		}

		return nil
	}

	if existingSecret.Data == nil {
		existingSecret.Data = map[string][]byte{}
	}
	existingSecret.Data[SpecDataKey] = renderedSpec
	existingSecret.ObjectMeta.Labels = kotstypes.GetKotsadmLabels()

	_, err = clientset.CoreV1().Secrets(os.Getenv("POD_NAMESPACE")).Update(context.TODO(), existingSecret, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update preflight secret")
	}

	return nil
}

func injectDefaultPreflights(preflight *troubleshootv1beta2.Preflight, kotskinds *kotsutil.KotsKinds, registrySettings registrytypes.RegistrySettings) {
	if !registrySettings.IsValid() || !registrySettings.IsReadOnly {
		return
	}

	// Get images from Installation.KnownImages, see UpdateCollectorSpecsWithRegistryData
	images := []string{}
	for _, image := range kotskinds.Installation.Spec.KnownImages {
		images = append(images, image.Image)
	}

	preflight.Spec.Collectors = append(preflight.Spec.Collectors, &troubleshootv1beta2.Collect{
		RegistryImages: &troubleshootv1beta2.RegistryImages{
			Images: images,
		},
	})
	preflight.Spec.Analyzers = append(preflight.Spec.Analyzers, &troubleshootv1beta2.Analyze{
		RegistryImages: &troubleshootv1beta2.RegistryImagesAnalyze{
			AnalyzeMeta: troubleshootv1beta2.AnalyzeMeta{
				CheckName: "Private Registry Images Available",
			},
			Outcomes: []*troubleshootv1beta2.Outcome{
				{
					Fail: &troubleshootv1beta2.SingleOutcome{
						When:    "missing > 0",
						Message: "Application uses images that cannot be found in the private registry",
					},
				},
				{
					Warn: &troubleshootv1beta2.SingleOutcome{
						When:    "errors > 0",
						Message: "Availability of application images in the private registry could not be verified.",
					},
				},
				{
					Pass: &troubleshootv1beta2.SingleOutcome{
						Message: "All images used by the application are present in the private registry",
					},
				},
			},
		},
	})
}
