package kotsutil

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	kotsv1beta1 "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	kotsscheme "github.com/replicatedhq/kots/kotskinds/client/kotsclientset/scheme"
	"github.com/replicatedhq/kots/pkg/binaries"
	"github.com/replicatedhq/kots/pkg/buildversion"
	"github.com/replicatedhq/kots/pkg/crypto"
	"github.com/replicatedhq/kots/pkg/k8sutil"
	"github.com/replicatedhq/kots/pkg/logger"
	"github.com/replicatedhq/kots/pkg/util"
	troubleshootv1beta2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	troubleshootscheme "github.com/replicatedhq/troubleshoot/pkg/client/troubleshootclientset/scheme"
	"github.com/replicatedhq/troubleshoot/pkg/docrewrite"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	applicationv1beta1 "sigs.k8s.io/application/api/v1beta1"
)

func init() {
	kotsscheme.AddToScheme(scheme.Scheme)
	troubleshootscheme.AddToScheme(scheme.Scheme)
	velerov1.AddToScheme(scheme.Scheme)
	applicationv1beta1.AddToScheme(scheme.Scheme)
}

// KotsKinds are all of the special "client-side" kinds that are packaged in
// an application. These should be pointers because they are all optional.
// But a few are still expected in the code later, so we make them not pointers,
// because other codepaths expect them to be present
type KotsKinds struct {
	KotsApplication kotsv1beta1.Application
	Application     *applicationv1beta1.Application
	HelmCharts      []*kotsv1beta1.HelmChart

	Collector     *troubleshootv1beta2.Collector
	Preflight     *troubleshootv1beta2.Preflight
	Analyzer      *troubleshootv1beta2.Analyzer
	SupportBundle *troubleshootv1beta2.SupportBundle
	Redactor      *troubleshootv1beta2.Redactor

	Config       *kotsv1beta1.Config
	ConfigValues *kotsv1beta1.ConfigValues

	Installation kotsv1beta1.Installation
	License      *kotsv1beta1.License

	Identity       *kotsv1beta1.Identity
	IdentityConfig *kotsv1beta1.IdentityConfig

	Backup *velerov1.Backup
}

func (k *KotsKinds) EncryptConfigValues() error {
	if k.ConfigValues == nil || k.Config == nil {
		return nil
	}

	updated := map[string]kotsv1beta1.ConfigValue{}

	for name, configValue := range k.ConfigValues.Spec.Values {
		updated[name] = configValue

		if configValue.ValuePlaintext != "" {
			// ensure it's a password type
			configItemType := ""

			for _, group := range k.Config.Spec.Groups {
				for _, item := range group.Items {
					if item.Name == name {
						configItemType = item.Type
						goto Found
					}
				}
			}
		Found:

			if configItemType == "" {
				return errors.Errorf("Cannot encrypt item %q because item type was not found", name)
			}
			if configItemType != "password" {
				return errors.Errorf("Cannot encrypt item %q because item type was %q (not password)", name, configItemType)
			}

			encrypted := crypto.Encrypt([]byte(configValue.ValuePlaintext))
			encoded := base64.StdEncoding.EncodeToString(encrypted)

			configValue.Value = encoded
			configValue.ValuePlaintext = ""

			updated[name] = configValue
		}
	}

	k.ConfigValues.Spec.Values = updated

	return nil
}

func (k *KotsKinds) DecryptConfigValues() error {
	if k.ConfigValues == nil {
		return nil
	}

	updated := map[string]kotsv1beta1.ConfigValue{}

	for name, configValue := range k.ConfigValues.Spec.Values {
		// config values doesn't know the type..
		// we could look it up in the config
		// or we can just try to decode and decrypt it

		updated[name] = configValue // will be overwritten if we decrypt anything

		if configValue.Value != "" {
			decoded, err := base64.StdEncoding.DecodeString(configValue.Value)
			if err != nil {
				continue
			}
			decrypted, err := crypto.Decrypt(decoded)
			if err != nil {
				continue
			}

			configValue.Value = ""
			configValue.ValuePlaintext = string(decrypted)

			updated[name] = configValue
		}
	}

	k.ConfigValues.Spec.Values = updated

	return nil
}

func (k *KotsKinds) IsConfigurable() bool {
	if k == nil || k.Config == nil {
		return false
	}
	return len(k.Config.Spec.Groups) > 0
}

func (k *KotsKinds) HasPreflights() bool {
	if k == nil || k.Preflight == nil {
		return false
	}
	return len(k.Preflight.Spec.Analyzers) > 0
}

// GetKustomizeBinaryPath will return the kustomize binary version to use for this application
// applying the default, if there is one, for the current version of kots
func (k KotsKinds) GetKustomizeBinaryPath() string {
	path, err := binaries.GetKustomizePathForVersion(k.KotsApplication.Spec.KustomizeVersion)
	if err != nil {
		logger.Infof("Failed to get kustomize path: %v", err)
		return "kustomize"
	}
	return path
}

func (o KotsKinds) Marshal(g string, v string, k string) (string, error) {
	s := serializer.NewYAMLSerializer(serializer.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	if g == "kots.io" {
		if v == "v1beta1" {
			switch k {
			case "Application":
				var b bytes.Buffer
				if err := s.Encode(&o.KotsApplication, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode kots application")
				}
				return string(b.Bytes()), nil
			case "Installation":
				var b bytes.Buffer
				if err := s.Encode(&o.Installation, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode installation")
				}
				return string(b.Bytes()), nil
			case "License":
				if o.License == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.License, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode license")
				}
				return string(b.Bytes()), nil
			case "Config":
				if o.Config == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.Config, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode config")
				}
				return string(b.Bytes()), nil
			case "ConfigValues":
				if o.ConfigValues == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.ConfigValues, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode configvalues")
				}
				return string(b.Bytes()), nil
			case "Identity":
				if o.Identity == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.Identity, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode identity")
				}
				return string(b.Bytes()), nil
			case "IdentityConfig":
				if o.IdentityConfig == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.IdentityConfig, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode identityconfig")
				}
				return string(b.Bytes()), nil
			}
		}
	}

	if g == "troubleshoot.replicated.com" || g == "troubleshoot.sh" {
		if v == "v1beta1" || v == "v1beta2" {
			switch k {
			case "Collector":
				collector := o.Collector
				// SupportBundle overwrites Collector if defined
				if o.SupportBundle != nil {
					collector = SupportBundleToCollector(o.SupportBundle)
				}
				if collector == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(collector, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode collector")
				}
				return string(b.Bytes()), nil
			case "Analyzer":
				analyzer := o.Analyzer
				// SupportBundle overwrites Analyzer if defined
				if o.SupportBundle != nil {
					analyzer = SupportBundleToAnalyzer(o.SupportBundle)
				}
				if analyzer == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(analyzer, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode analyzer")
				}
				return string(b.Bytes()), nil
			case "Preflight":
				if o.Preflight == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.Preflight, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode preflight")
				}
				return string(b.Bytes()), nil
			case "SupportBundle":
				if o.SupportBundle == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.SupportBundle, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode support bundle")
				}
				return string(b.Bytes()), nil
			case "Redactor":
				if o.Redactor == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.Redactor, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode redactor")
				}
				return string(b.Bytes()), nil
			}
		}
	}

	if g == "app.k8s.io" {
		if v == "v1beta1" {
			if k == "Application" {
				if o.Application == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.Application, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode application")
				}
				return string(b.Bytes()), nil
			}
		}
	}

	if g == "velero.io" {
		if v == "v1" {
			if k == "Backup" {
				if o.Backup == nil {
					return "", nil
				}
				var b bytes.Buffer
				if err := s.Encode(o.Backup, &b); err != nil {
					return "", errors.Wrap(err, "failed to encode backup")
				}
				return string(b.Bytes()), nil
			}
		}
	}

	return "", errors.Errorf("unknown gvk %s/%s, Kind=%s", g, v, k)
}

// create a new kots kinds, ensuring that the require objets exist as empty defaults
func EmptyKotsKinds() KotsKinds {
	kotsKinds := KotsKinds{
		Installation: kotsv1beta1.Installation{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "kots.io/v1beta1",
				Kind:       "Installation",
			},
		},
		KotsApplication: kotsv1beta1.Application{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "kots.io/v1beta1",
				Kind:       "Application",
			},
		},
	}

	return kotsKinds
}

func LoadKotsKindsFromPath(fromDir string) (*KotsKinds, error) {
	kotsKinds := EmptyKotsKinds()

	if fromDir == "" {
		return &kotsKinds, nil
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	err := filepath.Walk(fromDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			contents, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// kots kinds could be part of a multi-yaml doc
			docs := util.ConvertToSingleDocs(contents)

			for _, doc := range docs {
				decoded, gvk, err := decode(doc, nil, nil)
				if err != nil {
					// TODO: log something on yaml errors (based on file extention)
					return nil // not an error because the file might not be yaml
				}

				if strings.HasPrefix(gvk.String(), "troubleshoot.replicated.com/v1beta1,") {
					doc, err = docrewrite.ConvertToV1Beta2(doc)
					if err != nil {
						return errors.Wrap(err, "failed to convert to v1beta2")
					}
					decoded, gvk, err = decode(doc, nil, nil)
					if err != nil {
						return err
					}
				}

				switch gvk.String() {
				case "kots.io/v1beta1, Kind=Config":
					kotsKinds.Config = decoded.(*kotsv1beta1.Config)
				case "kots.io/v1beta1, Kind=ConfigValues":
					kotsKinds.ConfigValues = decoded.(*kotsv1beta1.ConfigValues)
				case "kots.io/v1beta1, Kind=Application":
					kotsKinds.KotsApplication = *decoded.(*kotsv1beta1.Application)
				case "kots.io/v1beta1, Kind=License":
					kotsKinds.License = decoded.(*kotsv1beta1.License)
				case "kots.io/v1beta1, Kind=Identity":
					kotsKinds.Identity = decoded.(*kotsv1beta1.Identity)
				case "kots.io/v1beta1, Kind=IdentityConfig":
					kotsKinds.IdentityConfig = decoded.(*kotsv1beta1.IdentityConfig)
				case "kots.io/v1beta1, Kind=Installation":
					kotsKinds.Installation = *decoded.(*kotsv1beta1.Installation)
				case "kots.io/v1beta1, Kind=HelmChart":
					kotsKinds.HelmCharts = append(kotsKinds.HelmCharts, decoded.(*kotsv1beta1.HelmChart))
				case "troubleshoot.sh/v1beta2, Kind=Collector":
					kotsKinds.Collector = decoded.(*troubleshootv1beta2.Collector)
				case "troubleshoot.sh/v1beta2, Kind=Analyzer":
					kotsKinds.Analyzer = decoded.(*troubleshootv1beta2.Analyzer)
				case "troubleshoot.sh/v1beta2, Kind=SupportBundle":
					kotsKinds.SupportBundle = decoded.(*troubleshootv1beta2.SupportBundle)
				case "troubleshoot.sh/v1beta2, Kind=Redactor":
					kotsKinds.Redactor = decoded.(*troubleshootv1beta2.Redactor)
				case "troubleshoot.sh/v1beta2, Kind=Preflight":
					kotsKinds.Preflight = decoded.(*troubleshootv1beta2.Preflight)
				case "velero.io/v1, Kind=Backup":
					kotsKinds.Backup = decoded.(*velerov1.Backup)
				case "app.k8s.io/v1beta1, Kind=Application":
					kotsKinds.Application = decoded.(*applicationv1beta1.Application)
				}
			}

			return nil
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk upstream dir")
	}

	return &kotsKinds, nil
}

func LoadHelmChartsFromPath(fromDir string) ([]*kotsv1beta1.HelmChart, error) {
	charts := []*kotsv1beta1.HelmChart{}
	decode := scheme.Codecs.UniversalDeserializer().Decode
	err := filepath.Walk(fromDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			contents, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.Wrap(err, "failed to read file")
			}

			decoded, gvk, err := decode(contents, nil, nil)
			if err != nil {
				return nil
			}

			if gvk.String() == "kots.io/v1beta1, Kind=HelmChart" {
				charts = append(charts, decoded.(*kotsv1beta1.HelmChart))
			}

			return nil
		})
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			return nil, errors.Wrap(err, "failed to walk upstream dir")
		}
	}

	return charts, nil
}

func LoadInstallationFromPath(installationFilePath string) (*kotsv1beta1.Installation, error) {
	installationData, err := ioutil.ReadFile(installationFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read installation file")
	}

	return LoadInstallationFromContents(installationData)
}

func LoadSupportBundleFromContents(data []byte) (*troubleshootv1beta2.SupportBundle, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, gvk, err := decode([]byte(data), nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode support bundle data of length %d", len(data))
	}

	if gvk.Group != "troubleshoot.sh" || gvk.Version != "v1beta2" || gvk.Kind != "SupportBundle" {
		return nil, errors.Errorf("unexpected GVK: %s", gvk.String())
	}

	return obj.(*troubleshootv1beta2.SupportBundle), nil
}

func LoadKotsAppFromContents(data []byte) (*kotsv1beta1.Application, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, gvk, err := decode([]byte(data), nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode kots app data of length %d", len(data))
	}

	if gvk.Group != "kots.io" || gvk.Version != "v1beta1" || gvk.Kind != "Application" {
		return nil, errors.Errorf("unexpected GVK: %s", gvk.String())
	}

	return obj.(*kotsv1beta1.Application), nil
}

func LoadInstallationFromContents(installationData []byte) (*kotsv1beta1.Installation, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, gvk, err := decode([]byte(installationData), nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode installation data of length %d", len(installationData))
	}

	if gvk.Group != "kots.io" || gvk.Version != "v1beta1" || gvk.Kind != "Installation" {
		return nil, errors.Errorf("unexpected GVK: %s", gvk.String())
	}

	return obj.(*kotsv1beta1.Installation), nil
}

func LoadLicenseFromPath(licenseFilePath string) (*kotsv1beta1.License, error) {
	licenseData, err := ioutil.ReadFile(licenseFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read license file")
	}

	return LoadLicenseFromBytes(licenseData)
}

func LoadLicenseFromBytes(data []byte) (*kotsv1beta1.License, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, gvk, err := decode([]byte(data), nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode license data")
	}

	if gvk.Group != "kots.io" || gvk.Version != "v1beta1" || gvk.Kind != "License" {
		return nil, errors.Errorf("unexpected GVK: %s", gvk.String())
	}

	return obj.(*kotsv1beta1.License), nil
}

func LoadConfigValuesFromFile(configValuesFilePath string) (*kotsv1beta1.ConfigValues, error) {
	configValuesData, err := ioutil.ReadFile(configValuesFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read configvalues file")
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, gvk, err := decode([]byte(configValuesData), nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode configvalues data")
	}

	if gvk.Group != "kots.io" || gvk.Version != "v1beta1" || gvk.Kind != "ConfigValues" {
		return nil, errors.Errorf("unexpected GVK: %s", gvk.String())
	}

	return obj.(*kotsv1beta1.ConfigValues), nil
}

func LoadConfigFromBytes(data []byte) (*kotsv1beta1.Config, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, gvk, err := decode(data, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode config data")
	}

	if gvk.Group != "kots.io" || gvk.Version != "v1beta1" || gvk.Kind != "Config" {
		return nil, errors.Errorf("unexpected GVK: %s", gvk.String())
	}

	return obj.(*kotsv1beta1.Config), nil
}

func LoadPreflightFromContents(content []byte) (*troubleshootv1beta2.Preflight, error) {
	content, err := docrewrite.ConvertToV1Beta2(content)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert to v1beta2")
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, gvk, err := decode(content, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "not a preflight")
	}

	if gvk.String() == "troubleshoot.sh/v1beta2, Kind=Preflight" {
		return obj.(*troubleshootv1beta2.Preflight), nil
	}

	if gvk.String() == "troubleshoot.replicated.com/v1beta1, Kind=Preflight" {
		return obj.(*troubleshootv1beta2.Preflight), nil
	}

	return nil, errors.Errorf("not a preflight: %s", gvk.String())

}

func LoadBackupFromContents(content []byte) (*velerov1.Backup, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, gvk, err := decode(content, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode content")
	}

	if gvk.String() != "velero.io/v1, Kind=Backup" {
		return nil, errors.Errorf("unexpected gvk: %s", gvk.String())
	}

	return obj.(*velerov1.Backup), nil
}

func SupportBundleToCollector(sb *troubleshootv1beta2.SupportBundle) *troubleshootv1beta2.Collector {
	return &troubleshootv1beta2.Collector{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "troubleshoot.sh/v1beta2",
			Kind:       "Collector",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-collector", sb.Name),
		},
		Spec: troubleshootv1beta2.CollectorSpec{
			Collectors: sb.Spec.Collectors,
		},
	}
}

func SupportBundleToAnalyzer(sb *troubleshootv1beta2.SupportBundle) *troubleshootv1beta2.Analyzer {
	return &troubleshootv1beta2.Analyzer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "troubleshoot.sh/v1beta2",
			Kind:       "Analyzer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-analyzer", sb.Name),
		},
		Spec: troubleshootv1beta2.AnalyzerSpec{
			Analyzers: sb.Spec.Analyzers,
		},
	}
}

type InstallationParams struct {
	KotsadmRegistry        string
	SkipImagePush          bool
	SkipPreflights         bool
	SkipCompatibilityCheck bool
	RegistryIsReadOnly     bool
	EnableImageDeletion    bool
	EnsureRBAC             bool
	SkipRBACCheck          bool
	StrictSecurityContext  bool
	WaitDuration           time.Duration
	WithMinio              bool
}

func GetInstallationParams(configMapName string) (InstallationParams, error) {
	autoConfig := InstallationParams{}

	clientset, err := k8sutil.GetClientset()
	if err != nil {
		return autoConfig, errors.Wrap(err, "failed to get k8s clientset")
	}

	autoConfig.EnableImageDeletion = IsKurl(clientset)

	kotsadmConfigMap, err := clientset.CoreV1().ConfigMaps(util.PodNamespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		if kuberneteserrors.IsNotFound(err) {
			return autoConfig, nil
		}
		return autoConfig, errors.Wrap(err, "failed to get existing kotsadm config map")
	}

	autoConfig.KotsadmRegistry = kotsadmConfigMap.Data["kotsadm-registry"]
	autoConfig.SkipImagePush, _ = strconv.ParseBool(kotsadmConfigMap.Data["initial-app-images-pushed"])
	autoConfig.SkipPreflights, _ = strconv.ParseBool(kotsadmConfigMap.Data["skip-preflights"])
	autoConfig.SkipCompatibilityCheck, _ = strconv.ParseBool(kotsadmConfigMap.Data["skip-compatibility-check"])
	autoConfig.RegistryIsReadOnly, _ = strconv.ParseBool(kotsadmConfigMap.Data["registry-is-read-only"])
	autoConfig.EnsureRBAC, _ = strconv.ParseBool(kotsadmConfigMap.Data["ensure-rbac"])
	autoConfig.SkipRBACCheck, _ = strconv.ParseBool(kotsadmConfigMap.Data["skip-rbac-check"])
	autoConfig.StrictSecurityContext, _ = strconv.ParseBool(kotsadmConfigMap.Data["strict-security-context"])
	autoConfig.WaitDuration, _ = time.ParseDuration(kotsadmConfigMap.Data["wait-duration"])
	autoConfig.WithMinio, _ = strconv.ParseBool(kotsadmConfigMap.Data["with-minio"])

	return autoConfig, nil
}

func LoadIngressConfigFromContents(content []byte) (*kotsv1beta1.IngressConfig, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, gvk, err := decode(content, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode content")
	}

	if gvk.Group == "kots.io" && gvk.Version == "v1beta1" && gvk.Kind == "IngressConfig" {
		return obj.(*kotsv1beta1.IngressConfig), nil
	}

	return nil, errors.Errorf("unexpected gvk: %s", gvk.String())
}

func LoadIdentityFromContents(content []byte) (*kotsv1beta1.Identity, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, gvk, err := decode(content, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode content")
	}

	if gvk.Group == "kots.io" && gvk.Version == "v1beta1" && gvk.Kind == "Identity" {
		return obj.(*kotsv1beta1.Identity), nil
	}

	return nil, errors.Errorf("unexpected gvk: %s", gvk.String())
}

func LoadIdentityConfigFromContents(content []byte) (*kotsv1beta1.IdentityConfig, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, gvk, err := decode(content, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode content")
	}

	if gvk.Group == "kots.io" && gvk.Version == "v1beta1" && gvk.Kind == "IdentityConfig" {
		return obj.(*kotsv1beta1.IdentityConfig), nil
	}

	return nil, errors.Errorf("unexpected gvk: %s", gvk.String())
}

func EncodeIngressConfig(ingressConfig kotsv1beta1.IngressConfig) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	s := serializer.NewYAMLSerializer(serializer.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	err := s.Encode(&ingressConfig, buf)
	return buf.Bytes(), err
}

func EncodeIdentityConfig(spec kotsv1beta1.IdentityConfig) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	s := serializer.NewYAMLSerializer(serializer.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	err := s.Encode(&spec, buf)
	return buf.Bytes(), err
}

func IsKotsVersionCompatibleWithApp(kotsApplication kotsv1beta1.Application, isInstall bool) bool {
	actualSemver, err := semver.ParseTolerant(buildversion.Version())
	if err != nil {
		logger.Error(errors.Wrap(err, "kots build version is invalid"))
		return true
	}

	if kotsApplication.Spec.MinKotsVersion != "" {
		minSemver, err := semver.ParseTolerant(kotsApplication.Spec.MinKotsVersion)
		if err != nil {
			logger.Error(errors.Wrap(err, "minimum kots version specified in the application spec is invalid"))
		} else if actualSemver.LT(minSemver) {
			return false
		}
	}

	if isInstall && kotsApplication.Spec.TargetKotsVersion != "" {
		targetSemver, err := semver.ParseTolerant(kotsApplication.Spec.TargetKotsVersion)
		if err != nil {
			logger.Error(errors.Wrap(err, "target kots version specified in the application spec is invalid"))
		} else if actualSemver.GT(targetSemver) {
			return false
		}
	}

	return true
}

func GetIncompatbileKotsVersionMessage(kotsApplication kotsv1beta1.Application, isInstall bool) string {
	appName := kotsApplication.Spec.Title
	if appName == "" {
		appName = "the app"
	}

	desiredKotsVersion := kotsApplication.Spec.TargetKotsVersion
	if desiredKotsVersion == "" {
		desiredKotsVersion = kotsApplication.Spec.MinKotsVersion
	}

	if isInstall {
		return fmt.Sprintf(
			"This version of %s requires a different version of KOTS from what you currently have installed.\nInstall KOTS version %s and try again.",
			appName,
			desiredKotsVersion,
		)
	}

	return fmt.Sprintf(
		"This version of %s requires a version of KOTS that is different than what you currently have installed. Upgrade KOTS to version %s, and then download this application version again in the admin console or with the 'kots upstream download' command.",
		appName,
		desiredKotsVersion,
	)
}

func IsKotsAutoUpgradeSupported(app *kotsv1beta1.Application) bool {
	if app == nil {
		return false
	}

	for _, f := range app.Spec.ConsoleFeatureFlags {
		if f == "admin-console-auto-updates" {
			return true
		}
	}

	return false
}
